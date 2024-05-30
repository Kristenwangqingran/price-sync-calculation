package config

import (
	"context"
	"fmt"

	"git.garena.com/shopee/common/gdbc/gdbc"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
)

type ListingDbSession *gdbc.DB
type PriceSyncDBSession *gdbc.DB
type SipDbSession *gdbc.DB
type SipV2DbSession *gdbc.DB
type AuditLogDBSessione *gdbc.DB

var (
	// TODO: rename after decoupling with listing DB
	listingDbClient   *gdbc.DB
	priceSyncDBClient *gdbc.DB

	sipDB      *gdbc.DB
	sipV2DB    *gdbc.DB
	auditLogDB *gdbc.DB
)

func applyMerchantConfigDBConfig() {
	if confVal == nil || confVal.MerchantConfigDBConfig == nil {
		panic("failed to init DB client, since MerchantConfigDBConfig is nil")
	}

	db, err := gdbc.Hardy("mysql", *confVal.MerchantConfigDBConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to init MerchantConfigDB client, err=%s", err.Error()))
	}

	listingDbClient = db
	logging.GetLogger(context.Background()).Info("success to init MerchantConfigDB client")
}

func applyPriceSyncDBConfig() {
	if confVal == nil || confVal.PriceSyncDBConfig == nil {
		panic("failed to init DB client, since PriceSyncDBConfig is nil")
	}

	db, err := gdbc.Hardy("mysql", *confVal.PriceSyncDBConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to init PriceSyncDBConfig client, err=%s", err.Error()))
	}

	priceSyncDBClient = db
	logging.GetLogger(context.Background()).Info("success to init price client")
}

func GetListingDbClient() ListingDbSession {
	return listingDbClient
}

func GetPriceSyncDBClient() PriceSyncDBSession {
	return priceSyncDBClient
}

func applySipDBConfig() {
	if confVal == nil || confVal.SipDBCfg == nil {
		panic("failed to init DB client, since SipDBCfg is nil")
	}

	db, err := gdbc.Hardy("mysql", *confVal.SipDBCfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init sipDB client, err=%s", err.Error()))
	}

	sipDB = db
	logging.GetLogger(context.Background()).Info("success to init sipDB client")
}

func GetSipDbClient() SipDbSession {
	return sipDB
}

func applySipV2DBConfig() {
	if confVal == nil || confVal.SipV2DBCfg == nil {
		panic("failed to init DB client, since SipV2DBCfg is nil")
	}

	db, err := gdbc.Hardy("mysql", *confVal.SipV2DBCfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init sipV2DB client, err=%s", err.Error()))
	}

	sipV2DB = db
	logging.GetLogger(context.Background()).Info("success to init sipV2DB client")
}

func GetSipV2DbClient() SipV2DbSession {
	return sipV2DB
}

func applyAuditLogConfig() {
	if confVal == nil || confVal.AuditLogDBCfg == nil {
		panic("failed to init DB client, since AuditLogDBCfg is nil")
	}

	db, err := gdbc.Hardy("mysql", *confVal.AuditLogDBCfg)
	if err != nil {
		panic(fmt.Sprintf("failed to init AuditLogDB client, err=%s", err.Error()))
	}

	auditLogDB = db
	logging.GetLogger(context.Background()).Info("success to init AuditLogDB client")
}

func GetAuditLogDBClient() AuditLogDBSessione {
	return auditLogDB
}
