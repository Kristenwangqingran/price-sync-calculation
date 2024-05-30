package region_rate_table_config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/wire"

	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

const (
	typeRegionRateTableConfig = 20
)

var ProviderSet = wire.NewSet(
	NewRegionRateTableConfigRepoImpl,
	wire.Bind(new(RegionRateTableConfigRepo), new(*RegionRateTableConfigRepoImpl)),
)

type RegionRateTableConfigRepo interface {
	GetRegionRateTableConfig() map[string]map[string]*model.RateTableCfg
}

type RegionRateTableConfigRepoImpl struct {
	localCache map[string]map[string]*model.RateTableCfg

	sipRepo sip_db.SipRepo
}

func NewRegionRateTableConfigRepoImpl(sipRepo sip_db.SipRepo) *RegionRateTableConfigRepoImpl {
	v := &RegionRateTableConfigRepoImpl{
		sipRepo: sipRepo,

		localCache: map[string]map[string]*model.RateTableCfg{},
	}

	v.init()
	return v
}

func (r *RegionRateTableConfigRepoImpl) init() {
	ticker := time.NewTicker(10 * time.Second)
	ch := make(chan struct{})
	go func() {
		var err error
		ctx := context.TODO()
		firstLoad := true
		for {
			err = r.LoadAllRegionTableConfig(ctx)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			if firstLoad {
				firstLoad = false
				logging.GetLogger(ctx).Debug(fmt.Sprintf("region rate table cfg =%+v", r.localCache))
				ch <- struct{}{}
			}
			<-ticker.C
		}
	}()
	<-ch
}

func (r *RegionRateTableConfigRepoImpl) LoadAllRegionTableConfig(ctx context.Context) error {
	session := r.sipRepo.DbSession()
	syscfg, err := r.sipRepo.GetSystemConfigRecordByType(ctx, session, typeRegionRateTableConfig)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("load region rate table cfg failed, err=%v", err))
		return err
	}
	if syscfg == nil {
		logging.GetLogger(ctx).Error("region rate table cfg not set")
		return cerr.New("region rate table cfg not set", uint32(pb.Constant_ERROR_NOT_FOUND))
	}
	tmpCfg := map[string]map[string]*model.RateTableCfg{}
	err = json.Unmarshal([]byte(syscfg.ConfigData), &tmpCfg)
	if err != nil {
		logging.GetLogger(ctx).Error(fmt.Sprintf("unmarshal region rate table cfg failed, err=%v", err))
		return err
	}
	r.localCache = tmpCfg
	return nil
}

func (r *RegionRateTableConfigRepoImpl) GetRegionRateTableConfig() map[string]map[string]*model.RateTableCfg {
	return r.localCache
}
