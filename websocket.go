package banexg

import (
	"github.com/anyongjin/banexg/errs"
	"github.com/anyongjin/banexg/log"
	"github.com/anyongjin/banexg/utils"
	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type WsClient struct {
	Conn       WsConn
	URL        string
	MarketType string
	Send       chan []byte
	control    chan int              // 用于内部同步控制命令
	JobInfos   map[string]*WsJobInfo // request id: Sub Data
	ChanCaps   map[string]int        // msgHash: cap size of cache msg
	OnMessage  FuncOnWsMsg
	OnError    FuncOnWsErr
	OnClose    FuncOnWsClose
}

type WebSocket struct {
	Conn *websocket.Conn
}

func (ws *WebSocket) Close() error {
	return ws.Conn.Close()
}
func (ws *WebSocket) WriteClose() error {
	exitData := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	return ws.Conn.WriteMessage(websocket.CloseMessage, exitData)
}
func (ws *WebSocket) NextWriter() (io.WriteCloser, error) {
	return ws.Conn.NextWriter(websocket.TextMessage)
}
func (ws *WebSocket) ReadMsg() ([]byte, error) {
	for {
		msgType, msgRaw, err := ws.Conn.ReadMessage()
		if err != nil || msgType == websocket.TextMessage {
			return msgRaw, err
		}
	}
}

func newWebSocket(reqUrl string, args map[string]interface{}) (*WebSocket, error) {
	var dialer websocket.Dialer
	dialer.HandshakeTimeout = utils.GetMapVal(args, ParamHandshakeTimeout, time.Second*15)
	var defProxy *url.URL
	var proxy = utils.GetMapVal(args, ParamProxy, defProxy)
	if proxy != nil {
		dialer.Proxy = http.ProxyURL(proxy)
	}
	conn, _, err := dialer.Dial(reqUrl, http.Header{})
	if err != nil {
		return nil, errs.New(errs.CodeConnectFail, err)
	}
	return &WebSocket{Conn: conn}, nil
}

var (
	ParamHandshakeTimeout = "HandshakeTimeout"
	ParamChanCaps         = "ChanCaps"
	ParamChanCap          = "ChanCap"
)

const (
	ctrlDoClose = iota
	ctrlClosed
)

var (
	DefChanCaps = map[string]int{
		"@depth": 1000,
	}
)

func newWsClient(reqUrl string, onMsg FuncOnWsMsg, onErr FuncOnWsErr, onClose FuncOnWsClose,
	params *map[string]interface{}) (*WsClient, *errs.Error) {
	var result = &WsClient{
		URL:       reqUrl,
		Send:      make(chan []byte, 1024),
		JobInfos:  make(map[string]*WsJobInfo),
		OnMessage: onMsg,
		OnError:   onErr,
		OnClose:   onClose,
		control:   make(chan int, 1),
	}
	args := utils.SafeParams(params)
	result.ChanCaps = DefChanCaps
	chanCaps := utils.GetMapVal(args, ParamChanCaps, map[string]int{})
	for k, v := range chanCaps {
		result.ChanCaps[k] = v
	}
	var conn WsConn
	conn = utils.GetMapVal(args, OptWsConn, conn)
	if conn == nil {
		var err error
		conn, err = newWebSocket(reqUrl, args)
		if err != nil {
			return nil, errs.New(errs.CodeConnectFail, err)
		}
	}
	result.Conn = conn
	go result.read()
	go result.write()
	return result, nil
}

func (e *Exchange) GetClient(wsUrl string, marketType string) (*WsClient, *errs.Error) {
	client, ok := e.WSClients[wsUrl]
	if ok && client.Conn != nil {
		return client, nil
	}
	params := map[string]interface{}{}
	if e.Proxy != nil {
		params[ParamProxy] = e.Proxy
	}
	if conn, ok := e.Options[OptWsConn]; ok {
		params[OptWsConn] = conn
	}
	if e.OnWsMsg == nil {
		return nil, errs.NewMsg(errs.CodeParamInvalid, "OnWsMsg is required for ws client")
	}
	onClosed := func(wsUrl string, err *errs.Error) {
		if e.OnWsClose != nil {
			e.OnWsClose(wsUrl, err)
		}
		num := e.handleWsClientClosed(client)
		log.Info("closed out chan for ws client", zap.Int("num", num))
	}
	client, err := newWsClient(wsUrl, e.OnWsMsg, e.OnWsErr, onClosed, &params)
	if err != nil {
		return nil, err
	}
	client.MarketType = marketType
	e.WSClients[wsUrl] = client
	return client, nil
}

/*
GetWsOutChan
获取指定msgHash的输出通道
如果不存在则创建新的并存储
*/
func GetWsOutChan[T any](e *Exchange, chanKey string, create func(int) T, args map[string]interface{}) T {
	outRaw, oldChan := e.WsOutChans[chanKey]
	if oldChan {
		res := outRaw.(T)
		return res
	} else {
		chanCap := utils.PopMapVal(args, ParamChanCap, 0)
		res := create(chanCap)
		e.WsOutChans[chanKey] = res
		return res
	}
}

func WriteOutChan[T any](e *Exchange, chanKey string, msg T) bool {
	outRaw, outOk := e.WsOutChans[chanKey]
	var out chan T
	if outOk {
		out = outRaw.(chan T)
		out <- msg
	} else {
		log.Error("write ws out chan fail", zap.String("k", chanKey))
	}
	return outOk
}

func (e *Exchange) AddWsChanRefs(chanKey string, keys ...string) {
	data, ok := e.WsChanRefs[chanKey]
	if !ok {
		data = make(map[string]struct{})
		e.WsChanRefs[chanKey] = data
	}
	for _, k := range keys {
		data[k] = struct{}{}
	}
}

func (e *Exchange) DelWsChanRefs(chanKey string, keys ...string) int {
	data, ok := e.WsChanRefs[chanKey]
	if !ok {
		return -1
	}
	for _, k := range keys {
		delete(data, k)
	}
	hasNum := len(data)
	if hasNum == 0 {
		if out, ok := e.WsOutChans[chanKey]; ok {
			val := reflect.ValueOf(out)
			if val.Kind() == reflect.Chan {
				val.Close()
			}
			delete(e.WsOutChans, chanKey)
			log.Info("remove chan", zap.String("key", chanKey))
		}
	}
	return hasNum
}

func (e *Exchange) handleWsClientClosed(client *WsClient) int {
	prefix := client.URL + "#"
	removeNum := 0
	for key, _ := range e.WsChanRefs {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		delete(e.WsChanRefs, key)
		if out, ok := e.WsOutChans[key]; ok {
			val := reflect.ValueOf(out)
			if val.Kind() == reflect.Chan {
				val.Close()
			}
			delete(e.WsOutChans, key)
			removeNum += 1
		}
	}
	return removeNum
}

/*
CheckWsError
从websocket返回的消息结果中，检查是否有错误信息
*/
func CheckWsError(msg map[string]string) *errs.Error {
	errRaw, ok := msg["error"]
	if ok {
		var err = &errs.Error{}
		errData, _ := sonic.Marshal(errRaw)
		_ = sonic.Unmarshal(errData, err)
		return err
	}
	status, ok := msg["status"]
	if ok && status != "200" {
		statusVal, e := strconv.Atoi(status)
		if e != nil {
			return nil
		}
		msgStr, _ := sonic.MarshalString(msg)
		return errs.NewMsg(statusVal, msgStr)
	}
	return nil
}

/*
Write
发送消息到ws服务器，可设置处理任务结果需要的信息
jobID: 此次消息的任务ID，唯一标识此次请求
jobInfo: 此次任务的主要信息，在收到任务结果时使用
*/
func (c *WsClient) Write(msg interface{}, info *WsJobInfo) *errs.Error {
	data, err2 := sonic.Marshal(msg)
	if err2 != nil {
		return errs.New(errs.CodeUnmarshalFail, err2)
	}
	if info != nil {
		if info.ID == "" {
			return errs.NewMsg(errs.CodeParamRequired, "WsJobInfo.ID is required")
		}
		if _, ok := c.JobInfos[info.ID]; !ok {
			c.JobInfos[info.ID] = info
		}
	}
	c.Send <- data
	return nil
}

func (c *WsClient) Close() {
	c.control <- ctrlDoClose
}

func (c *WsClient) write() {
	zapUrl := zap.String("url", c.URL)
	defer func() {
		log.Debug("stop write ws", zapUrl)
		err := c.Conn.Close()
		if err != nil {
			log.Error("close ws error", zapUrl, zap.Error(err))
		}
		close(c.control)
		c.Conn = nil // 置为nil表示连接已关闭
	}()
	for {
		select {
		case ctrlType, ok := <-c.control:
			if !ok {
				log.Error("read control fail", zap.Int("flag", ctrlType))
				continue
			}
			if ctrlType == ctrlClosed {
				return
			} else if ctrlType == ctrlDoClose {
				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.Conn.WriteClose()
				if err != nil {
					log.Error("write ws close error", zapUrl, zap.Error(err))
					return
				}
			} else {
				log.Error("invalid ws control type", zapUrl, zap.Int("val", ctrlType))
			}
		case msg, ok := <-c.Send:
			if !ok {
				err := c.Conn.WriteClose()
				if err != nil {
					log.Error("write ws close error", zapUrl, zap.Error(err))
					return
				}
				log.Info("WsClient.Send closed", zapUrl)
				return
			}
			w, err := c.Conn.NextWriter()
			if err != nil {
				log.Error("failed to create Ws.Writer", zapUrl, zap.Error(err))
				return
			}
			_, err = w.Write(msg)
			if err != nil {
				log.Error("write ws fail", zapUrl, zap.Error(err))
			}
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, err = w.Write(<-c.Send)
				if err != nil {
					log.Error("write ws fail", zapUrl, zap.Error(err))
				}
			}
			if err := w.Close(); err != nil {
				log.Error("close WriteCloser fail", zapUrl, zap.Error(err))
				return
			}
		}
	}
}

