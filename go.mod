module git.garena.com/shopee/core-server/price/price-sync-calculation

go 1.17

require (
	git.garena.com/shopee/common/cache v0.17.1
	git.garena.com/shopee/common/gdbc/datum v0.1.0
	git.garena.com/shopee/common/gdbc/hardy v0.4.1
	git.garena.com/shopee/common/spex-contrib/interceptor/logging v0.10.1
	git.garena.com/shopee/common/spex-contrib/interceptor/monitoring v0.9.1
	git.garena.com/shopee/common/spex-contrib/interceptor/tracing v0.5.0
	git.garena.com/shopee/common/spkit v0.19.0
	git.garena.com/shopee/common/ulog v0.2.3
	git.garena.com/shopee/common/uniconfig v0.9.0
	git.garena.com/shopee/core-server/core-logic v0.0.93
	git.garena.com/shopee/core-server/internal-tools/depck v0.4.0
	git.garena.com/shopee/platform/golang_splib v1.1.7-rc
	git.garena.com/shopee/seller-server/seller-listing/service-kits/ssk v1.1.24
	git.garena.com/shopee/sp_protocol v1.3.10-rc2
	github.com/golang/protobuf v1.5.3
	github.com/google/wire v0.5.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.3.1
	google.golang.org/protobuf v1.29.0
)

require (
	git.garena.com/shopee/common/gdbc/parser v0.4.0 // indirect
	git.garena.com/shopee/common/observability_config v0.1.0 // indirect
	git.garena.com/shopee/golang_splib v0.3.3 // indirect
	git.garena.com/shopee/mts/go-application-server/spi/spex v0.1.0 // indirect
	git.garena.com/shopee/platform/tracing-contrib/dynamic-sampler v0.1.1 // indirect
	git.garena.com/zhouz/redis.v3 v0.0.0-20171113075242-c73350239ae6 // indirect
	github.com/DATA-DOG/go-sqlmock v1.5.0 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/google/subcommands v1.0.1 // indirect
	github.com/pingcap/errors v0.11.5-0.20210425183316-da1aaba5fb63 // indirect
	github.com/pingcap/log v0.0.0-20210625125904-98ed8e2eb1c7 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/tools v0.1.5 // indirect
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20160220154919-db14e161995a // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

require (
	git.garena.com/common/gommon/crypt v0.0.0-20210211075301-867bb6bc3c33 // indirect
	git.garena.com/shopee/common/gdbc/gdbc v0.3.4
	git.garena.com/shopee/common/gomemcache v0.6.1 // indirect
	git.garena.com/shopee/common/jsonext v0.1.0 // indirect
	git.garena.com/shopee/common/log v0.3.0 // indirect
	git.garena.com/shopee/common/redigo v1.8.8-fifo // indirect
	git.garena.com/shopee/common/spex-contrib/interceptor/gateway_protocol v0.1.0 // indirect
	git.garena.com/shopee/core-server/hashring v0.0.1 // indirect
	git.garena.com/shopee/core-server/hashring-server-selector v0.0.4 // indirect
	git.garena.com/shopee/data-infra/dataservice-sdk-golang v0.5.2
	git.garena.com/shopee/devops/golang_aegislib v0.0.10 // indirect
	git.garena.com/shopee/platform/config-sdk-go v0.6.1 // indirect
	git.garena.com/shopee/platform/config-sdk-go/adapter/seller-config-sdk v0.3.0 // indirect
	git.garena.com/shopee/platform/jaeger-tracer v1.7.0 // indirect
	git.garena.com/shopee/platform/local-tracer v1.2.0 // indirect
	git.garena.com/shopee/platform/service-governance/observability/metric v1.0.16
	git.garena.com/shopee/platform/service-governance/viewercontext v1.0.12
	git.garena.com/shopee/platform/splog v1.4.10 // indirect
	git.garena.com/shopee/platform/trace v0.1.0 // indirect
	git.garena.com/shopee/platform/tracing v1.11.1 // indirect
	git.garena.com/shopee/seller-server/mktzlib/http_over_spex v1.1.4
	git.garena.com/shopee/seller-server/seller-data/config-center-sdk/go v0.0.3 // indirect
	git.garena.com/shopee/seller-server/seller-listing/seller_error_go v0.1.5-0.20221017060306-ddd2a1f9ff55 // indirect
	git.garena.com/shopee/seller-server/seller-listing/service-kits/ssf v1.0.12
	git.garena.com/shopee/seller-server/sip/sip-gocommon v1.25.0
	github.com/andybalholm/brotli v1.0.0 // indirect
	github.com/avast/retry-go v2.7.0+incompatible // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b // indirect
	github.com/brentp/intintmap v0.0.0-20190211203843-30dc0ade9af9 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/getsentry/sentry-go v0.11.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.7.7 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/mock v1.4.4
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.10.7 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6 // indirect
	github.com/panjf2000/ants/v2 v2.4.6 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.3.4 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.4.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.8.1
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber/jaeger-client-go v2.22.1+incompatible // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.17.0 // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/zhenjl/cityhash v0.0.0-20131128155616-cdd6a94144ab // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.20.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
	gopkg.in/ini.v1 v1.54.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
