package hpfn_config

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type HpfnConfigRepo interface {
	GetOne(ctx context.Context, hpfnKey string, weight int) (*sip_db.HpfnConfig, error)
	GetAllHpfnConfig(ctx context.Context) (map[string][]*sip_db.HpfnConfig, error)
	GetAllHpfnConfigFromDb(ctx context.Context) ([]*sip_db.HpfnConfig, error)
}

type HpfnConfigRepoImpl struct {
	mu         *sync.RWMutex
	localCache map[string][]*sip_db.HpfnConfig

	sipRepo sip_db.SipRepo
}

func NewHpfnConfigRepoImpl(sipRepo sip_db.SipRepo) *HpfnConfigRepoImpl {
	v := &HpfnConfigRepoImpl{
		mu:         &sync.RWMutex{},
		localCache: map[string][]*sip_db.HpfnConfig{},
		sipRepo:    sipRepo,
	}
	v.init()
	return v
}

func (h *HpfnConfigRepoImpl) init() {
	tick := time.NewTicker(10 * time.Second)
	ch := make(chan struct{})

	go func() {
		ctx := context.TODO()
		firstLoad := true
		for {
			data, err := h.GetAllHpfnConfigFromDb(ctx)
			if err != nil {
				logging.GetLogger(ctx).Error("list all hpfn_config_tab", ulog.Error(err))
				time.Sleep(1 * time.Second)
				continue
			}
			h.UpdateHpfnConfigInLocalCache(ctx, data)
			if firstLoad {
				firstLoad = false
				s, _ := json.Marshal(&h.localCache)
				logging.GetLogger(ctx).Debug(fmt.Sprintf("hpfn_config_tab=%s", s))
				ch <- struct{}{}
			}

			<-tick.C
		}
	}()
	<-ch
}

func (h *HpfnConfigRepoImpl) GetAllHpfnConfigFromDb(ctx context.Context) ([]*sip_db.HpfnConfig, error) {
	session := h.sipRepo.DbSession()
	return h.sipRepo.GetAllHpfnConfig(ctx, session)
}

func (h *HpfnConfigRepoImpl) GetAllHpfnConfig(ctx context.Context) (map[string][]*sip_db.HpfnConfig, error) {
	h.mu.Lock()
	m := h.localCache
	h.mu.Unlock()
	return m, nil
}

func (h *HpfnConfigRepoImpl) UpdateHpfnConfigInLocalCache(ctx context.Context, data []*sip_db.HpfnConfig) {
	newLocalCache := map[string][]*sip_db.HpfnConfig{}
	for _, datum := range data {
		old, ok := newLocalCache[datum.HpfnKey]
		if !ok {
			old = []*sip_db.HpfnConfig{}
		}
		newLocalCache[datum.HpfnKey] = append(old, datum)
	}
	for _, v := range newLocalCache {
		sort.Slice(v, func(i, j int) bool {
			return v[i].WeightRange < v[j].WeightRange
		})
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.localCache = newLocalCache
}

func (c *HpfnConfigRepoImpl) GetOne(ctx context.Context, hpfnKey string, weight int) (*sip_db.HpfnConfig, error) {
	var m map[string][]*sip_db.HpfnConfig
	c.mu.RLock()
	defer c.mu.RUnlock()
	m = c.localCache

	cfgs, ok := m[hpfnKey]
	if ok {
		for _, cfg := range cfgs {
			if int(cfg.WeightRange) >= weight {
				return cfg, nil
			}
		}
	}
	return nil, cerr.New("hpfn_config not found", uint32(pb.Constant_ERROR_NOT_FOUND))
}
