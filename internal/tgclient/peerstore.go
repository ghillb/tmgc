package tgclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gotd/td/telegram/peers"
)

type peerStoreData struct {
	Peers        map[string]peers.Value `json:"peers"`
	Phones       map[string]peers.Key   `json:"phones"`
	ContactsHash int64                  `json:"contacts_hash"`
}

type PeerStore struct {
	path string
	mu   sync.Mutex
	data peerStoreData
}

func NewPeerStore(path string) (*PeerStore, error) {
	store := &PeerStore{path: path}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *PeerStore) Save(ctx context.Context, key peers.Key, value peers.Value) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensure()
	s.data.Peers[keyString(key)] = value
	return s.persistLocked()
}

func (s *PeerStore) Find(ctx context.Context, key peers.Key) (peers.Value, bool, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensure()
	value, ok := s.data.Peers[keyString(key)]
	return value, ok, nil
}

func (s *PeerStore) SavePhone(ctx context.Context, phone string, key peers.Key) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensure()
	s.data.Phones[phone] = key
	return s.persistLocked()
}

func (s *PeerStore) FindPhone(ctx context.Context, phone string) (peers.Key, peers.Value, bool, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensure()
	key, ok := s.data.Phones[phone]
	if !ok {
		return peers.Key{}, peers.Value{}, false, nil
	}
	value, ok := s.data.Peers[keyString(key)]
	if !ok {
		return peers.Key{}, peers.Value{}, false, nil
	}
	return key, value, true, nil
}

func (s *PeerStore) GetContactsHash(ctx context.Context) (int64, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensure()
	return s.data.ContactsHash, nil
}

func (s *PeerStore) SaveContactsHash(ctx context.Context, hash int64) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensure()
	s.data.ContactsHash = hash
	return s.persistLocked()
}

func (s *PeerStore) load() error {
	s.ensure()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read peers store: %w", err)
	}
	if err := json.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("parse peers store: %w", err)
	}
	s.ensure()
	return nil
}

func (s *PeerStore) ensure() {
	if s.data.Peers == nil {
		s.data.Peers = make(map[string]peers.Value)
	}
	if s.data.Phones == nil {
		s.data.Phones = make(map[string]peers.Key)
	}
}

func (s *PeerStore) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create peers dir: %w", err)
	}
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode peers store: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write peers store: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename peers store: %w", err)
	}
	return nil
}

func keyString(key peers.Key) string {
	return fmt.Sprintf("%s%d", key.Prefix, key.ID)
}
