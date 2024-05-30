package edit_item_price_allow_list

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/wire"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

var ProviderSet = wire.NewSet(
	NewPShopWhiteListToEditItemPriceRepoImpl,
	wire.Bind(new(PShopWhiteListToEditItemPriceRepo), new(*PShopWhiteListToEditItemPriceRepoImpl)),
)

type PShopWhiteListToEditItemPriceRepo interface {
	Exist(ctx context.Context, pShopId uint64) bool
}

type PShopWhiteListToEditItemPriceRepoImpl struct {
	mu         *sync.RWMutex
	once       *sync.Once
	localCache map[int64]struct{} // pShopId set

	sipRepo sip_db.SipRepo
}

func NewPShopWhiteListToEditItemPriceRepoImpl(sipRepo sip_db.SipRepo) *PShopWhiteListToEditItemPriceRepoImpl {
	v := &PShopWhiteListToEditItemPriceRepoImpl{
		mu:         &sync.RWMutex{},
		once:       &sync.Once{},
		localCache: make(map[int64]struct{}),

		sipRepo: sipRepo,
	}
	v.init()
	return v
}

func (p *PShopWhiteListToEditItemPriceRepoImpl) init() {
	p.refresh()
	go func() {
		tick := time.NewTicker(30 * time.Second)
		for {
			<-tick.C
			p.refresh()
		}
	}()
}

func (p *PShopWhiteListToEditItemPriceRepoImpl) refresh() {
	session := p.sipRepo.DbSession()
	newRecords, err := p.sipRepo.GetAllEditItemPriceAllowList(context.Background(), session)
	if err != nil {
		logging.GetLogger(context.Background()).Error(fmt.Sprintf("failed to refresh edit_item_price_allow_list, err=%v", err))
		return
	}

	newMap := make(map[int64]struct{}, len(newRecords))
	for _, record := range newRecords {
		newMap[record.MstShopId] = struct{}{}
	}

	p.mu.Lock()
	p.localCache = newMap
	p.mu.Unlock()
	logging.GetLogger(context.Background()).Info(fmt.Sprintf("refreshed edit_item_price_allow_list, size=%v", len(newRecords)))
}

func (p *PShopWhiteListToEditItemPriceRepoImpl) Exist(ctx context.Context, pShopId uint64) bool {
	p.mu.RLock()
	_, exist := p.localCache[int64(pShopId)]
	p.mu.RUnlock()
	logging.GetLogger(ctx).Debug(fmt.Sprintf("editItemPriceAllowListCache, shopId=%v, exist=%v", pShopId, exist))
	return exist
}
