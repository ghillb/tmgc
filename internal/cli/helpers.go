package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

func parsePeerRef(input string) (constant.TDLibPeerID, bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return 0, false
	}

	var (
		prefix string
		numStr string
	)

	switch {
	case strings.HasPrefix(s, "ch"):
		prefix = "ch"
		numStr = strings.TrimPrefix(s, "ch")
	case strings.HasPrefix(s, "u"):
		prefix = "u"
		numStr = strings.TrimPrefix(s, "u")
	case strings.HasPrefix(s, "c"):
		prefix = "c"
		numStr = strings.TrimPrefix(s, "c")
	default:
		return 0, false
	}

	id, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, false
	}

	var td constant.TDLibPeerID
	switch prefix {
	case "u":
		td.User(id)
	case "c":
		td.Chat(id)
	case "ch":
		td.Channel(id)
	default:
		return 0, false
	}
	return td, true
}

func peerRefFromID(id constant.TDLibPeerID) string {
	switch {
	case id.IsUser():
		return fmt.Sprintf("u%d", id.ToPlain())
	case id.IsChat():
		return fmt.Sprintf("c%d", id.ToPlain())
	case id.IsChannel():
		return fmt.Sprintf("ch%d", id.ToPlain())
	default:
		return fmt.Sprintf("p%d", int64(id))
	}
}

func peerTypeFromID(id constant.TDLibPeerID) string {
	switch {
	case id.IsUser():
		return "user"
	case id.IsChat():
		return "chat"
	case id.IsChannel():
		return "channel"
	default:
		return "unknown"
	}
}

func peerIDFromPeerClass(p tg.PeerClass) (constant.TDLibPeerID, bool) {
	var id constant.TDLibPeerID
	switch v := p.(type) {
	case *tg.PeerUser:
		id.User(v.UserID)
		return id, true
	case *tg.PeerChat:
		id.Chat(v.ChatID)
		return id, true
	case *tg.PeerChannel:
		id.Channel(v.ChannelID)
		return id, true
	default:
		return 0, false
	}
}

func parseSince(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}
