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
	// 打印请求参数
	logx.Infof("Fetching order book for symbol: %s, limit: %d", symbol, limit)

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

	// 获取深度数据
	ctx := context.Background()
	depth, err := quoteContext.Depth(ctx, symbol)
	if err != nil {
		logx.Errorf("Failed to fetch depth: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to fetch depth: %v", err)
	}

	// 转换深度数据
	orderBook := &banexg.OrderBook{
		Symbol:    symbol,
		TimeStamp: time.Now().UnixMilli(),
		Limit:     limit,
		Bids:      &banexg.OdBookSide{IsBuy: true},
		Asks:      &banexg.OdBookSide{IsBuy: false},
	}

	// 转换买单
	for _, bid := range depth.Bid {
		orderBook.Bids.Price = append(orderBook.Bids.Price, bid.Price.InexactFloat64())
		orderBook.Bids.Size = append(orderBook.Bids.Size, float64(bid.Volume))
	}

	// 转换卖单
	for _, ask := range depth.Ask {
		orderBook.Asks.Price = append(orderBook.Asks.Price, ask.Price.InexactFloat64())
		orderBook.Asks.Size = append(orderBook.Asks.Size, float64(ask.Volume))
	}

	// 打印转换后的数据
	prettyJSON, _ := json.MarshalIndent(orderBook, "", "  ")
	logx.Infof("Order book data:\n%s", string(prettyJSON))

	return orderBook, nil
}

// 交易接口
func (e *Longp) CreateOrder(symbol, odType, side string, amount, price float64, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	// 打印请求参数
	logx.Infof("Creating order: symbol=%s, type=%s, side=%s, amount=%.2f, price=%.2f",
		symbol, odType, side, amount, price)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 转换订单类型
	var orderType trade.OrderType
	switch odType {
	case "limit":
		orderType = trade.OrderTypeLO
	case "market":
		orderType = trade.OrderTypeMO
	default:
		logx.Errorf("Unsupported order type: %s", odType)
		return nil, errs.NewMsg(errs.CodeParamInvalid, "unsupported order type: %s", odType)
	}

	// 转换订单方向
	var orderSide trade.OrderSide
	switch side {
	case "buy":
		orderSide = trade.OrderSideBuy
	case "sell":
		orderSide = trade.OrderSideSell
	default:
		logx.Errorf("Unsupported order side: %s", side)
		return nil, errs.NewMsg(errs.CodeParamInvalid, "unsupported order side: %s", side)
	}

	// 创建订单请求
	req := &trade.SubmitOrder{
		Symbol:            symbol,
		OrderType:         orderType,
		Side:              orderSide,
		SubmittedQuantity: uint64(amount),
		SubmittedPrice:    decimal.NewFromFloat(price),
		TimeInForce:       trade.TimeTypeDay,
	}

	// 提交订单
	ctx := context.Background()
	orderID, err := tradeContext.SubmitOrder(ctx, req)
	if err != nil {
		logx.Errorf("Failed to submit order: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to submit order: %v", err)
	}

	// 转换订单数据
	result := &banexg.Order{
		ID:        orderID,
		Symbol:    symbol,
		Type:      odType,
		Side:      side,
		Price:     price,
		Amount:    amount,
		Status:    "open", // 默认状态
		Timestamp: time.Now().UnixMilli(),
	}

	// 打印订单数据
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	logx.Infof("Created order:\n%s", string(prettyJSON))

	return result, nil
}

func (e *Longp) CancelOrder(id string, symbol string, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	// 打印请求参数
	logx.Infof("Cancelling order: id=%s, symbol=%s", id, symbol)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 取消订单
	ctx := context.Background()
	err = tradeContext.CancelOrder(ctx, id)
	if err != nil {
		logx.Errorf("Failed to cancel order: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to cancel order: %v", err)
	}

	// 获取订单详情
	order, err := tradeContext.OrderDetail(ctx, id)
	if err != nil {
		logx.Errorf("Failed to get order details: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to get order details: %v", err)
	}

	// 转换订单数据
	price := 0.0
	if order.Price != nil {
		price, _ = order.Price.Float64()
	}

	amount := float64(order.Quantity)
	executedAmount := float64(order.ExecutedQuantity)

	executedPrice := 0.0
	if order.ExecutedPrice != nil {
		executedPrice, _ = order.ExecutedPrice.Float64()
	}

	// 转换时间戳
	timestamp, _ := time.Parse(time.RFC3339, order.SubmittedAt)

	orderData := &banexg.Order{
		ID:        order.OrderId,
		Symbol:    order.Symbol,
		Type:      string(order.OrderType),
		Side:      string(order.Side),
		Price:     price,
		Amount:    amount,
		Status:    string(order.Status),
		Timestamp: timestamp.UnixMilli(),
		Filled:    executedAmount,
		Average:   executedPrice,
	}

	// 打印订单数据
	prettyJSON, _ := json.MarshalIndent(orderData, "", "  ")
	logx.Infof("Cancelled order:\n%s", string(prettyJSON))

	return orderData, nil
}

