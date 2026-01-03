package cli

import (
	"context"

	"github.com/gotd/td/telegram/peers"
)

func resolvePeer(ctx context.Context, pm *peers.Manager, input string) (peers.Peer, error) {
	if id, ok := parsePeerRef(input); ok {
		return pm.ResolveTDLibID(ctx, id)
	}
	return pm.Resolve(ctx, input)
}
