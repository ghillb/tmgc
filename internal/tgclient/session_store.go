package tgclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/keyring"
	"github.com/gotd/td/session"

	"github.com/ghillb/tmgc/internal/config"
	"github.com/ghillb/tmgc/internal/output"
)

const (
	sessionStoreKeyring = "keyring"
	sessionStoreFile    = "file"
)

func NewSessionStorage(cfg config.Config, paths config.Paths, printer *output.Printer) session.Storage {
	store := normalizeStore(cfg.SessionStore)
	if store == "" {
		store = sessionStoreKeyring
	}

	if store == sessionStoreFile {
		warnUnencryptedSession(printer, paths.SessionPath)
		return &session.FileStorage{Path: paths.SessionPath}
	}

	kr, err := openKeyring()
	if err != nil {
		warnKeyringFallback(printer, paths.SessionPath, err)
		return &session.FileStorage{Path: paths.SessionPath}
	}

	if err := checkKeyring(kr, paths.Profile); err != nil {
		warnKeyringFallback(printer, paths.SessionPath, err)
		return &session.FileStorage{Path: paths.SessionPath}
	}

	return &keyringStorage{kr: kr, key: sessionKey(paths.Profile)}
}

func ClearSession(cfg config.Config, paths config.Paths, printer *output.Printer) {
	_ = os.Remove(paths.SessionPath)
	_ = os.Remove(paths.PeersPath)
	kr, err := openKeyring()
	if err != nil {
		return
	}
	_ = kr.Remove(sessionKey(paths.Profile))
}

type keyringStorage struct {
	kr  keyring.Keyring
	key string
}

func (k *keyringStorage) LoadSession(_ context.Context) ([]byte, error) {
	item, err := k.kr.Get(k.key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, session.ErrNotFound
		}
		return nil, err
	}
	return item.Data, nil
}

func (k *keyringStorage) StoreSession(_ context.Context, data []byte) error {
	return k.kr.Set(keyring.Item{Key: k.key, Data: data})
}

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: "tmgc",
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.PassBackend,
			keyring.KeyCtlBackend,
		},
	})
}

func normalizeStore(store string) string {
	store = strings.TrimSpace(strings.ToLower(store))
	if store == sessionStoreKeyring || store == sessionStoreFile {
		return store
	}
	return ""
}

func warnKeyringFallback(printer *output.Printer, path string, err error) {
	if printer == nil {
		return
	}
	fmt.Fprintf(printer.Err, "Keyring unavailable (%v). Falling back to unencrypted session file at %s.\n", err, path)
}

func warnUnencryptedSession(printer *output.Printer, path string) {
	if printer == nil {
		return
	}
	fmt.Fprintf(printer.Err, "Warning: using unencrypted session file at %s. Anyone with access to this file can reuse your session.\n", path)
}

func sessionKey(profile string) string {
	if profile == "" {
		profile = "default"
	}
	return "session/" + profile
}

func checkKeyring(kr keyring.Keyring, profile string) error {
	key := "tmgc/health/" + profile
	item := keyring.Item{Key: key, Data: []byte("ok")}
	if err := kr.Set(item); err != nil {
		return err
	}
	got, err := kr.Get(key)
	if err != nil {
		return err
	}
	if !bytes.Equal(got.Data, item.Data) {
		return errors.New("keyring data mismatch")
	}
	_ = kr.Remove(key)
	return nil
}
