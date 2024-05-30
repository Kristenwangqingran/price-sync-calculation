package spex

import (
	"context"

	ac "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/account_core.pb"
)

const (
	cmdGetAccountBatch = "account.core.get_account_batch"
)

type AccountCore interface {
	GetAccountBatch(ctx context.Context, req *ac.GetAccountBatchRequest) (*ac.GetAccountBatchResponse, error)
}

type accountCoreProxy struct {
}

func NewAccountCore() AccountCore {
	return &accountCoreProxy{}
}

func (p *accountCoreProxy) GetAccountBatch(ctx context.Context, req *ac.GetAccountBatchRequest) (*ac.GetAccountBatchResponse, error) {
	resp := &ac.GetAccountBatchResponse{}
	err := callSPEX(ctx, cmdGetAccountBatch, req, resp)
	return resp, err
}
