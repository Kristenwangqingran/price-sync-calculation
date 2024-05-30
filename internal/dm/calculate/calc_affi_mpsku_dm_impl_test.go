package calculate

import (
	"context"
	"testing"

	mockfactors "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/mock/repository/factors"
	internal "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/internal_sip.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/factors"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/dm/data"
	mockdb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/mock/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
	db2 "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/repository/sip_db"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/service"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/testconfig"
)

func Test_calculateAffiMpskuDm_doCalcLocalSipOverseaDiscount(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	cfg := &config.Config{
		SIPMigrationCfg: testconfig.BuildSIPConfigFromLiveCfg(),
	}

	config.SetGlobal(cfg)

	primaryRegion := "ID"
	affiRegion := "MY"

	primaryShopId := uint64(67691500)
	primaryItemId := uint64(7403605465)
	primaryModelId := uint64(51163303439)

	primaryItemModelId := model.ItemModelId{
		ItemId:  primaryItemId,
		ModelId: primaryModelId,
	}

	affiShopId := int64(273320833)
	affiItemId := uint64(3838352825)
	affiModelId := uint64(61163290778)

	affiItemModelId := model.ItemModelId{
		ItemId:  affiItemId,
		ModelId: affiModelId,
	}

	calculationFactorsRepo := mockfactors.NewMockCalculationFactorsRepo(ctrl)
	calculationFactorsRepo.EXPECT().
		GetInitialHiddenPriceForLocalSip(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]model.LocalSipHiddenPriceResult{
			{
				QueryId:     0,
				Err:         nil,
				HiddenPrice: 10,
			},
		}, nil)

	localShippingFeeConfigDB := mockdb.NewMockLocalShippingFeeConfigDB(ctrl)
	localShippingFeeConfigDB.EXPECT().
		GetLocalShippingFeeConfigRecordByWeight(ctx, primaryRegion, affiRegion, int64(120000)).
		Return(&sip_db.LocalShippingFeeConfigRecord{ShippingFeePrice: 1089000}, nil)

	type fields struct {
		calculationFactorsRepo   factors.CalculationFactorsRepo
		localShippingFeeConfigDB db2.LocalShippingFeeConfigDB
		logisticService          service.LogisticService
	}
	type args struct {
		ctx             context.Context
		affiShopId      int64
		affiRegion      string
		discountRate    int64
		affiItemModelId model.ItemModelId
		calcData        *data.CalcFactorDataForAffiMpsku
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "calc test",
			fields: fields{
				calculationFactorsRepo:   calculationFactorsRepo,
				localShippingFeeConfigDB: localShippingFeeConfigDB,
				logisticService:          nil,
			},
			args: args{
				ctx:             ctx,
				affiShopId:      affiShopId,
				affiRegion:      affiRegion,
				discountRate:    0.0,
				affiItemModelId: affiItemModelId,
				calcData: &data.CalcFactorDataForAffiMpsku{
					PrimaryShopId: primaryShopId,
					PrimaryRegion: primaryRegion,
					AffiRegion:    affiRegion,
					ItemModelMapping: map[model.ItemModelId]model.ItemModelId{
						affiItemModelId: primaryItemModelId,
					},

					PrimaryOriginPrices: map[model.ItemModelId]int64{
						primaryItemModelId: 8290000000,
					},

					AffiOriginPrices: map[model.ItemModelId]int64{
						affiItemModelId: 4590000,
					},

					LocalPriceConfig: &model.CommonPriceConfig{
						Buffer:            proto.Float64(1.22),
						ExchangeRate:      proto.Float64(3400),
						InitHiddenPrice:   proto.Float64(2.3),
						ShippingFeeToggle: proto.Int32(model.PriceSyncToggleReadFromDB),
						HiddenPriceToggle: proto.Int32(model.PriceSyncToggleReadFromDB),
					},

					ShopMargin: 100000,

					PrimaryItemData: map[uint64]service.PrimaryItemData{
						primaryItemId: {
							PrimaryItemId: primaryItemId,
							Weight:        120000,
						},
					},

					AItemIdToPItemIdMapping: map[uint64]uint64 {
						primaryItemId: affiItemId,
					},

					AItemData: map[uint64]*internal.AItemData{
						affiItemId: {
							AffiItemid:    proto.Uint64(affiItemId),

							ItemMargin:     proto.Int32(0),
							AffiRealWeight: proto.Int32(0),
						},
					},
				},
			},
			want: 5550000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &calculateAffiMpskuDm{
				calculationFactorsRepo:   tt.fields.calculationFactorsRepo,
				localShippingFeeConfigDB: tt.fields.localShippingFeeConfigDB,
				logisticService:          tt.fields.logisticService,
			}

			result := dm.doCalcLocalSipOverseaDiscount(tt.args.ctx, tt.args.affiShopId, tt.args.affiRegion, tt.args.discountRate, tt.args.affiItemModelId, tt.args.calcData)

			assert.Equal(t, tt.want, result.GetAffiPrice())
		})
	}
}
