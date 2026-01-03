package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/ghillb/tmgc/internal/tgclient"
	"github.com/ghillb/tmgc/internal/types"
)

func newChatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Chat operations",
	}

	cmd.AddCommand(newChatListCmd())
	cmd.AddCommand(newChatHistoryCmd())

	return cmd
}

func newChatListCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}
			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), true, func(ctx context.Context, b *tgclient.Bundle) error {
				api := b.Client.API()
				res, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
					Limit:      limit,
					OffsetPeer: &tg.InputPeerEmpty{},
				})
				if err != nil {
					return err
				}

				dialogs, users, chats := extractDialogs(res)
				if err := b.Peers.Apply(ctx, users, chats); err != nil {
					return err
				}

				userMap, chatMap, channelMap := buildPeerMaps(users, chats)
				items := make([]types.ChatListItem, 0, len(dialogs))
				for _, d := range dialogs {
					dialog, ok := d.(*tg.Dialog)
					if !ok {
						continue
					}
					peer := peerFromDialog(b.Peers, dialog.Peer, userMap, chatMap, channelMap)
					if peer == nil {
						continue
					}

					id := peer.TDLibPeerID()
					item := types.ChatListItem{
						PeerID:        int64(id),
						PeerRef:       peerRefFromID(id),
						PeerType:      peerTypeFromID(id),
						Title:         peer.VisibleName(),
						UnreadCount:   dialog.UnreadCount,
						LastMessageID: dialog.TopMessage,
						Pinned:        dialog.Pinned,
					}
					if username, ok := peer.Username(); ok {
						item.Username = username
					}
					items = append(items, item)
				}

				switch rt.Printer.Mode {
				case "json":
					return rt.Printer.JSON(items)
				case "plain":
					lines := make([]string, 0, len(items))
					for _, item := range items {
						line := fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%d\t%t",
							item.PeerRef, item.PeerType, item.Title, item.Username, item.UnreadCount, item.LastMessageID, item.Pinned,
						)
						lines = append(lines, line)
					}
					rt.Printer.Plain(lines)
				default:
					rows := [][]string{{"PEER", "TYPE", "TITLE", "USERNAME", "UNREAD", "TOP", "PINNED"}}
					for _, item := range items {
						rows = append(rows, []string{
							item.PeerRef,
							item.PeerType,
							item.Title,
							item.Username,
							strconv.Itoa(item.UnreadCount),
							strconv.Itoa(item.LastMessageID),
							strconv.FormatBool(item.Pinned),
						})
					}
					rt.Printer.Table(rows)
				}
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "limit number of chats")
	return cmd
}

func newChatHistoryCmd() *cobra.Command {
	var (
		limit int
		since string
	)

	cmd := &cobra.Command{
		Use:   "history <peer>",
		Short: "Read chat history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}
			cutoff, err := parseSince(since)
			if err != nil {
				return err
			}

			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), true, func(ctx context.Context, b *tgclient.Bundle) error {
				peer, err := resolvePeer(ctx, b.Peers, args[0])
				if err != nil {
					return err
				}

				api := b.Client.API()
				res, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
					Peer:  peer.InputPeer(),
					Limit: limit,
				})
				if err != nil {
					return err
				}

				messages, users, chats := extractMessages(res)
				if err := b.Peers.Apply(ctx, users, chats); err != nil {
					return err
				}

				items := buildMessageItems(messages, cutoff)
				switch rt.Printer.Mode {
				case "json":
					return rt.Printer.JSON(items)
				case "plain":
					lines := make([]string, 0, len(items))
					for _, item := range items {
						line := fmt.Sprintf("%d\t%s\t%d\t%s",
							item.ID,
							item.Date.Format(time.RFC3339),
							item.FromPeerID,
							item.Text,
						)
						lines = append(lines, line)
					}
					rt.Printer.Plain(lines)
				default:
					rows := [][]string{{"ID", "DATE", "FROM", "TEXT"}}
					for _, item := range items {
						rows = append(rows, []string{
							strconv.Itoa(item.ID),
							item.Date.Format(time.RFC3339),
							strconv.FormatInt(item.FromPeerID, 10),
							item.Text,
						})
					}
					rt.Printer.Table(rows)
				}
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "limit number of messages")
	cmd.Flags().StringVar(&since, "since", "", "filter messages after RFC3339 timestamp")
	return cmd
}

