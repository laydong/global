package protocol

import (
	"fmt"
	"github.com/climber-dong/global/core/metautils"
	"github.com/climber-dong/global/gcal/contextx"
	"github.com/climber-dong/global/gcal/converter"
	"github.com/climber-dong/global/gcal/service"
	"github.com/climber-dong/global/genv"
	"github.com/climber-dong/global/gtools"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
)

const UA = "GCAL/" + genv.VERSION + " (laya gcal http client)"
const HttpClientAlive = 5 * time.Minute

// HTTPRequest http request 对象
type HTTPRequest struct {
	CustomAddr string

	Header      map[string][]string
	Method      string
	Body        interface{}
	Path        string
	QueryParams url.Values
	RequestId   string

	Converter converter.ConverterType
	Ctx       contextx.RequestContext
}

// HTTPHead HTTPResponse, 兼容历史
type HTTPHead struct {
	Status        string
	StatusCode    int
	Proto         string
	Header        map[string][]string
	ContentLength int64
}

// HTTPProtocol http 协议
type HTTPProtocol struct {
	protocol  string
	serv      service.Service
	requestId string

	originReq *HTTPRequest
	RawReq    *http.Request
	// OriginRsp *http.Response
}

// Protocol 返回类型
func (hp *HTTPProtocol) Protocol() string {
	return hp.protocol
}

// initRequestId 生成requestId
func (hp *HTTPProtocol) initRequestId(ctx *contextx.Context) {
	requestId := hp.originReq.RequestId

	if requestId == "" {
		if ctx.ReqContext != nil {
			requestId = ctx.ReqContext.GetLogId()
		}
	}

	if requestId == "" {
		requestId = gtools.GenerateLogId()
	}

	hp.requestId, ctx.LogId = requestId, requestId
}

// NewHTTPProtocol 创建一个 Http Protocol
func NewHTTPProtocol(ctx *contextx.Context, serv service.Service, req *HTTPRequest, isHTTPS bool) (hp *HTTPProtocol, err error) {
	hp = &HTTPProtocol{
		serv:      serv,
		originReq: req,
		protocol:  "http",
	}
	if isHTTPS {
		hp.protocol = "https"
	}

	ctx.ReqContext = req.Ctx
	hp.initRequestId(ctx)
	ctx.Method = strings.ToLower(req.Method)

	hp.RawReq = &http.Request{
		Method:     strings.ToUpper(req.Method),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     req.Header,
		Body:       http.NoBody,
		GetBody:    func() (io.ReadCloser, error) { return http.NoBody, nil },
	}
	if hp.RawReq.Header == nil {
		hp.RawReq.Header = make(http.Header)
	}

	bb := []byte{}
	if req.Body != nil {
		conver, _ := converter.GetConverter(req.Converter)
		if conver == nil {
			err = fmt.Errorf("get convert %s failed", req.Converter)
			return
		}

		ctx.PackStatis.StartPoint = time.Now()
		bb, err = conver.Pack(req.Body)
		ctx.PackStatis.StopPoint = time.Now()
		if err != nil {
			return
		}
	}

	if len(bb) > 0 {
		body := strings.NewReader(string(bb))
		hp.RawReq.ContentLength = int64(body.Len())
		hp.RawReq.Body = ioutil.NopCloser(body)
		snapshot := *body
		hp.RawReq.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
	}

	ctx.ReqLen = hp.RawReq.ContentLength

	// set logId and reject tracex
	hp.RawReq.Header.Set(gtools.RequestIdKey, hp.requestId)
	req.Ctx.SpanInject(metautils.NiceMD(hp.RawReq.Header))

	// If the user doesn't set User-Agent, set the default User-Agent
	if hp.RawReq.Header.Get("User-Agent") == "" {
		hp.RawReq.Header.Set("User-Agent", UA)
	}

	return
}

