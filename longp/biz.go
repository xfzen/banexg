package longp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
	"github.com/longportapp/openapi-go/config"
	"github.com/longportapp/openapi-go/quote"
	"github.com/longportapp/openapi-go/trade"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

// UnmarshalResponse 解析API响应
func (e *Longp) UnmarshalResponse(resp *banexg.HttpRes, result interface{}) *errs.Error {
	if err := json.Unmarshal([]byte(resp.Content), result); err != nil {
		return errs.NewMsg(errs.CodeUnmarshalFail, "failed to unmarshal response: %v", err)
	}
	return nil
}

// 市场数据接口
func (e *Longp) LoadMarkets(reload bool, params map[string]interface{}) (banexg.MarketMap, *errs.Error) {
	if len(e.Markets) > 0 && !reload {
		return e.Markets, nil
	}

	// 初始化市场数据
	e.Markets = make(banexg.MarketMap)
	e.MarketsById = make(banexg.MarketArrMap)

	// 添加示例市场
	for market, exchange := range ExchangeMap {
		symbol := fmt.Sprintf("EXAMPLE.%s", market)
		marketInfo := &banexg.Market{
			ID:        symbol,
			Symbol:    symbol,
			Base:      "EXAMPLE",
			Quote:     market,
			Active:    true,
			Type:      banexg.MarketSpot,
			Spot:      true,
			Info:      map[string]interface{}{"exchange": exchange},
			Precision: &banexg.Precision{Price: 2, Amount: 2},
		}
		e.Markets[symbol] = marketInfo
		e.MarketsById[symbol] = []*banexg.Market{marketInfo}
	}

	return e.Markets, nil
}

func (e *Longp) FetchTicker(symbol string, params map[string]interface{}) (*banexg.Ticker, *errs.Error) {
	// 解析市场类型
	parts := strings.Split(symbol, ".")
	if len(parts) != 2 {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "invalid symbol format: %s", symbol)
	}

	market := parts[1]
	if _, ok := MarketTypeMap[market]; !ok {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "unsupported market: %s", market)
	}

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建行情上下文
	quoteContext, err := quote.NewFromCfg(conf)
	if err != nil {
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create quote context: %v", err)
	}
	defer quoteContext.Close()

	// 获取行情数据
	ctx := context.Background()
	quotes, err := quoteContext.Quote(ctx, []string{symbol})
	if err != nil {
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to fetch quote: %v", err)
	}

	if len(quotes) == 0 {
		return nil, errs.NewMsg(errs.CodeInvalidData, "no data for symbol: %s", symbol)
	}

	quote := quotes[0]
	return &banexg.Ticker{
		Symbol:     quote.Symbol,
		Last:       quote.LastDone.InexactFloat64(),
		Open:       quote.Open.InexactFloat64(),
		High:       quote.High.InexactFloat64(),
		Low:        quote.Low.InexactFloat64(),
		BaseVolume: float64(quote.Volume),
		TimeStamp:  quote.Timestamp,
	}, nil
}

func (e *Longp) FetchTickers(symbols []string, params map[string]interface{}) ([]*banexg.Ticker, *errs.Error) {
	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建行情上下文
	quoteContext, err := quote.NewFromCfg(conf)
	if err != nil {
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create quote context: %v", err)
	}
	defer quoteContext.Close()

	// 获取行情数据
	ctx := context.Background()
	quotes, err := quoteContext.Quote(ctx, symbols)
	if err != nil {
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to fetch quotes: %v", err)
	}

	tickers := make([]*banexg.Ticker, len(quotes))
	for i, quote := range quotes {
		tickers[i] = &banexg.Ticker{
			Symbol:     quote.Symbol,
			Last:       quote.LastDone.InexactFloat64(),
			Open:       quote.Open.InexactFloat64(),
			High:       quote.High.InexactFloat64(),
			Low:        quote.Low.InexactFloat64(),
			BaseVolume: float64(quote.Volume),
			TimeStamp:  quote.Timestamp,
		}
	}

	return tickers, nil
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
	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建账户上下文
	accountContext, err := trade.NewFromCfg(conf)
	if err != nil {
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create account context: %v", err)
	}
	defer accountContext.Close()

	// 获取账户资产
	ctx := context.Background()
	assets, err := accountContext.AccountBalance(ctx, &trade.GetAccountBalance{})
	if err != nil {
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to fetch account balance: %v", err)
	}

	// 转换资产信息
	balances := &banexg.Balances{
		TimeStamp: time.Now().UnixMilli(),
		Free:      make(map[string]float64),
		Used:      make(map[string]float64),
		Total:     make(map[string]float64),
		Assets:    make(map[string]*banexg.Asset),
	}

	for _, asset := range assets {
		currency := asset.Currency
		free := asset.TotalCash.InexactFloat64()
		used := decimal.Zero.InexactFloat64() // 暂时设置为0，因为API没有提供冻结金额
		total := asset.TotalCash.InexactFloat64()

		balances.Free[currency] = free
		balances.Used[currency] = used
		balances.Total[currency] = total

		balances.Assets[currency] = &banexg.Asset{
			Code:  currency,
			Free:  free,
			Used:  used,
			Total: total,
		}
	}

	return balances, nil
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

// FetchOHLCV 获取历史K线数据
func (e *Longp) FetchOHLCV(symbol string, timeframe string, since int64, limit int, params map[string]interface{}) ([]*banexg.Kline, *errs.Error) {
	// 打印请求参数
	logx.Infof("Fetching OHLCV for symbol: %s, timeframe: %s, since: %d, limit: %d", symbol, timeframe, since, limit)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建行情上下文
	quoteContext, err := quote.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create quote context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create quote context: %v", err)
	}
	defer quoteContext.Close()

	// 转换时间周期
	period, ok := TimeframeMap[timeframe]
	if !ok {
		logx.Errorf("Unsupported timeframe: %s", timeframe)
		return nil, errs.NewMsg(errs.CodeParamInvalid, "unsupported timeframe: %s", timeframe)
	}
	logx.Infof("Converted timeframe %s to period: %v", timeframe, period)

	// 获取K线数据
	ctx := context.Background()
	logx.Infof("Requesting candlesticks from LongPort API...")
	candles, err := quoteContext.Candlesticks(ctx, symbol, period, int32(limit), quote.AdjustTypeNo)
	if err != nil {
		logx.Errorf("Failed to fetch candlesticks: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to fetch candlesticks: %v", err)
	}
	logx.Infof("Received %d candlesticks from API", len(candles))

	// 转换K线数据
	klines := make([]*banexg.Kline, len(candles))
	for i, candle := range candles {
		klines[i] = &banexg.Kline{
			Time:   candle.Timestamp,
			Open:   candle.Open.InexactFloat64(),
			High:   candle.High.InexactFloat64(),
			Low:    candle.Low.InexactFloat64(),
			Close:  candle.Close.InexactFloat64(),
			Volume: float64(candle.Volume),
		}
	}

	// 打印转换后的K线数据
	prettyJSON, _ := json.MarshalIndent(klines, "", "  ")
	logx.Infof("Converted Klines:\n%s", string(prettyJSON))

	return klines, nil
}

// TimeframeMap 时间周期映射
var TimeframeMap = map[string]quote.Period{
	"1m":  quote.PeriodOneMinute,
	"5m":  quote.PeriodFiveMinute,
	"15m": quote.PeriodFifteenMinute,
	"30m": quote.PeriodThirtyMinute,
	"1d":  quote.PeriodDay,
	"1w":  quote.PeriodWeek,
	"1M":  quote.PeriodMonth,
	"1y":  quote.PeriodYear,
}
