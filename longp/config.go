package longp

import (
	"github.com/banbox/banexg"
)

// 交易所基础配置
const (
	ExchangeID   = "longp"
	ExchangeName = "Longp"
	RateLimit    = 50
	Min1mHole    = 1
)

// API 地址配置
const (
	ApiBaseUrl = "https://openapi.longportapp.com"
	WsBaseUrl  = "wss://openapi-quote.longportapp.com"
	RecvWindow = 30000
)

// 手续费配置
const (
	FeeSide   = "quote"
	TakerRate = 0.0002
	MakerRate = 0.0002
)

// 创建交易所实例
func createExchange(options map[string]interface{}) *banexg.Exchange {
	return &banexg.Exchange{
		ExgInfo:   createExgInfo(),
		RateLimit: RateLimit,
		Options:   options,
		Hosts:     &banexg.ExgHosts{},
		Fees:      createFees(),
		Has:       createApiSupport(),
	}
}

// 创建交易所基础信息
func createExgInfo() *banexg.ExgInfo {
	return &banexg.ExgInfo{
		ID:        ExchangeID,
		Name:      ExchangeName,
		Countries: []string{"CN"},
		FixedLvg:  true,
	}
}

// 创建手续费配置
func createFees() *banexg.ExgFee {
	return &banexg.ExgFee{
		Linear: &banexg.TradeFee{
			FeeSide:    FeeSide,
			TierBased:  false,
			Percentage: true,
			Taker:      TakerRate,
			Maker:      MakerRate,
		},
	}
}

// 创建API支持配置
func createApiSupport() map[string]map[string]int {
	return map[string]map[string]int{
		"": {
			banexg.ApiFetchTicker:           banexg.HasOk,
			banexg.ApiFetchTickers:          banexg.HasOk,
			banexg.ApiFetchTickerPrice:      banexg.HasOk,
			banexg.ApiLoadLeverageBrackets:  banexg.HasOk,
			banexg.ApiGetLeverage:           banexg.HasOk,
			banexg.ApiFetchOHLCV:            banexg.HasOk,
			banexg.ApiFetchOrderBook:        banexg.HasOk,
			banexg.ApiFetchOrder:            banexg.HasOk,
			banexg.ApiFetchOrders:           banexg.HasOk,
			banexg.ApiFetchBalance:          banexg.HasOk,
			banexg.ApiFetchAccountPositions: banexg.HasOk,
			banexg.ApiFetchPositions:        banexg.HasOk,
			banexg.ApiFetchOpenOrders:       banexg.HasOk,
			banexg.ApiCreateOrder:           banexg.HasOk,
			banexg.ApiEditOrder:             banexg.HasOk,
			banexg.ApiCancelOrder:           banexg.HasOk,
			banexg.ApiSetLeverage:           banexg.HasOk,
			banexg.ApiCalcMaintMargin:       banexg.HasOk,
			banexg.ApiWatchOrderBooks:       banexg.HasOk,
			banexg.ApiUnWatchOrderBooks:     banexg.HasOk,
			banexg.ApiWatchOHLCVs:           banexg.HasOk,
			banexg.ApiUnWatchOHLCVs:         banexg.HasOk,
			banexg.ApiWatchMarkPrices:       banexg.HasOk,
			banexg.ApiUnWatchMarkPrices:     banexg.HasOk,
			banexg.ApiWatchTrades:           banexg.HasOk,
			banexg.ApiUnWatchTrades:         banexg.HasOk,
			banexg.ApiWatchMyTrades:         banexg.HasOk,
			banexg.ApiWatchBalance:          banexg.HasOk,
			banexg.ApiWatchPositions:        banexg.HasOk,
			banexg.ApiWatchAccountConfig:    banexg.HasOk,
		},
	}
}
