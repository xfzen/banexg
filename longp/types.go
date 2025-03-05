package longp

import "github.com/banbox/banexg"

// 常量定义
const (
	HostPublic  = "public"
	HostPrivate = "private"
	HostWs      = "ws"
)

// 选项常量
const (
	OptRecvWindow  = "RecvWindow"
	OptAppKey      = "appKey"
	OptAppSecret   = "appSecret"
	OptAccessToken = "accessToken"
)

// 默认市场类型
var DefCareMarkets = []string{
	banexg.MarketSpot,
	banexg.MarketLinear,
	banexg.MarketInverse,
}

// 市场类型映射
var MarketTypeMap = map[string]string{
	"HK": "HK",
	"US": "US",
	"CN": "CN",
	"SG": "SG",
	"JP": "JP",
}

// 交易所映射
var ExchangeMap = map[string]string{
	"HK": "HKEX",
	"US": "NYSE",
	"CN": "SSE",
	"SG": "SGX",
	"JP": "TSE",
}
