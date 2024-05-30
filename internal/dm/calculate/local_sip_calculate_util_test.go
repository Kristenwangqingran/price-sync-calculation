package calculate

import (
	"context"
	"testing"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/calcutil"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

func TestCalcAffiPriceForLocalWithOverseaDiscount(t *testing.T) {
	type args struct {
		ctx                 context.Context
		primaryPrice        float64
		affiRealWeight      float64
		itemMargin          float64
		shopMargin          float64
		shippingFee         float64
		initHiddenPrice     float64
		localPriceConfig    *model.CommonPriceConfig
		overseaDiscountRate float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			// copy live data from P-itemId=7403605465, P-modelId=51163303439, A-region=MY
			name: "calc test",
			args: args{
				ctx:             context.Background(),
				primaryPrice:    82900,
				affiRealWeight:  1200.0,
				itemMargin:      1.0,
				shopMargin:      1.0,
				shippingFee:     10.80,
				initHiddenPrice: 2.300000,
				localPriceConfig: &model.CommonPriceConfig{
					Buffer:          proto.Float64(1.22),
					ExchangeRate:    proto.Float64(3400),
					InitHiddenPrice: proto.Float64(2.3),
				},
				overseaDiscountRate: 0.0,
			},
			want: 45.728471,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcutil.CalculateAffiPriceForLocalSip(tt.args.ctx, tt.args.primaryPrice, tt.args.affiRealWeight, tt.args.itemMargin, tt.args.shopMargin, tt.args.initHiddenPrice, tt.args.shippingFee, tt.args.localPriceConfig, &tt.args.overseaDiscountRate)
			if !assert.InDelta(t, tt.want, got, 0.0001) {
				t.Errorf("CalcAffiPriceForLocalWithOverseaDiscount() = %v, want %v", got, tt.want)
			}
		})
	}
}
