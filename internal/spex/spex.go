package spex

import (
	"context"
	"fmt"
	"time"

	"git.garena.com/shopee/common/spkit/pkg/spex"
	"git.garena.com/shopee/core-server/core-logic/cerr"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/platform/golang_splib/sps"
	sp_common "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

func callSPEX(ctx context.Context, api string, req, resp interface{}, reqOptions ...sps.RequestOption) error {
	agent, err := spex.SpsAgent()
	if err != nil {
		return cerr.Wrap(err, "get sps failed agent", uint32(pb.Constant_ERROR_INTERNAL))
	}

	header := &sps.Header{}

	reqOptions = append(reqOptions, header)

	startTime := time.Now()
	code := agent.RPCRequest(ctx, api, req, resp, reqOptions...)
	costInMs := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if code != uint32(sp_common.Constant_SUCCESS) {
		return cerr.New(fmt.Sprintf("cmd=%s|trace_id=%s|cost(ms)=%f|errcode=%d|resp=%s",
			api, header.TraceID(), costInMs, code, cutil.LazyJSONEncoder(resp)), code)
	}

	return nil
}
