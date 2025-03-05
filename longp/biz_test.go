package longp

import (
	"encoding/json"
	"testing"

	"github.com/banbox/banexg/bex"
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
