package model

import (
	"github.com/golang/protobuf/proto"
)

type AccountStatus = int32

const (
	StatusAccountDelete AccountStatus = 0
	StatusAccountNormal AccountStatus = 1
	StatusAccountBanned AccountStatus = 2
	StatusAccountFrozen AccountStatus = 3
)

type CNSCShop struct {
	Region *string `protobuf:"bytes,1,opt,name=region" json:"region"`
	ShopId *uint64 `protobuf:"varint,2,opt,name=shop_id" json:"shop_id"`
}

func (s *CNSCShop) GetShopId() uint64 {
	if s == nil || s.ShopId == nil {
		return 0
	}
	return *s.ShopId
}

func (s *CNSCShop) GetRegion() string {
	if s == nil || s.Region == nil {
		return ""
	}
	return *s.Region
}

func NewCnscShopListFromShopIdList(shopIdList []uint64) []*CNSCShop {
	shops := make([]*CNSCShop, 0, len(shopIdList))
	for _, shopId := range shopIdList {
		shops = append(shops, &CNSCShop{
			ShopId: proto.Uint64(shopId),
		})
	}
	return shops
}

type CheckCNSCWhiteListRequest struct {
	ShopIds []uint64 `json:"shop_ids"`
}

type CheckCNSCWhiteListResponse struct {
	Items []uint64 `json:"items"`
}

type GetCnscShopsByAccountIdResponse struct {
	Shops []*CNSCShop `protobuf:"bytes,1,opt,name=shops" json:"shops"`
}
