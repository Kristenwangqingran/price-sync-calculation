package main

import (
	"os"

	"google.golang.org/protobuf/proto"

	"git.garena.com/shopee/common/spex-contrib/interceptor/logging"
	"git.garena.com/shopee/common/spex-contrib/interceptor/monitoring"
	"git.garena.com/shopee/common/spex-contrib/interceptor/tracing"
	"git.garena.com/shopee/common/spkit"
	"git.garena.com/shopee/common/spkit/app"
	"git.garena.com/shopee/common/spkit/pkg/spex"
	"git.garena.com/shopee/common/spkit/runtime"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/di"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/health"
	pb "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/proto/spex/gen/go/price_sync_price_calculation.pb"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/servicesetup"
)

func main() {
	initializeSpkit()

	config.InitializeSpexConfigs()
	servicesetup.InitializeHttpClient()

	spexApp := createSpexApp()
	httpApp := createHTTPApp()

	registerGlobalServerInterceptor()
	registerGlobalClientInterceptor()

	spkit.Run(spexApp, httpApp)
}

// initializeSpkit init spkit service
func initializeSpkit() {
	spkit.Init(
		spkit.ConfigPath("etc/service.yml"),
		spkit.HealthCheckFunction(health.HealthCheck),
		spkit.SpexSduID(os.Getenv("SPEX_SDU_ID")),
	)
}

// createSpexApp create spex app
func createSpexApp() *app.SpexApp {
	service, err := di.CreateApplication()
	if err != nil {
		panic(err)
	}

	service.AsyncDataLogic.AsyncSetOrderMartExchangeRate()

	spexApp, err := app.NewSpexApp(pb.NewCalculationServer(service))
	if err != nil {
		panic(err)
	}
	return spexApp
}

func createHTTPApp() *app.HTTPApp {
	// create http app
	// in localhost set a forward proxy http server so that it can simplify the use of spex gateway
	// send request to 127.0.0.1 like '127.0.0.1:9001/sprpc/price.sync_price.core.get_primary_item_price_list'
	// no need to set the x-sp-destination and serve rule param
	httpApp, err := app.NewHTTPApp(app.ConfigKey("httpserver"))
	if err != nil {
		panic(err)
	}
	httpApp.Handle("/", NewHTTPHandler())
	httpApp.Handle("/health_check", &healthCheckHandler{})

	return httpApp
}

// registerGlobalServerInterceptor register server side global interceptors for log, metric and tracing
func registerGlobalServerInterceptor() {
	// create global server interceptors
	loggingInterceptor, err := logging.NewServerInterceptor(
		logging.IDC(runtime.IDC()),
		logging.ServerConfig(&logging.ServerInterceptorConfig{
			IncludeRequest:      proto.Bool(true),
			IncludeResponse:     true,
			IncludeRequestSize:  proto.Bool(true),
			IncludeResponseSize: proto.Bool(true),
			SpecificRules:       getCommandLoggingConfig(),
		}),
	)
	if err != nil {
		panic(err)
	}

	monitoringInterceptor, err := monitoring.NewServerInterceptor(
		spkit.DefaultService().Name(),
	)
	if err != nil {
		panic(err)
	}

	tracingInterceptor, err := tracing.NewServerInterceptor()
	if err != nil {
		panic(err)
	}

	// register global server interceptors
	err = spex.RegisterGlobalServerInterceptors(
		loggingInterceptor.InterceptorFunc(),
		monitoringInterceptor.InterceptorFunc(),
		tracingInterceptor.InterceptorFunc(),
	)
	if err != nil {
		panic(err)
	}
}

func getCommandLoggingConfig() []logging.ServerInterceptorSpecificConfig {
	cmdInterceptorConfig := make([]logging.ServerInterceptorSpecificConfig,
		0, len(config.GetCmdIgnoreReqRespInLogListConfig()))

	// set on cmd level based on spex config if needs to print request or response.
	for _, cmdIgnoreReqRespInLog := range config.GetCmdIgnoreReqRespInLogListConfig() {
		cmdInterceptorConfig = append(cmdInterceptorConfig,
			logging.ServerInterceptorSpecificConfig{
				IncludeRequest:      proto.Bool(!cmdIgnoreReqRespInLog.IgnoreReq),
				IncludeResponse:     proto.Bool(!cmdIgnoreReqRespInLog.IgnoreResp),
				IncludeRequestSize:  proto.Bool(true),
				IncludeResponseSize: proto.Bool(true),
				Commands:            []string{cmdIgnoreReqRespInLog.Command},
			})
	}

	return cmdInterceptorConfig
}

// registerGlobalClientInterceptor register client side global interceptors for log and metric
func registerGlobalClientInterceptor() {
	// create global client interceptors
	clientLoggingInterceptor, err := logging.NewClientInterceptor()
	if err != nil {
		panic(err)
	}

	clientMonitoringInterceptor, err := monitoring.NewClientInterceptor(
		spkit.DefaultService().Name(),
	)
	if err != nil {
		panic(err)
	}

	// register global client interceptors
	err = spex.RegisterGlobalClientInterceptors(
		clientLoggingInterceptor.InterceptorFunc(),
		clientMonitoringInterceptor.InterceptorFunc(),
	)
	if err != nil {
		panic(err)
	}
}
