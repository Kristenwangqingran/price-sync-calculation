package miscellaneousinner

import (
	"context"
	"net/http"

	iRouter "git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/miscellaneousinner/router"

	"git.garena.com/shopee/platform/golang_splib/desc"
	httpdesc "git.garena.com/shopee/platform/golang_splib/desc/http"
	"git.garena.com/shopee/platform/golang_splib/errors"
	sphttp "git.garena.com/shopee/platform/golang_splib/http"
	sp_common "git.garena.com/shopee/sp_protocol/golang/common.pb"
)

type service struct {
	methods []desc.Method
}

func (s *service) Name() string {
	return "Miscellaneousinner"
}

func (s *service) Methods() []desc.Method {
	return s.methods
}

func (s *service) Backend() desc.Backend {
	return desc.HTTP
}

type Miscellaneousinner interface {
	GetApiInnerMiscellaneousExchangeRateGet(ctx sphttp.RequestCtx) error
	GetApiInnerMiscellaneousErrorConfigList(ctx sphttp.RequestCtx) error
	GetTest(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousPopupWindowEdit(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousPopupWindowEdit1OrAdd(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousPopupWindowAdd(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousPopupWindowDelete(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousPopupWindowRead(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousWindowUser(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousWindowBatchSet(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousV1PayPwdFlowCreateFlow(ctx sphttp.RequestCtx) error
	PostApiInnerMiscellaneousV1PayPwdFlowCheckFlow(ctx sphttp.RequestCtx) error
}

func newService(impl Miscellaneousinner) desc.Service {
	ms := []desc.Method{}

	getApiInnerMiscellaneousExchangeRateGet := getApiInnerMiscellaneousExchangeRateGet{
		impl: impl,
	}

	ms = append(ms, &getApiInnerMiscellaneousExchangeRateGet)

	getApiInnerMiscellaneousErrorConfigList := getApiInnerMiscellaneousErrorConfigList{
		impl: impl,
	}

	ms = append(ms, &getApiInnerMiscellaneousErrorConfigList)

	getTest := getTest{
		impl: impl,
	}

	ms = append(ms, &getTest)

	postApiInnerMiscellaneousPopupWindowEdit := postApiInnerMiscellaneousPopupWindowEdit{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousPopupWindowEdit)

	postApiInnerMiscellaneousPopupWindowEdit1OrAdd := postApiInnerMiscellaneousPopupWindowEdit1OrAdd{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousPopupWindowEdit1OrAdd)

	postApiInnerMiscellaneousPopupWindowAdd := postApiInnerMiscellaneousPopupWindowAdd{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousPopupWindowAdd)

	postApiInnerMiscellaneousPopupWindowDelete := postApiInnerMiscellaneousPopupWindowDelete{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousPopupWindowDelete)

	postApiInnerMiscellaneousPopupWindowRead := postApiInnerMiscellaneousPopupWindowRead{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousPopupWindowRead)

	postApiInnerMiscellaneousWindowUser := postApiInnerMiscellaneousWindowUser{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousWindowUser)

	postApiInnerMiscellaneousWindowBatchSet := postApiInnerMiscellaneousWindowBatchSet{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousWindowBatchSet)

	postApiInnerMiscellaneousV1PayPwdFlowCreateFlow := postApiInnerMiscellaneousV1PayPwdFlowCreateFlow{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousV1PayPwdFlowCreateFlow)

	postApiInnerMiscellaneousV1PayPwdFlowCheckFlow := postApiInnerMiscellaneousV1PayPwdFlowCheckFlow{
		impl: impl,
	}

	ms = append(ms, &postApiInnerMiscellaneousV1PayPwdFlowCheckFlow)

	hs := service{
		methods: ms,
	}

	return &hs
}

type getApiInnerMiscellaneousExchangeRateGet struct {
	impl Miscellaneousinner
}

func (s *getApiInnerMiscellaneousExchangeRateGet) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.get_api_inner_miscellaneous_exchange05frate_get"
}

func (s *getApiInnerMiscellaneousExchangeRateGet) RequestType() interface{} {
	return &http.Request{}
}

func (s *getApiInnerMiscellaneousExchangeRateGet) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *getApiInnerMiscellaneousExchangeRateGet) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.GetApiInnerMiscellaneousExchangeRateGet(req)

		return nil, err
	}
}

type getApiInnerMiscellaneousErrorConfigList struct {
	impl Miscellaneousinner
}

func (s *getApiInnerMiscellaneousErrorConfigList) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.get_api_inner_miscellaneous_error05fconfig_list"
}

func (s *getApiInnerMiscellaneousErrorConfigList) RequestType() interface{} {
	return &http.Request{}
}

func (s *getApiInnerMiscellaneousErrorConfigList) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *getApiInnerMiscellaneousErrorConfigList) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.GetApiInnerMiscellaneousErrorConfigList(req)

		return nil, err
	}
}

