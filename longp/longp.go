package longp

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/shopspring/decimal"
)

type Longp struct {
	*banexg.Exchange
}

func New(Options map[string]interface{}) (*Longp, *errs.Error) {
	exg := &Longp{
		Exchange: &banexg.Exchange{
			ExgInfo: &banexg.ExgInfo{
				ID:        "longp",
				Name:      "Longp",
				Countries: []string{"CN"},
				FixedLvg:  true,
			},
			RateLimit: 50,
			Options:   Options,
			Hosts:     &banexg.ExgHosts{},
			Fees: &banexg.ExgFee{
				Linear: &banexg.TradeFee{
					FeeSide:    "quote",
					TierBased:  false,
					Percentage: true,
					Taker:      0.0002,
					Maker:      0.0002,
				},
			},
			Has: map[string]map[string]int{
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
			},
		},
	}

	err := exg.Init()
	if err != nil {
		return nil, err
	}

	exg.CalcFee = makeCalcFee(exg)
	return exg, nil
}

func NewExchange(Options map[string]interface{}) (banexg.BanExchange, *errs.Error) {
	return New(Options)
}

func (e *Longp) Init() *errs.Error {
	err := e.Exchange.Init()
	if err != nil {
		return err
	}

	e.ExgInfo.Min1mHole = 1

	if err := e.initConnection(); err != nil {
		return err
	}

	if err := e.initMarkets(); err != nil {
		return err
	}

	return nil
}

func (e *Longp) initConnection() *errs.Error {
	apiKey, _ := e.Options["apiKey"].(string)
	secret, _ := e.Options["secret"].(string)

	if apiKey == "" || secret == "" {
		return errs.NewMsg(errs.CodeParamRequired, "apiKey and secret are required")
	}

	e.Hosts.Prod = map[string]string{
		"public":  "https://api.longp.com",
		"private": "https://api.longp.com",
		"ws":      "wss://ws.longp.com/stream",
	}

	return nil
}

func (e *Longp) initMarkets() *errs.Error {
	if e.Markets == nil {
		e.Markets = make(banexg.MarketMap)
	}
	if e.MarketsById == nil {
		e.MarketsById = make(banexg.MarketArrMap)
	}
	return nil
}

// 市场数据接口
func (e *Longp) LoadMarkets(reload bool, params map[string]interface{}) (banexg.MarketMap, *errs.Error) {
	if len(e.Markets) > 0 && !reload {
		return e.Markets, nil
	}
	// TODO: 实现市场数据加载逻辑
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchTicker(symbol string, params map[string]interface{}) (*banexg.Ticker, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchTickers(symbols []string, params map[string]interface{}) ([]*banexg.Ticker, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchOrderBook(symbol string, limit int, params map[string]interface{}) (*banexg.OrderBook, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

// 交易接口
func (e *Longp) CreateOrder(symbol, odType, side string, amount, price float64, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) CancelOrder(id string, symbol string, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchOrder(symbol, orderId string, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchOrders(symbol string, since int64, limit int, params map[string]interface{}) ([]*banexg.Order, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchOpenOrders(symbol string, since int64, limit int, params map[string]interface{}) ([]*banexg.Order, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

// 账户接口
func (e *Longp) FetchBalance(params map[string]interface{}) (*banexg.Balances, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

func (e *Longp) FetchPositions(symbols []string, params map[string]interface{}) ([]*banexg.Position, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

// 杠杆接口
func (e *Longp) GetLeverage(symbol string, notional float64, account string) (float64, float64) {
	mar, exist := e.Markets[symbol]
	if !exist {
		return 0, 0
	}
	if mar.Type == banexg.MarketSpot {
		return 1, 1
	}
	// TODO: 实现具体的杠杆计算逻辑
	return 0, 0
}

func (e *Longp) SetLeverage(leverage float64, symbol string, params map[string]interface{}) (map[string]interface{}, *errs.Error) {
	return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
}

// 手续费计算
func makeCalcFee(e *Longp) banexg.FuncCalcFee {
	return func(market *banexg.Market, curr string, maker bool, amount, price decimal.Decimal, params map[string]interface{}) (*banexg.Fee, *errs.Error) {
		// TODO: 实现具体的手续费计算逻辑
		return nil, errs.NewMsg(errs.CodeNotImplement, "method not implement")
	}
}

// 关闭连接
func (e *Longp) Close() *errs.Error {
	// TODO: 实现关闭连接的逻辑
	return nil
}