func (c *WsClient) read() {
	defer func() {
		c.control <- ctrlClosed
	}()
	for {
		msgRaw, err := c.Conn.ReadMsg()
		if err != nil {
			if c.OnClose != nil {
				c.OnClose(c.URL, errs.New(errs.CodeWsReadFail, err))
			}
			log.Error("read fail, ws closed", zap.String("url", c.URL), zap.Error(err))
			return
		}
		// 这里不能对每个消息启动一个goroutine，否则会导致消息处理顺序错误
		c.handleRawMsg(msgRaw)
	}
}

func (c *WsClient) handleRawMsg(msgRaw []byte) {
	msgText := string(msgRaw)
	var err *errs.Error
	var err_ error
	var id string
	if strings.HasPrefix(msgText, "{") {
		var msg = make(map[string]interface{})
		err_ = sonic.UnmarshalString(msgText, &msg)
		if err_ == nil {
			id = c.handleMsg(utils.MapValStr(msg))
		}
	} else if strings.HasPrefix(msgText, "[") {
		var msgs = make([]map[string]interface{}, 0)
		err_ = sonic.UnmarshalString(msgText, &msgs)
		if err_ == nil && len(msgs) > 0 {
			for _, it := range msgs {
				id = c.handleMsg(utils.MapValStr(it))
			}
		}
	} else {
		err = errs.NewMsg(errs.CodeWsInvalidMsg, "invalid ws msg, not dict or list")
	}
	if err_ != nil {
		err = errs.New(errs.CodeUnmarshalFail, err_)
	}
	if err != nil {
		if c.OnError != nil {
			c.OnError(c.URL, err)
		}
		log.Error("invalid ws msg", zap.String("msg", msgText), zap.Error(err))
	} else if id != "" {
		delete(c.JobInfos, id)
	}
}

func (c *WsClient) handleMsg(msg map[string]string) string {
	id, ok := msg["id"]
	if ok {
		if sub, ok := c.JobInfos[id]; ok && sub.Method != nil {
			// 订阅信息中提供了处理函数，则调用处理函数
			sub.Method(c.URL, msg, sub)
			delete(c.JobInfos, id)
			return id
		}
	}
	// 未匹配则调用通用消息处理
	c.OnMessage(c.URL, msg)
	return ""
}
