package cli

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/ghillb/tmgc/internal/tgclient"
	"github.com/ghillb/tmgc/internal/types"
)

func newMessageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message",
		Short: "Message operations",
	}

	cmd.AddCommand(newMessageSendCmd())

	return cmd
}

func newMessageSendCmd() *cobra.Command {
	var (
		replyID int
		silent  bool
		file    string
		caption string
	)

	cmd := &cobra.Command{
		Use:   "send <peer> [text]",
		Short: "Send a text message or file",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("peer is required")
			}
			if file == "" && len(args) < 2 {
				return fmt.Errorf("message text cannot be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}

			peerArg := args[0]
			textArgs := args[1:]
			if file != "" && caption != "" && len(textArgs) > 0 {
				return fmt.Errorf("use --caption or trailing text, not both")
			}
			text := strings.Join(textArgs, " ")
			if file == "" && strings.TrimSpace(text) == "" {
				return fmt.Errorf("message text cannot be empty")
			}
			if file != "" && caption != "" {
				text = caption
			}

			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), true, func(ctx context.Context, b *tgclient.Bundle) error {
				peer, err := resolvePeer(ctx, b.Peers, peerArg)
				if err != nil {
					return err
				}

				var updates tg.UpdatesClass
				if file == "" {
					req := &tg.MessagesSendMessageRequest{
						Peer:     peer.InputPeer(),
						Message:  text,
						RandomID: rand.Int63(),
						Silent:   silent,
					}
					if replyID != 0 {
						req.ReplyTo = &tg.InputReplyToMessage{ReplyToMsgID: replyID}
					}

					updates, err = b.Client.API().MessagesSendMessage(ctx, req)
					if err != nil {
						return err
					}
				} else {
					media, err := uploadMedia(ctx, b.Client.API(), file)
					if err != nil {
						return err
					}

					req := &tg.MessagesSendMediaRequest{
						Peer:     peer.InputPeer(),
						Media:    media,
						Message:  text,
						RandomID: rand.Int63(),
						Silent:   silent,
					}
					if replyID != 0 {
						req.ReplyTo = &tg.InputReplyToMessage{ReplyToMsgID: replyID}
					}
					updates, err = b.Client.API().MessagesSendMedia(ctx, req)
					if err != nil {
						return err
					}
				}

				result := types.SendResult{OK: true}
				if id, ok := extractSentMessageID(updates); ok {
					result.MessageID = id
				}
				result.Updates = fmt.Sprintf("%T", updates)

				switch rt.Printer.Mode {
				case "json":
					return rt.Printer.JSON(result)
				case "plain":
					line := fmt.Sprintf("%t\t%d", result.OK, result.MessageID)
					rt.Printer.Plain([]string{line})
				default:
					rt.Printer.Table([][]string{{"OK", "MESSAGE_ID"}, {
						fmt.Sprintf("%t", result.OK),
						fmt.Sprintf("%d", result.MessageID),
					}})
				}
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&replyID, "reply", 0, "reply to message id")
	cmd.Flags().BoolVar(&silent, "silent", false, "send silently")
	cmd.Flags().StringVar(&file, "file", "", "path to file to upload")
	cmd.Flags().StringVar(&caption, "caption", "", "caption for uploaded media")
	return cmd
}

func uploadMedia(ctx context.Context, api *tg.Client, path string) (tg.InputMediaClass, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", path)
	}

	mimeType, isPhoto, err := detectMedia(file, info.Name())
	if err != nil {
		return nil, err
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	upload := uploader.NewUpload(info.Name(), file, info.Size())
	up := uploader.NewUploader(api)
	inputFile, err := up.Upload(ctx, upload)
	if err != nil {
		return nil, err
	}

	if isPhoto {
		return &tg.InputMediaUploadedPhoto{File: inputFile}, nil
	}

	attrs := []tg.DocumentAttributeClass{
		&tg.DocumentAttributeFilename{FileName: info.Name()},
	}
	media := &tg.InputMediaUploadedDocument{
		File:       inputFile,
		MimeType:   mimeType,
		Attributes: attrs,
		ForceFile:  true,
	}
	return media, nil
}

func detectMedia(file *os.File, name string) (string, bool, error) {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", false, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", false, err
	}

	ext := strings.ToLower(filepath.Ext(name))
	mimeType := http.DetectContentType(buf[:n])
	if mimeType == "application/octet-stream" && ext != "" {
		if m := mime.TypeByExtension(ext); m != "" {
			mimeType = m
		}
	}
	if strings.Contains(mimeType, ";") {
		parts := strings.Split(mimeType, ";")
		mimeType = strings.TrimSpace(parts[0])
	}

	isPhoto := false
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		isPhoto = true
	default:
		if strings.HasPrefix(mimeType, "image/") && !strings.Contains(mimeType, "gif") {
			isPhoto = true
		}
	}

	return mimeType, isPhoto, nil
}

func extractSentMessageID(updates tg.UpdatesClass) (int, bool) {
	switch u := updates.(type) {
	case *tg.UpdateShortSentMessage:
		return u.ID, true
	case *tg.Updates:
		return extractFromUpdatesList(u.Updates)
	case *tg.UpdatesCombined:
		return extractFromUpdatesList(u.Updates)
	default:
		return 0, false
	}
}

func extractFromUpdatesList(list []tg.UpdateClass) (int, bool) {
	for _, upd := range list {
		switch u := upd.(type) {
		case *tg.UpdateNewMessage:
			if msg, ok := u.Message.(*tg.Message); ok {
				return msg.ID, true
			}
		case *tg.UpdateNewChannelMessage:
			if msg, ok := u.Message.(*tg.Message); ok {
				return msg.ID, true
			}
		}
	}
	return 0, false
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
