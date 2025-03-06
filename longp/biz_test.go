package longp

import (
	"encoding/json"
	"testing"

	"github.com/banbox/banexg/bex"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
)

var testOptions map[string]interface{}

func init() {
	logx.SetUp(logx.LogConf{
		Encoding: "plain",
	})

	// 初始化测试配置
	testOptions = map[string]interface{}{
		"apiKey":      "c0759bd8a76fec167e14378e620a846f",
		"secret":      "e500d6af1a16ff975a54f03299d68b1d13cbb1470fa725dd54ad041c8186dfff",
		"accessToken": "m_eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJsb25nYnJpZGdlIiwic3ViIjoiYWNjZXNzX3Rva2VuIiwiZXhwIjoxNzQ4OTQwMjk3LCJpYXQiOjE3NDExNjQyOTcsImFrIjoiYzA3NTliZDhhNzZmZWMxNjdlMTQzNzhlNjIwYTg0NmYiLCJhYWlkIjoyMDIzODgxMCwiYWMiOiJsYiIsIm1pZCI6MTEyNDI0MTEsInNpZCI6IlgrcGQ3Ky94ZFlCeUgxTXhTd2dzcUE9PSIsImJsIjozLCJ1bCI6MCwiaWsiOiJsYl8yMDIzODgxMCJ9.egw0xkiqF9F9MuTfdZK78_6iHRnt98j4KQyL5lXQT6__xTMx7Jqt9ifCmJfpdXjJ9cPJD542GdnD-39VyJsnrpU_GAg8ii1AGsscj_BMZokOLtgtwY9wnZo5cxzcX7_J9NqwJeBG4ePHxvXlqqJ5CM8lPUe1-68gETaO76-Vukr1V28l4a6cVPiEIItWzflbF-PXQY3GKg9_l1qo56SG3jizQ-FNMxG-ZNwC81d-K22tF0kMwQiMQJdvgv9qtuNuYKozS6VCUMicVc4Aekcb9tlxIeZD7S4qRCfMpaliIFS5VPRXqhnnOFeviG-FyxFpri9onNoRdJIGlqOzXEqL6RNIoY3Z8lyM_n1ezApcMVn9JUNMbLd7K85wnSOJ1fFvXV78mXZJiM4d1too3zzq5mW0TfLm17L4an2CrkmprHsPkwfvfsUl_kZmpforHaQZN0pxjM-z8hZ0ETjaoxXXpxGntqgJv_PEnGYBixZzKViQcFVQB83roSDBAWV4Vgto_k2H4IZT-cSX7PO_PzeZRxcCdUXyL5-yy1En996oRZ51QzQQvby0RWv7ycClJqEPiyg9FIBTZ9TPrGNlloMlmvTTf75fmpTwAPefinng8lIKmUlWuCacjOvL7NYDd82bu1XC2hDccwfyQhFUrM7ZUYSv6wOju_9VZwObS5qN5nY",
	}
}

func createTestExchange(t *testing.T) *Longp {
	exg, err := New(testOptions)
	if err != nil {
		logx.Errorf("Failed to create exchange: %v", err)
		t.Fatalf("Failed to create exchange: %v", err)
	}
	return exg
}

func TestLoadMarkets(t *testing.T) {
	exg := createTestExchange(t)
	defer exg.Close()

	markets, err := exg.LoadMarkets(true, nil)
	if err != nil {
		t.Fatalf("Failed to load markets: %v", err)
	}

	logx.Infof("Loaded %d markets", len(markets))
	for symbol, market := range markets {
		logx.Infof("Market: %s, Type: %s, Base: %s, Quote: %s",
			symbol, market.Type, market.Base, market.Quote)
	}

	// 验证市场数量
	expectedMarkets := len(ExchangeMap)
	if len(markets) != expectedMarkets {
		t.Errorf("Expected %d markets, got %d", expectedMarkets, len(markets))
	}

	// 验证示例市场
	for market := range ExchangeMap {
		symbol := "EXAMPLE." + market
		if _, ok := markets[symbol]; !ok {
			t.Errorf("Market %s not found", symbol)
		}
	}
}

func TestFetchTicker(t *testing.T) {
	exg := createTestExchange(t)
	defer exg.Close()

	// 测试有效市场
	symbol := "700.HK"
	ticker, err := exg.FetchTicker(symbol, nil)
	if err != nil {
		t.Fatalf("Failed to fetch ticker for %s: %v", symbol, err)
	}

	logx.Infof("Ticker for %s: Last=%.2f, Open=%.2f, High=%.2f, Low=%.2f, Volume=%.2f",
		ticker.Symbol, ticker.Last, ticker.Open, ticker.High, ticker.Low, ticker.BaseVolume)

	// 验证ticker字段
	if ticker.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, ticker.Symbol)
	}

	// 测试无效市场
	invalidSymbol := "INVALID"
	_, err = exg.FetchTicker(invalidSymbol, nil)
	if err == nil {
		t.Error("Expected error for invalid symbol, got nil")
	}
}

