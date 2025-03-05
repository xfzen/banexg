package longp

import (
	"github.com/banbox/banexg"
	"github.com/banbox/banexg/errs"
)

type Longp struct {
	*banexg.Exchange
}

func New(Options map[string]interface{}) (*Longp, *errs.Error) {
	exg := &Longp{
		Exchange: createExchange(Options),
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

	e.ExgInfo.Min1mHole = Min1mHole

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
		HostPublic:  ApiBaseUrl,
		HostPrivate: ApiBaseUrl,
		HostWs:      WsBaseUrl,
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
