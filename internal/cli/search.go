package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/ghillb/tmgc/internal/tgclient"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search messages",
	}

	cmd.AddCommand(newSearchMessagesCmd())
	return cmd
}

func newSearchMessagesCmd() *cobra.Command {
	var (
		peerRef string
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "messages <query>",
		Short: "Search messages (global or per chat)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}
			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), true, func(ctx context.Context, b *tgclient.Bundle) error {
				query := args[0]
				var res tg.MessagesMessagesClass
				if peerRef == "" {
					res, err = b.Client.API().MessagesSearchGlobal(ctx, &tg.MessagesSearchGlobalRequest{
						Q:          query,
						OffsetPeer: &tg.InputPeerEmpty{},
						Limit:      limit,
					})
				} else {
					peer, err := resolvePeer(ctx, b.Peers, peerRef)
					if err != nil {
						return err
					}
					res, err = b.Client.API().MessagesSearch(ctx, &tg.MessagesSearchRequest{
						Peer:  peer.InputPeer(),
						Q:     query,
						Limit: limit,
					})
				}
				if err != nil {
					return err
				}

				messages, users, chats := extractMessages(res)
				if err := b.Peers.Apply(ctx, users, chats); err != nil {
					return err
				}

				items := buildMessageItems(messages, time.Time{})
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

	cmd.Flags().StringVar(&peerRef, "chat", "", "chat peer (u123, c123, ch123, @username, or phone)")
	cmd.Flags().IntVar(&limit, "limit", 20, "limit number of results")
	return cmd
}
