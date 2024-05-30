package cidutil

import (
	"context"
	"strings"

	"git.garena.com/shopee/platform/service-governance/viewercontext"
)

const (
	GlobalCID = "global"
)

// FillCtxWithNewCID if original ctx has CID, then will use new CID to replace.
// If has no CID, then fill new CID into the context.
func FillCtxWithNewCID(ctx context.Context, cid string) (context.Context, error) {
	ctx, err := viewercontext.WithCID(
		ctx,
		strings.ToLower(cid),
	)

	return ctx, err
}

func FetchCIDFromCtx(ctx context.Context) string {
	cid, _ := viewercontext.CID(ctx)
	return cid
}