type getTest struct {
	impl Miscellaneousinner
}

func (s *getTest) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.get_test"
}

func (s *getTest) RequestType() interface{} {
	return &http.Request{}
}

func (s *getTest) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *getTest) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.GetTest(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousPopupWindowEdit struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousPopupWindowEdit) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_popup05fwindow_edit"
}

func (s *postApiInnerMiscellaneousPopupWindowEdit) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousPopupWindowEdit) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousPopupWindowEdit) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousPopupWindowEdit(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousPopupWindowEdit1OrAdd struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousPopupWindowEdit1OrAdd) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_popup05fwindow_edit05for05fadd"
}

func (s *postApiInnerMiscellaneousPopupWindowEdit1OrAdd) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousPopupWindowEdit1OrAdd) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousPopupWindowEdit1OrAdd) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousPopupWindowEdit1OrAdd(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousPopupWindowAdd struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousPopupWindowAdd) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_popup05fwindow_add"
}

func (s *postApiInnerMiscellaneousPopupWindowAdd) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousPopupWindowAdd) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousPopupWindowAdd) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousPopupWindowAdd(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousPopupWindowDelete struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousPopupWindowDelete) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_popup05fwindow_delete"
}

func (s *postApiInnerMiscellaneousPopupWindowDelete) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousPopupWindowDelete) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousPopupWindowDelete) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousPopupWindowDelete(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousPopupWindowRead struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousPopupWindowRead) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_popup05fwindow_read"
}

func (s *postApiInnerMiscellaneousPopupWindowRead) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousPopupWindowRead) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousPopupWindowRead) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousPopupWindowRead(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousWindowUser struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousWindowUser) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_window_user"
}

func (s *postApiInnerMiscellaneousWindowUser) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousWindowUser) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousWindowUser) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousWindowUser(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousWindowBatchSet struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousWindowBatchSet) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_window_batch05fset"
}

func (s *postApiInnerMiscellaneousWindowBatchSet) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousWindowBatchSet) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousWindowBatchSet) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousWindowBatchSet(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousV1PayPwdFlowCreateFlow struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCreateFlow) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_v1_pay05fpwd05fflow_create05fflow"
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCreateFlow) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCreateFlow) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCreateFlow) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousV1PayPwdFlowCreateFlow(req)

		return nil, err
	}
}

type postApiInnerMiscellaneousV1PayPwdFlowCheckFlow struct {
	impl Miscellaneousinner
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCheckFlow) Command() string {
	return "mkt_http.sellerplatform.miscellaneousinner.post_api_inner_miscellaneous_v1_pay05fpwd05fflow_check05fflow"
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCheckFlow) RequestType() interface{} {
	return &http.Request{}
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCheckFlow) ResponseType() interface{} {
	// nil, golang_splib will handle this
	return nil
}

func (s *postApiInnerMiscellaneousV1PayPwdFlowCheckFlow) Handler() desc.Handler {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(sphttp.RequestCtx)
		if !ok {
			return nil, errors.FromCode(uint32(sp_common.Constant_ERROR_PARAMS))
		}

		err := s.impl.PostApiInnerMiscellaneousV1PayPwdFlowCheckFlow(req)

		return nil, err
	}
}

func router() httpdesc.Router {
	rs := []httpdesc.Rule{
		{Method: "get", Path: "/api/inner/miscellaneous/exchange_rate/get", Command: "mkt_http.sellerplatform.miscellaneousinner.get_api_inner_miscellaneous_exchange05frate_get"},
	}

	return iRouter.New(rs)
}