func TestFetchTickers(t *testing.T) {
	exg := createTestExchange(t)
	defer exg.Close()

	symbols := []string{"700.HK", "AAPL.US", "TSLA.US", "NFLX.US"}
	tickers, err := exg.FetchTickers(symbols, nil)
	if err != nil {
		t.Fatalf("Failed to fetch tickers: %v", err)
	}

	logx.Infof("Fetched %d tickers", len(tickers))
	for _, ticker := range tickers {
		logx.Infof("Ticker: %s, Last=%.2f, Open=%.2f, High=%.2f, Low=%.2f, Volume=%.2f",
			ticker.Symbol, ticker.Last, ticker.Open, ticker.High, ticker.Low, ticker.BaseVolume)
	}

	// 验证tickers数量
	if len(tickers) != len(symbols) {
		t.Errorf("Expected %d tickers, got %d", len(symbols), len(tickers))
	}

	// 验证每个ticker
	for i, ticker := range tickers {
		if ticker.Symbol != symbols[i] {
			t.Errorf("Expected symbol %s, got %s", symbols[i], ticker.Symbol)
		}
	}
}

func TestBexIntegration(t *testing.T) {
	exg, err := bex.New("longp", testOptions)
	if err != nil {
		t.Fatalf("Failed to create exchange: %v", err)
	}
	defer exg.Close()

	// 获取市场数据
	symbols := []string{"700.HK", "AAPL.US", "TSLA.US", "NFLX.US"}

	// 测试单个ticker
	ticker, err := exg.FetchTicker(symbols[0], nil)
	if err != nil {
		t.Fatalf("Failed to fetch ticker: %v", err)
	}
	logx.Infof("BEX Ticker for %s: Last=%.2f, Open=%.2f, High=%.2f, Low=%.2f, Volume=%.2f",
		ticker.Symbol, ticker.Last, ticker.Open, ticker.High, ticker.Low, ticker.BaseVolume)

	if ticker.Symbol != symbols[0] {
		t.Errorf("Expected symbol %s, got %s", symbols[0], ticker.Symbol)
	}

	// 测试多个tickers
	tickers, err := exg.FetchTickers(symbols, nil)
	if err != nil {
		t.Fatalf("Failed to fetch tickers: %v", err)
	}
	logx.Infof("BEX Fetched %d tickers", len(tickers))
	for _, ticker := range tickers {
		logx.Infof("BEX Ticker: %s, Last=%.2f, Open=%.2f, High=%.2f, Low=%.2f, Volume=%.2f",
			ticker.Symbol, ticker.Last, ticker.Open, ticker.High, ticker.Low, ticker.BaseVolume)
	}

	if len(tickers) != len(symbols) {
		t.Errorf("Expected %d tickers, got %d", len(symbols), len(tickers))
	}
}

func TestFetchBalance(t *testing.T) {
	exg := createTestExchange(t)
	defer exg.Close()

	// 获取账户余额
	balances, err := exg.FetchBalance(nil)
	if err != nil {
		t.Fatalf("Failed to fetch balance: %v", err)
	}

	// 验证返回的余额信息
	if balances == nil {
		t.Fatal("Balances should not be nil")
	}

	// 验证时间戳
	if balances.TimeStamp <= 0 {
		t.Error("Timestamp should be positive")
	}

	// 验证资产信息
	if len(balances.Assets) == 0 {
		t.Error("Should have at least one asset")
	}

	// 打印账户信息
	logx.Infof("Account balance timestamp: %d", balances.TimeStamp)
	for currency, asset := range balances.Assets {
		logx.Infof("Currency: %s, Free: %f, Used: %f, Total: %f",
			currency, asset.Free, asset.Used, asset.Total)

		// 验证资产字段
		if asset.Code != currency {
			t.Errorf("Asset code mismatch: expected %s, got %s", currency, asset.Code)
		}
		if asset.Free < 0 {
			t.Errorf("Free balance should not be negative for %s", currency)
		}
		if asset.Used < 0 {
			t.Errorf("Used balance should not be negative for %s", currency)
		}
		if asset.Total < 0 {
			t.Errorf("Total balance should not be negative for %s", currency)
		}
		if asset.Total != asset.Free+asset.Used {
			t.Errorf("Total balance should equal Free + Used for %s", currency)
		}

		// 验证余额映射
		if balances.Free[currency] != asset.Free {
			t.Errorf("Free balance mismatch for %s", currency)
		}
		if balances.Used[currency] != asset.Used {
			t.Errorf("Used balance mismatch for %s", currency)
		}
		if balances.Total[currency] != asset.Total {
			t.Errorf("Total balance mismatch for %s", currency)
		}
	}
}

