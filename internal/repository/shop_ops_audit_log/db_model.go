package shop_ops_audit_log

import "git.garena.com/shopee/common/gdbc/gdbc/tablereflect"

type OpsShopLog struct {
	Id        int64  `gdbc:"primary_key=true, column=schema_id"`
	AuditType int    `gdbc:"column=audit_type"`
	UserId    int64  `gdbc:"column=column:userid"` // need to use column to specify column name, otherwise will use 'use_id'
	EntityId  string `gdbc:"column=column:entityid"`
	AuxId     string `gdbc:"column=column:auxid"`
	OldValue  string `gdbc:"column=old_value"`
	NewValue  string `gdbc:"column=new_value"`
	Result    string `gdbc:"column=result"`
	Extinfo   string `gdbc:"column=extinfo"`
	Ctime     int64  `gdbc:"column=ctime"`
	Mtime     int64  `gdbc:"column=mtime"`
}

func init() {
	tablereflect.TypeInit(
		&OpsShopLog{},
		tablereflect.Table("ops_shop_log_tab"),
	)
}
