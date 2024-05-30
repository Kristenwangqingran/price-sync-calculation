package logic

import (
	"context"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

type CurrencyConvertLogic interface {
	ConvertCurrency(ctx context.Context, req model.ConvertCurrencyRequest) (model.ConvertCurrencyResult, error)
}