func TestFetchOHLCV(t *testing.T) {
	exg := createTestExchange(t)
	defer exg.Close()

	// 测试获取日K线
	symbol := "700.HK"
	timeframe := "1d"
	limit := 10

	logx.Infof("Starting OHLCV test for symbol: %s, timeframe: %s, limit: %d", symbol, timeframe, limit)

	klines, err := exg.FetchOHLCV(symbol, timeframe, 0, limit, nil)
	if err != nil {
		logx.Errorf("Failed to fetch OHLCV: %v", err)
		t.Fatalf("Failed to fetch OHLCV: %v", err)
	}

	// 验证返回的K线数据
	if len(klines) == 0 {
		logx.Error("No Klines returned")
		t.Fatal("Should have at least one Kline")
	}

	// 打印K线数据
	logx.Infof("Successfully fetched %d Kline records for %s", len(klines), symbol)

	// 使用 pretty JSON 格式打印数据
	prettyJSON, _ := json.MarshalIndent(klines, "", "  ")
	logx.Infof("Kline data:\n%s", string(prettyJSON))

	// 验证数据完整性
	for i, kline := range klines {
		logx.Infof("Validating Kline at index %d: Time=%d, Open=%.2f, High=%.2f, Low=%.2f, Close=%.2f, Volume=%.2f",
			i, kline.Time, kline.Open, kline.High, kline.Low, kline.Close, kline.Volume)

		if kline.Time <= 0 {
			logx.Errorf("Invalid timestamp at index %d: %d", i, kline.Time)
			t.Errorf("Invalid timestamp at index %d", i)
		}
		if kline.Open <= 0 {
			logx.Errorf("Invalid open price at index %d: %.2f", i, kline.Open)
			t.Errorf("Invalid open price at index %d", i)
		}
		if kline.High <= 0 {
			logx.Errorf("Invalid high price at index %d: %.2f", i, kline.High)
			t.Errorf("Invalid high price at index %d", i)
		}
		if kline.Low <= 0 {
			logx.Errorf("Invalid low price at index %d: %.2f", i, kline.Low)
			t.Errorf("Invalid low price at index %d", i)
		}
		if kline.Close <= 0 {
			logx.Errorf("Invalid close price at index %d: %.2f", i, kline.Close)
			t.Errorf("Invalid close price at index %d", i)
		}
		if kline.Volume < 0 {
			logx.Errorf("Invalid volume at index %d: %.2f", i, kline.Volume)
			t.Errorf("Invalid volume at index %d", i)
		}
	}

	logx.Info("OHLCV test completed successfully")
}

func TestFetchOrderBook(t *testing.T) {
	exg := createTestExchange(t)
	defer exg.Close()

	// 测试获取订单簿
	symbol := "700.HK"
	limit := 10

	logx.Infof("Starting order book test for symbol: %s, limit: %d", symbol, limit)

	orderBook, err := exg.FetchOrderBook(symbol, limit, nil)
	if err != nil {
		logx.Errorf("Failed to fetch order book: %v", err)
		t.Fatalf("Failed to fetch order book: %v", err)
	}

	// 验证返回的订单簿数据
	if orderBook == nil {
		logx.Error("Order book should not be nil")
		t.Fatal("Order book should not be nil")
	}

	// 验证基本字段
	if orderBook.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, orderBook.Symbol)
	}
	if orderBook.TimeStamp <= 0 {
		t.Error("Timestamp should be positive")
	}
	if orderBook.Limit != limit {
		t.Errorf("Expected limit %d, got %d", limit, orderBook.Limit)
	}

	// 验证买卖盘
	if orderBook.Bids == nil {
		t.Error("Bids should not be nil")
	}
	if orderBook.Asks == nil {
		t.Error("Asks should not be nil")
	}

	// 验证买卖盘数据
	if len(orderBook.Bids.Price) != len(orderBook.Bids.Size) {
		t.Error("Bids price and size arrays should have the same length")
	}
	if len(orderBook.Asks.Price) != len(orderBook.Asks.Size) {
		t.Error("Asks price and size arrays should have the same length")
	}

	// 打印订单簿数据
	logx.Infof("Order book for %s:", symbol)
	logx.Infof("Bids: %d levels", len(orderBook.Bids.Price))
	for i := 0; i < len(orderBook.Bids.Price); i++ {
		logx.Infof("  Price: %.2f, Size: %.2f", orderBook.Bids.Price[i], orderBook.Bids.Size[i])
	}
	logx.Infof("Asks: %d levels", len(orderBook.Asks.Price))
	for i := 0; i < len(orderBook.Asks.Price); i++ {
		logx.Infof("  Price: %.2f, Size: %.2f", orderBook.Asks.Price[i], orderBook.Asks.Size[i])
	}

	// 验证价格排序
	for i := 1; i < len(orderBook.Bids.Price); i++ {
		if orderBook.Bids.Price[i] > orderBook.Bids.Price[i-1] {
			t.Errorf("Bids should be in descending order, got %.2f > %.2f",
				orderBook.Bids.Price[i], orderBook.Bids.Price[i-1])
		}
	}
	for i := 1; i < len(orderBook.Asks.Price); i++ {
		if orderBook.Asks.Price[i] < orderBook.Asks.Price[i-1] {
			t.Errorf("Asks should be in ascending order, got %.2f < %.2f",
				orderBook.Asks.Price[i], orderBook.Asks.Price[i-1])
		}
	}

	logx.Info("Order book test completed successfully")
}

