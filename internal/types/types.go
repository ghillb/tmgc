package types

import "time"

type AuthStatus struct {
	Authorized bool   `json:"authorized"`
	UserID     int64  `json:"user_id,omitempty"`
	Username   string `json:"username,omitempty"`
	Phone      string `json:"phone,omitempty"`
	IsBot      bool   `json:"is_bot"`
}

type ChatListItem struct {
	PeerID        int64  `json:"peer_id"`
	PeerRef       string `json:"peer_ref"`
	PeerType      string `json:"peer_type"`
	Title         string `json:"title"`
	Username      string `json:"username,omitempty"`
	UnreadCount   int    `json:"unread_count"`
	LastMessageID int    `json:"last_message_id,omitempty"`
	Pinned        bool   `json:"pinned"`
}

type MessageItem struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	Text       string    `json:"text,omitempty"`
	FromPeerID int64     `json:"from_peer_id,omitempty"`
	PeerID     int64     `json:"peer_id,omitempty"`
	Out        bool      `json:"out"`
	Service    bool      `json:"service"`
}

type SendResult struct {
	OK        bool   `json:"ok"`
	MessageID int    `json:"message_id,omitempty"`
	Updates   string `json:"updates_type,omitempty"`
}
