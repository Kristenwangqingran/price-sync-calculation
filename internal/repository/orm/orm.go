package orm

import (
	"git.garena.com/shopee/common/gdbc/gdbc"
)

type DbSession interface {
	gdbc.Session
}

type DbSessionFactory interface {
	DbSession() DbSession
}