func (e *Longp) FetchOrder(symbol, orderId string, params map[string]interface{}) (*banexg.Order, *errs.Error) {
	// 打印请求参数
	logx.Infof("Fetching order: symbol=%s, orderId=%s", symbol, orderId)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 获取订单详情
	ctx := context.Background()
	order, err := tradeContext.OrderDetail(ctx, orderId)
	if err != nil {
		logx.Errorf("Failed to get order details: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to get order details: %v", err)
	}

	// 转换订单数据
	price := 0.0
	if order.Price != nil {
		price, _ = order.Price.Float64()
	}

	amount := float64(order.Quantity)
	executedAmount := float64(order.ExecutedQuantity)

	executedPrice := 0.0
	if order.ExecutedPrice != nil {
		executedPrice, _ = order.ExecutedPrice.Float64()
	}

	// 转换时间戳
	timestamp, _ := time.Parse(time.RFC3339, order.SubmittedAt)

	orderData := &banexg.Order{
		ID:        order.OrderId,
		Symbol:    order.Symbol,
		Type:      string(order.OrderType),
		Side:      string(order.Side),
		Price:     price,
		Amount:    amount,
		Status:    string(order.Status),
		Timestamp: timestamp.UnixMilli(),
		Filled:    executedAmount,
		Average:   executedPrice,
	}

	// 打印订单数据
	prettyJSON, _ := json.MarshalIndent(orderData, "", "  ")
	logx.Infof("Order details:\n%s", string(prettyJSON))

	return orderData, nil
}

func (e *Longp) FetchOrders(symbol string, since int64, limit int, params map[string]interface{}) ([]*banexg.Order, *errs.Error) {
	// 打印请求参数
	logx.Infof("Fetching orders: symbol=%s, since=%d, limit=%d", symbol, since, limit)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 获取订单列表
	ctx := context.Background()
	orders, _, err := tradeContext.HistoryOrders(ctx, &trade.GetHistoryOrders{
		Symbol:  symbol,
		StartAt: since,
	})
	if err != nil {
		logx.Errorf("Failed to get order history: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to get order history: %v", err)
	}

	// 转换订单数据
	result := make([]*banexg.Order, len(orders))
	for i, order := range orders {
		price := 0.0
		if order.Price != nil {
			price, _ = order.Price.Float64()
		}

		amount, _ := decimal.NewFromString(order.Quantity)
		executedAmount, _ := decimal.NewFromString(order.ExecutedQuantity)

		executedPrice := 0.0
		if order.ExecutedPrice != nil {
			executedPrice, _ = order.ExecutedPrice.Float64()
		}

		// 转换时间戳
		timestamp, _ := time.Parse(time.RFC3339, order.SubmittedAt)

		result[i] = &banexg.Order{
			ID:        order.OrderId,
			Symbol:    order.Symbol,
			Type:      string(order.OrderType),
			Side:      string(order.Side),
			Price:     price,
			Amount:    amount.InexactFloat64(),
			Status:    string(order.Status),
			Timestamp: timestamp.UnixMilli(),
			Filled:    executedAmount.InexactFloat64(),
			Average:   executedPrice,
		}
	}

	// 打印订单数据
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	logx.Infof("Order history:\n%s", string(prettyJSON))

	return result, nil
}

