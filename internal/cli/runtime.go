package cli

import (
	"context"
	"errors"
	"time"

	"github.com/ghillb/tmgc/internal/config"
	"github.com/ghillb/tmgc/internal/output"
)

type runtimeKey struct{}

type Runtime struct {
	Paths   config.Paths
	Config  *config.Config
	Printer *output.Printer
	Timeout time.Duration
}

func withRuntime(ctx context.Context, rt *Runtime) context.Context {
	return context.WithValue(ctx, runtimeKey{}, rt)
}

func runtimeFrom(ctx context.Context) (*Runtime, error) {
	rt, ok := ctx.Value(runtimeKey{}).(*Runtime)
	if !ok || rt == nil {
		return nil, errors.New("runtime not initialized")
	}
	return rt, nil
}
