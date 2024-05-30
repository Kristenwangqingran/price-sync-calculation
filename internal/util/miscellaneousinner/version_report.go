package miscellaneousinner

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	sphttp "git.garena.com/shopee/platform/golang_splib/http"
	"git.garena.com/shopee/platform/golang_splib/http/middleware"
	"git.garena.com/shopee/platform/golang_splib/sps"
	"git.garena.com/shopee/platform/golang_splib/util"
	"git.garena.com/shopee/platform/service-governance/observability/metric"
)

type reporter struct {
	version     string
	serviceName string
	next        http.RoundTripper
}

func (r *reporter) RoundTrip(req *http.Request) (*http.Response, error) {
	labels := map[string]string{"version": r.version, "type": "client", "caller": currentServiceName(), "callee": r.serviceName}
	_ = metric.RPCCustomReporter.ReportCounter("spsvg_stub_version", labels, 1)

	return r.next.RoundTrip(req)
}

func ClientVersionReport(version, serviceName string) middleware.ClientMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return &reporter{version: version, serviceName: serviceName, next: next}
	}
}

func ServerVersionReport(version string) middleware.ServerMiddleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx sphttp.RequestCtx) error {
			header := sps.FromIncomingContext(ctx.Request.Context())
			peerServiceName, _, _, _, _, _, _ := util.ParseInstanceID(header.InstanceID())
			labels := map[string]string{"version": version, "type": "server", "callee": currentServiceName(), "caller": peerServiceName}
			_ = metric.RPCCustomReporter.ReportCounter("spsvg_stub_version", labels, 1)

			return handler(ctx)
		}
	}
}

func currentServiceName() string {
	pro := strings.ToLower(os.Getenv("PROJECT_NAME"))
	mod := strings.ToLower(os.Getenv("MODULE_NAME"))

	return fmt.Sprintf("%s-%s", pro, mod)
}