func TestCancelOrder(t *testing.T) {
	// 创建测试实例
	e := createTestExchange(t)
	defer e.Close()

	// 创建订单
	order, err := e.CreateOrder("700.HK", "limit", "buy", 100, 100.0, nil)
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}

	// 取消订单
	cancelledOrder, err := e.CancelOrder(order.ID, order.Symbol, nil)
	if err != nil {
		t.Fatalf("取消订单失败: %v", err)
	}

	// 验证取消的订单数据
	if cancelledOrder == nil {
		t.Fatal("取消的订单数据为空")
	}

	// 验证基本字段
	if cancelledOrder.ID != order.ID {
		t.Errorf("订单ID不匹配: 期望=%s, 实际=%s", order.ID, cancelledOrder.ID)
	}
	if cancelledOrder.Symbol != order.Symbol {
		t.Errorf("交易对不匹配: 期望=%s, 实际=%s", order.Symbol, cancelledOrder.Symbol)
	}
	if cancelledOrder.Type != order.Type {
		t.Errorf("订单类型不匹配: 期望=%s, 实际=%s", order.Type, cancelledOrder.Type)
	}
	if cancelledOrder.Side != order.Side {
		t.Errorf("订单方向不匹配: 期望=%s, 实际=%s", order.Side, cancelledOrder.Side)
	}
	if cancelledOrder.Price != order.Price {
		t.Errorf("订单价格不匹配: 期望=%.2f, 实际=%.2f", order.Price, cancelledOrder.Price)
	}
	if cancelledOrder.Amount != order.Amount {
		t.Errorf("订单数量不匹配: 期望=%.2f, 实际=%.2f", order.Amount, cancelledOrder.Amount)
	}

	// 验证状态
	if cancelledOrder.Status != "canceled" {
		t.Errorf("订单状态不正确: 期望=canceled, 实际=%s", cancelledOrder.Status)
	}

	// 验证时间戳
	if cancelledOrder.Timestamp <= 0 {
		t.Error("订单时间戳无效")
	}

	// 打印取消的订单数据
	prettyJSON, _ := json.MarshalIndent(cancelledOrder, "", "  ")
	t.Logf("取消的订单数据:\n%s", string(prettyJSON))
}

func TestGetLeverage(t *testing.T) {
	// 创建测试实例
	exchange := createTestExchange(t)
	defer exchange.Close()

	// 加载市场数据
	markets, err := exchange.LoadMarkets(true, nil)
	if err != nil {
		t.Fatalf("Failed to load markets: %v", err)
	}

	// 验证市场数据已加载
	if len(markets) == 0 {
		t.Fatal("No markets loaded")
	}

	// 测试现货市场
	t.Run("Spot Market", func(t *testing.T) {
		// 确保市场存在
		if _, ok := markets["700.HK"]; !ok {
			t.Skip("Market 700.HK not found in test markets")
		}
		leverage, maxLeverage := exchange.GetLeverage("700.HK", 1000, "")
		assert.Equal(t, 1.0, leverage)
		assert.Equal(t, 1.0, maxLeverage)
	})

	// 测试不存在的市场
	t.Run("Non-existent Market", func(t *testing.T) {
		leverage, maxLeverage := exchange.GetLeverage("NONEXISTENT", 1000, "")
		assert.Equal(t, 0.0, leverage)
		assert.Equal(t, 0.0, maxLeverage)
	})

	// 测试合约市场
	t.Run("Contract Market", func(t *testing.T) {
		// 由于目前 GetLeverage 方法对合约市场返回 0,0
		// 这里我们验证这个行为
		leverage, maxLeverage := exchange.GetLeverage("BTC-USDT", 1000, "")
		assert.Equal(t, 0.0, leverage)
		assert.Equal(t, 0.0, maxLeverage)
	})
}
