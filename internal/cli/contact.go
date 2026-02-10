package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/spf13/cobra"

	"github.com/ghillb/tmgc/internal/tgclient"
	"github.com/ghillb/tmgc/internal/types"
)

func newContactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contact",
		Short: "Contact operations",
	}

	cmd.AddCommand(newContactSearchCmd())
	return cmd
}

func newContactSearchCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search contacts by display name or username",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}

			query := strings.TrimSpace(args[0])
			if query == "" {
				return errors.New("query cannot be empty")
			}
			if limit < 0 {
				return errors.New("--limit must be >= 0")
			}

			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), true, func(ctx context.Context, b *tgclient.Bundle) error {
				res, err := b.Client.API().ContactsGetContacts(ctx, 0)
				if err != nil {
					return err
				}

				needle := strings.ToLower(query)
				users := extractContactUsers(res)
				items := make([]types.ContactSearchItem, 0, len(users))
				for _, user := range users {
					if !contactMatchesQuery(user, needle) {
						continue
					}

					items = append(items, types.ContactSearchItem{
						DisplayName: contactDisplayName(user),
						Username:    formatContactUsername(user.Username),
						User:        userPeerRef(user.ID),
					})

					if limit > 0 && len(items) >= limit {
						break
					}
				}

				switch rt.Printer.Mode {
				case "json":
					return rt.Printer.JSON(items)
				case "plain":
					lines := make([]string, 0, len(items))
					for _, item := range items {
						line := fmt.Sprintf("%s\t%s\t%s", item.DisplayName, item.Username, item.User)
						lines = append(lines, line)
					}
					rt.Printer.Plain(lines)
				default:
					rows := [][]string{{"DISPLAY_NAME", "USERNAME", "USER"}}
					for _, item := range items {
						rows = append(rows, []string{item.DisplayName, item.Username, item.User})
					}
					rt.Printer.Table(rows)
				}
				return nil
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "limit number of results")
	return cmd
}

func extractContactUsers(res tg.ContactsContactsClass) []*tg.User {
	contacts, ok := res.(*tg.ContactsContacts)
	if !ok {
		return nil
	}

	contactIDs := make(map[int64]struct{}, len(contacts.Contacts))
	for _, contact := range contacts.Contacts {
		contactIDs[contact.UserID] = struct{}{}
	}

	users := make([]*tg.User, 0, len(contacts.Users))
	for _, u := range contacts.Users {
		user, ok := u.(*tg.User)
		if !ok {
			continue
		}
		if len(contactIDs) > 0 {
			if _, ok := contactIDs[user.ID]; !ok {
				continue
			}
		}
		users = append(users, user)
	}
	return users
}

func contactMatchesQuery(user *tg.User, needleLower string) bool {
	return strings.Contains(strings.ToLower(user.FirstName), needleLower) ||
		strings.Contains(strings.ToLower(user.LastName), needleLower) ||
		strings.Contains(strings.ToLower(user.Username), needleLower)
}

func contactDisplayName(user *tg.User) string {
	name := strings.TrimSpace(strings.TrimSpace(user.FirstName + " " + user.LastName))
	if name != "" {
		return name
	}
	if user.Username != "" {
		return formatContactUsername(user.Username)
	}
	return userPeerRef(user.ID)
}

func formatContactUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	return "@" + strings.TrimPrefix(username, "@")
}

func userPeerRef(userID int64) string {
	var id constant.TDLibPeerID
	id.User(userID)
	return peerRefFromID(id)
}