func (e *Longp) FetchOpenOrders(symbol string, since int64, limit int, params map[string]interface{}) ([]*banexg.Order, *errs.Error) {
	// 打印请求参数
	logx.Infof("Fetching open orders: symbol=%s, since=%d, limit=%d", symbol, since, limit)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 获取未完成订单列表
	ctx := context.Background()
	orders, err := tradeContext.TodayOrders(ctx, &trade.GetTodayOrders{
		Symbol: symbol,
	})
	if err != nil {
		logx.Errorf("Failed to get today's orders: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to get today's orders: %v", err)
	}

	// 转换订单数据
	var result []*banexg.Order
	for _, order := range orders {
		// 只保留未完成的订单
		if order.Status != "Filled" && order.Status != "Cancelled" {
			price := 0.0
			if order.Price != nil {
				price, _ = order.Price.Float64()
			}

			amount, _ := decimal.NewFromString(order.Quantity)
			executedAmount, _ := decimal.NewFromString(order.ExecutedQuantity)

			executedPrice := 0.0
			if order.ExecutedPrice != nil {
				executedPrice, _ = order.ExecutedPrice.Float64()
			}

			// 转换时间戳
			timestamp, _ := time.Parse(time.RFC3339, order.SubmittedAt)

			result = append(result, &banexg.Order{
				ID:        order.OrderId,
				Symbol:    order.Symbol,
				Type:      string(order.OrderType),
				Side:      string(order.Side),
				Price:     price,
				Amount:    amount.InexactFloat64(),
				Status:    string(order.Status),
				Timestamp: timestamp.UnixMilli(),
				Filled:    executedAmount.InexactFloat64(),
				Average:   executedPrice,
			})
		}
	}

	// 打印订单数据
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	logx.Infof("Open orders:\n%s", string(prettyJSON))

	return result, nil
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
	// 打印请求参数
	logx.Infof("Fetching positions for symbols: %v", symbols)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 获取持仓信息
	ctx := context.Background()
	positionChannels, err := tradeContext.StockPositions(ctx, symbols)
	if err != nil {
		logx.Errorf("Failed to get stock positions: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to get stock positions: %v", err)
	}

	// 转换持仓数据
	var result []*banexg.Position
	for _, channel := range positionChannels {
		for _, pos := range channel.Positions {
			// 获取当前价格
			price := 0.0
			if pos.CostPrice != nil {
				price, _ = pos.CostPrice.Float64()
			}

			// 获取持仓数量
			quantity, _ := decimal.NewFromString(pos.Quantity)
			availableQuantity, _ := decimal.NewFromString(pos.AvailableQuantity)

			// 计算持仓市值
			marketValue := quantity.Mul(decimal.NewFromFloat(price))

			position := &banexg.Position{
				ID:               pos.Symbol,
				Symbol:           pos.Symbol,
				TimeStamp:        time.Now().UnixMilli(),
				Isolated:         false,
				Hedged:           false,
				Side:             "long",
				Contracts:        quantity.InexactFloat64(),
				ContractSize:     1,
				EntryPrice:       price,
				MarkPrice:        price,
				Notional:         marketValue.InexactFloat64(),
				Leverage:         1,
				Collateral:       marketValue.InexactFloat64(),
				InitialMargin:    marketValue.InexactFloat64(),
				MaintMargin:      marketValue.InexactFloat64(),
				InitialMarginPct: 1.0,
				MaintMarginPct:   1.0,
				UnrealizedPnl:    0,
				LiquidationPrice: 0,
				MarginMode:       "cross",
				MarginRatio:      1.0,
				Percentage:       0,
				Info: map[string]interface{}{
					"quantity":           quantity.InexactFloat64(),
					"available_quantity": availableQuantity.InexactFloat64(),
					"cost_price":         price,
					"market_value":       marketValue.InexactFloat64(),
					"currency":           pos.Currency,
					"market":             pos.Market,
					"name":               pos.SymbolName,
				},
			}

			result = append(result, position)
		}
	}

	// 打印持仓数据
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	logx.Infof("Positions:\n%s", string(prettyJSON))

	return result, nil
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
	// 打印请求参数
	logx.Infof("Setting leverage: symbol=%s, leverage=%.2f", symbol, leverage)

	// 创建配置
	conf := &config.Config{
		AppKey:      e.Options["apiKey"].(string),
		AppSecret:   e.Options["secret"].(string),
		AccessToken: e.Options["accessToken"].(string),
	}

	// 创建交易上下文
	tradeContext, err := trade.NewFromCfg(conf)
	if err != nil {
		logx.Errorf("Failed to create trade context: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to create trade context: %v", err)
	}
	defer tradeContext.Close()

	// 获取当前保证金率
	ctx := context.Background()
	marginRatio, err := tradeContext.MarginRatio(ctx, symbol)
	if err != nil {
		logx.Errorf("Failed to get margin ratio: %v", err)
		return nil, errs.NewMsg(errs.CodeRunTime, "failed to get margin ratio: %v", err)
	}

	// 转换保证金率
	initialMarginRatio := 0.0
	if marginRatio.ImFactor != nil {
		initialMarginRatio, _ = marginRatio.ImFactor.Float64()
	}

	maintenanceMarginRatio := 0.0
	if marginRatio.MmFactor != nil {
		maintenanceMarginRatio, _ = marginRatio.MmFactor.Float64()
	}

	forcedCloseRatio := 0.0
	if marginRatio.FmFactor != nil {
		forcedCloseRatio, _ = marginRatio.FmFactor.Float64()
	}

	// 返回结果
	result := map[string]interface{}{
		"leverage": leverage,
		"margin_ratio": map[string]interface{}{
			"initial":     initialMarginRatio,
			"maintenance": maintenanceMarginRatio,
			"forced":      forcedCloseRatio,
		},
	}

	// 打印结果
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	logx.Infof("Set leverage result:\n%s", string(prettyJSON))

	return result, nil
}

// 手续费计算
func makeCalcFee(e *Longp) banexg.FuncCalcFee {
	return func(market *banexg.Market, curr string, maker bool, amount, price decimal.Decimal, params map[string]interface{}) (*banexg.Fee, *errs.Error) {
		// 计算交易金额
		cost := amount.Mul(price)

		// 获取手续费率
		var feeRate decimal.Decimal
		if maker {
			feeRate = decimal.NewFromFloat(0.0001) // 挂单费率 0.01%
		} else {
			feeRate = decimal.NewFromFloat(0.0002) // 吃单费率 0.02%
		}

		// 计算手续费
		fee := cost.Mul(feeRate)

		// 返回手续费信息
		return &banexg.Fee{
			Currency: curr,
			Cost:     fee.InexactFloat64(),
			Rate:     feeRate.InexactFloat64(),
		}, nil
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
