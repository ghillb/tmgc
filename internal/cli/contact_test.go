package cli

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestContactMatchesQuery(t *testing.T) {
	user := &tg.User{
		ID:        101,
		FirstName: "David",
		LastName:  "Miller",
		Username:  "DaViD_One",
	}

	tests := []struct {
		name   string
		query  string
		expect bool
	}{
		{name: "match first name", query: "vid", expect: true},
		{name: "match last name", query: "mill", expect: true},
		{name: "match username", query: "david_", expect: true},
		{name: "no match", query: "halle", expect: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contactMatchesQuery(user, tt.query)
			if got != tt.expect {
				t.Fatalf("contactMatchesQuery() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestContactDisplayName(t *testing.T) {
	tests := []struct {
		name string
		user *tg.User
		want string
	}{
		{
			name: "first and last",
			user: &tg.User{ID: 1, FirstName: "David", LastName: "Halle", Username: "dhalle"},
			want: "David Halle",
		},
		{
			name: "username fallback",
			user: &tg.User{ID: 2, Username: "alice"},
			want: "@alice",
		},
		{
			name: "id fallback",
			user: &tg.User{ID: 3},
			want: "u3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contactDisplayName(tt.user)
			if got != tt.want {
				t.Fatalf("contactDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractContactUsers(t *testing.T) {
	res := &tg.ContactsContacts{
		Contacts: []tg.Contact{
			{UserID: 10},
			{UserID: 30},
		},
		Users: []tg.UserClass{
			&tg.User{ID: 10, FirstName: "A"},
			&tg.User{ID: 20, FirstName: "B"},
			&tg.User{ID: 30, FirstName: "C"},
			&tg.UserEmpty{ID: 40},
		},
	}

	users := extractContactUsers(res)
	if len(users) != 2 {
		t.Fatalf("extractContactUsers() length = %d, want 2", len(users))
	}
	if users[0].ID != 10 || users[1].ID != 30 {
		t.Fatalf("extractContactUsers() ids = [%d, %d], want [10, 30]", users[0].ID, users[1].ID)
	}
}

func TestFormatContactUsername(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "alice", want: "@alice"},
		{in: "@bob", want: "@bob"},
		{in: "  carol  ", want: "@carol"},
		{in: "", want: ""},
	}

	for _, tt := range tests {
		if got := formatContactUsername(tt.in); got != tt.want {
			t.Fatalf("formatContactUsername(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