// Do 发送请求
func (hp *HTTPProtocol) Do(ctx *contextx.Context, addr string) (rsp *Response, err error) {
	var host string
	if hp.originReq.CustomAddr != "" {
		host = fmt.Sprintf("%s", hp.originReq.CustomAddr)
	} else {
		host = addr
	}
	ctx.CurRecord().IPPort = host

	path := hp.originReq.Path
	if len(hp.originReq.QueryParams) > 0 {
		path += "?"
		path += hp.originReq.QueryParams.Encode()
	}
	var fullPath string
	if path == "" {
		fullPath = fmt.Sprintf("%s://%s", hp.Protocol(), addr)
	} else {
		fullPath = fmt.Sprintf("%s://%s/%s", hp.Protocol(), addr, path)
	}

	u, err := url.Parse(fullPath)
	if err != nil {
		return nil, err
	}

	ctx.CurRecord().Path = u.Path

	hp.RawReq.URL = u
	if hp.RawReq.Host == "" {
		hp.RawReq.Host = u.Host
	}

	ctx.CurRecord().Host = hp.RawReq.Host

	trace := &httptrace.ClientTrace{
		GetConn: func(hostport string) {
			ctx.TimeStatisStart("connect")
			ctx.TimeStatisStart("talk")
			ctx.CurRecord().RecordTimePoint("req_start_time")
		},
		GotConn: func(info httptrace.GotConnInfo) {
			ctx.TimeStatisStop("connect")
			ctx.TimeStatisStart("write")
		},
		ConnectStart: func(network, addr string) {
			//ctx.TimeStatisStart("talk")
		},
		ConnectDone: func(network, addr string, err error) {
			//ctx.TimeStatisStart("talk")
		},
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			ctx.TimeStatisStart("dnslookup")
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			ctx.TimeStatisStop("dnslookup")
		},

		WroteRequest: func(writeRequestInfo httptrace.WroteRequestInfo) {
			ctx.TimeStatisStop("write")
		},
	}

	httpReq := hp.RawReq.WithContext(httptrace.WithClientTrace(hp.RawReq.Context(), trace))

	client, err := hp.getClient(ctx)
	if err != nil {
		return nil, err
	}
	if hp.serv.GetReuse() {
		defer hp.tryReuseClient(client)
	}

	originRsp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer func() {
		originRsp.Body.Close()
		ctx.TimeStatisStop("talk")
	}()

	ctx.CurRecord().RspCode = originRsp.StatusCode

	ctx.TimeStatisStart("read")
	raw, err := ioutil.ReadAll(originRsp.Body)
	ctx.TimeStatisStop("read")
	if err != nil {
		return nil, err
	}
	rsp = &Response{
		Request: originRsp.Request,
		Head: HTTPHead{
			Status:        originRsp.Status,
			StatusCode:    originRsp.StatusCode,
			Proto:         originRsp.Proto,
			Header:        originRsp.Header,
			ContentLength: originRsp.ContentLength,
		},
		Body:      raw,
		OriginRsp: originRsp,
	}

	ctx.RspLen = int64(len(raw))

	return
}

func (hp *HTTPProtocol) tryReuseClient(cli *http.Client) {
	service2httpClientMap.Store(hp.serv.GetName(), cli)
}

var service2httpClientMap sync.Map
var lock sync.Mutex

func (hp *HTTPProtocol) getClient(ctx *contextx.Context) (client *http.Client, err error) {
	if !hp.serv.GetReuse() {
		return DefaultHTTPClientFactory(hp.serv)
	}
	cli, ok := service2httpClientMap.Load(hp.serv.GetName())
	if !ok {
		lock.Lock()
		cli, ok = service2httpClientMap.Load(hp.serv.GetName())
		if !ok {
			client, err = DefaultHTTPClientFactory(hp.serv)
			service2httpClientMap.Store(hp.serv.GetName(), client)
			lock.Unlock()
			go func(name string) {
				select {
				case <-time.After(HttpClientAlive):
					service2httpClientMap.Delete(name)
				}
			}(hp.serv.GetName())
			return
		}
		lock.Unlock()
	}
	return cli.(*http.Client), nil
}

// DefaultHTTPClientFactory 默认的 http client factory
var DefaultHTTPClientFactory = func(serv service.Service) (cli *http.Client, err error) {
	var proxyURL *url.URL

	perHost := -1
	if serv.GetReuse() {
		perHost = 2
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{
				Timeout:   serv.GetConnTimeout(),
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConnsPerHost:   perHost,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: serv.GetTotalTimeout(),
	}, nil
}
