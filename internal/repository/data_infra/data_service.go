package data_infra

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"git.garena.com/shopee/core-server/core-logic/clog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/config"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/model"
)

const (
	outputChanSize = 10000
)

type DataService struct {
	cfg     *config.DataInfraConfig
	scanner *Scanner
}

func NewDataService() *DataService {
	d := &DataService{
		cfg: config.GetDataInfraConfig(),
	}

	d.scanner = NewScanner(d.cfg.ClientId, d.cfg.ClientSecret).
		WithBufferSize(outputChanSize).
		WithErrorHandler(func(err error) {
			clog.Errorf(context.Background(), "error from DS: %s", err.Error())
		})

	return d
}

func (d *DataService) GetOrderMartExchangeRateList(ctx context.Context) ([]*model.OrderMartExchangeRate, bool) {
	dataList := make([]*model.OrderMartExchangeRate, 0)
	success := false

	dsRowCh := d.scanner.Scan(ctx, d.cfg.OrderMartExchangeRate.API, d.cfg.OrderMartExchangeRate.Version, nil)
	for {
		var r interface{}
		var ok bool

		select {
		case r, ok = <-dsRowCh:
			if !ok {
				return dataList, success
			}
		}

		data, err := d.dsRowToOrderMartExchangeRate(r)
		if err != nil {
			clog.Errorf(ctx, "error converting DS row: %s", err.Error())
			continue
		}

		dataList = append(dataList, data)
		success = true
	}
}

func (d *DataService) dsRowToOrderMartExchangeRate(r interface{}) (*model.OrderMartExchangeRate, error) {
	rMap, ok := r.(map[string]interface{})
	if !ok {
		return nil, errors.New("r's concrete type is not map[string]interface{}")
	}

	var dsRow struct {
		RowNumber int                         `json:"rowNumber"`
		Row       model.OrderMartExchangeRate `json:"values"`
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json", // still allow the round-trip json marshal/unmarshal trick
		Result:  &dsRow,
	})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(rMap)
	if err != nil {
		return nil, err
	}

	err = dsRow.Row.Validate()
	if err != nil {
		return nil, err
	}

	return &dsRow.Row, nil
}
