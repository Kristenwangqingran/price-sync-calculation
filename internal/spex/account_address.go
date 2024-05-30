package spex

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/account_address.pb"
)

const (
	cmdGetDefaultPrivateAddress = "account.address.get_default_private_address"
)

type AccountAddress interface {
	GetDefaultPrivateAddress(ctx context.Context, req *account_address.GetDefaultPrivateAddressRequest) (*account_address.GetDefaultPrivateAddressResponse, error)
}

type accountAddressProxy struct {
}

func GetAccountAddress() AccountAddress {
	return &accountAddressProxy{}
}

func (p *accountAddressProxy) GetDefaultPrivateAddress(ctx context.Context,
	req *account_address.GetDefaultPrivateAddressRequest) (*account_address.GetDefaultPrivateAddressResponse, error) {
	resp := &account_address.GetDefaultPrivateAddressResponse{}

	if err := callSPEX(ctx, cmdGetDefaultPrivateAddress, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
