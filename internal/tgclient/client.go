package tgclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/ghillb/tmgc/internal/config"
	"github.com/ghillb/tmgc/internal/output"
)

type Bundle struct {
	Client     *telegram.Client
	Peers      *peers.Manager
	Dispatcher *tg.UpdateDispatcher
}

type Factory struct {
	Config  config.Config
	Paths   config.Paths
	Printer *output.Printer
	Timeout time.Duration
}

func NewFactory(cfg config.Config, paths config.Paths, printer *output.Printer, timeout time.Duration) *Factory {
	return &Factory{Config: cfg, Paths: paths, Printer: printer, Timeout: timeout}
}

func (f *Factory) Run(ctx context.Context, needsAuth bool, fn func(ctx context.Context, b *Bundle) error) error {
	if f.Config.APIID == 0 || f.Config.APIHash == "" {
		return errors.New("missing API credentials: set TMGC_API_ID and TMGC_API_HASH or run `tmgc auth login --api-id --api-hash`")
	}

	ctx = withTimeout(ctx, f.Timeout)

	dispatcher := tg.NewUpdateDispatcher()
	dispatcherPtr := &dispatcher

	client := telegram.NewClient(f.Config.APIID, f.Config.APIHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: f.Paths.SessionPath},
		UpdateHandler:  dispatcherPtr,
	})

	store, err := NewPeerStore(f.Paths.PeersPath)
	if err != nil {
		return err
	}
	peerManager := peers.Options{Storage: store, Cache: &peers.InmemoryCache{}}.Build(client.API())

	bundle := &Bundle{
		Client:     client,
		Peers:      peerManager,
		Dispatcher: dispatcherPtr,
	}

	return client.Run(ctx, func(ctx context.Context) error {
		if needsAuth {
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return err
			}
			if !status.Authorized {
				return errors.New("not authorized: run `tmgc auth login`")
			}
		}
		return fn(ctx, bundle)
	})
}

func withTimeout(ctx context.Context, timeout time.Duration) context.Context {
	if timeout <= 0 {
		return ctx
	}
	ctx, _ = context.WithTimeout(ctx, timeout)
	return ctx
}

func (f *Factory) Describe() string {
	return fmt.Sprintf("profile=%s", f.Paths.Profile)
}