func extractDialogs(res tg.MessagesDialogsClass) ([]tg.DialogClass, []tg.UserClass, []tg.ChatClass) {
	switch v := res.(type) {
	case *tg.MessagesDialogs:
		return v.Dialogs, v.Users, v.Chats
	case *tg.MessagesDialogsSlice:
		return v.Dialogs, v.Users, v.Chats
	default:
		return nil, nil, nil
	}
}

func extractMessages(res tg.MessagesMessagesClass) ([]tg.MessageClass, []tg.UserClass, []tg.ChatClass) {
	switch v := res.(type) {
	case *tg.MessagesMessages:
		return v.Messages, v.Users, v.Chats
	case *tg.MessagesMessagesSlice:
		return v.Messages, v.Users, v.Chats
	case *tg.MessagesChannelMessages:
		return v.Messages, v.Users, v.Chats
	default:
		return nil, nil, nil
	}
}

func buildPeerMaps(users []tg.UserClass, chats []tg.ChatClass) (map[int64]*tg.User, map[int64]*tg.Chat, map[int64]*tg.Channel) {
	userMap := make(map[int64]*tg.User)
	chatMap := make(map[int64]*tg.Chat)
	channelMap := make(map[int64]*tg.Channel)

	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			userMap[user.ID] = user
		}
	}
	for _, c := range chats {
		switch v := c.(type) {
		case *tg.Chat:
			chatMap[v.ID] = v
		case *tg.Channel:
			channelMap[v.ID] = v
		}
	}
	return userMap, chatMap, channelMap
}

func peerFromDialog(m *peers.Manager, p tg.PeerClass, users map[int64]*tg.User, chats map[int64]*tg.Chat, channels map[int64]*tg.Channel) peers.Peer {
	switch v := p.(type) {
	case *tg.PeerUser:
		if u, ok := users[v.UserID]; ok {
			return m.User(u)
		}
	case *tg.PeerChat:
		if c, ok := chats[v.ChatID]; ok {
			return m.Chat(c)
		}
	case *tg.PeerChannel:
		if c, ok := channels[v.ChannelID]; ok {
			return m.Channel(c)
		}
	}
	return nil
}

func buildMessageItems(messages []tg.MessageClass, cutoff time.Time) []types.MessageItem {
	items := make([]types.MessageItem, 0, len(messages))
	for _, msg := range messages {
		switch m := msg.(type) {
		case *tg.Message:
			item := types.MessageItem{
				ID:   m.ID,
				Date: time.Unix(int64(m.Date), 0),
				Text: m.Message,
				Out:  m.Out,
			}
			if m.FromID != nil {
				if id, ok := peerIDFromPeerClass(m.FromID); ok {
					item.FromPeerID = int64(id)
				}
			}
			if m.PeerID != nil {
				if id, ok := peerIDFromPeerClass(m.PeerID); ok {
					item.PeerID = int64(id)
				}
			}
			if cutoff.IsZero() || item.Date.After(cutoff) {
				if item.Text == "" && m.Media != nil {
					item.Text = "<non-text>"
				}
				items = append(items, item)
			}
		case *tg.MessageService:
			item := types.MessageItem{
				ID:      m.ID,
				Date:    time.Unix(int64(m.Date), 0),
				Text:    "<service>",
				Out:     m.Out,
				Service: true,
			}
			if cutoff.IsZero() || item.Date.After(cutoff) {
				items = append(items, item)
			}
		}
	}
	return items
}
