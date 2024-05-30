package cb_sip_logic

import (
	"context"
	"fmt"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
)

func (c *CbSipLogicImpl) GetCbSipRateConfig(ctx context.Context, infoType uint32) (model.CbSipRateConfigResult, error) {
	switch infoType {
	case uint32(pb.Constant_DEFAULT_SIP_RATE):
		res, err := c.getDefaultCbSipRateConfig(ctx)
		if err != nil {
			return model.CbSipRateConfigResult{}, err
		}
		return model.CbSipRateConfigResult{
			DefaultSipRateStr: res,
		}, nil
	case uint32(pb.Constant_LIMIT_LIST):
		res, err := c.getSipRateLimit(ctx)
		if err != nil {
			return model.CbSipRateConfigResult{}, err
		}
		return model.CbSipRateConfigResult{
			SipRateLimitStr: res,
		}, nil
	default:
		return model.CbSipRateConfigResult{}, cerr.New(fmt.Sprintf("invalid cb sip rate config info type: %v", infoType), uint32(pb.Constant_ERROR_PARAMS))
	}
}

func (c *CbSipLogicImpl) getDefaultCbSipRateConfig(ctx context.Context) (string, error) {
	session := c.sipRepo.DbSession()
	res, err := c.sipRepo.GetSystemConfigRecordByType(ctx, session, sip_db.SystemConfigTypeDefaultSipRate)
	if err != nil {
		return "", err
	}
	return res.ConfigData, nil
}

func (c *CbSipLogicImpl) getSipRateLimit(ctx context.Context) (string, error) {
	session := c.sipRepo.DbSession()
	res, err := c.sipRepo.GetSystemConfigRecordByType(ctx, session, sip_db.SystemConfigTypeCnscSipRateLimit)
	if err != nil {
		return "", err
	}
	return res.ConfigData, nil
}
