package cli

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

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
	)

	cmd := &cobra.Command{
		Use:   "send <peer> <text>",
		Short: "Send a text message",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}

			peerArg := args[0]
			text := strings.Join(args[1:], " ")
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("message text cannot be empty")
			}

			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), true, func(ctx context.Context, b *tgclient.Bundle) error {
				peer, err := resolvePeer(ctx, b.Peers, peerArg)
				if err != nil {
					return err
				}

				req := &tg.MessagesSendMessageRequest{
					Peer:     peer.InputPeer(),
					Message:  text,
					RandomID: rand.Int63(),
					Silent:   silent,
				}
				if replyID != 0 {
					req.ReplyTo = &tg.InputReplyToMessage{ReplyToMsgID: replyID}
				}

				updates, err := b.Client.API().MessagesSendMessage(ctx, req)
				if err != nil {
					return err
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
	return cmd
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
