// Package protocol 提供了 HTTP、HTTPS、NSHead、ProtoBuffer 协议支持
package protocol

import (
	"fmt"
	"github.com/climber-dong/global/gcal/contextx"
	"github.com/climber-dong/global/gcal/converter"
	"github.com/climber-dong/global/gcal/service"
)

// Protocoler 协议的接口
// 协议本身只完成数据请求
type Protocoler interface {
	Do(ctx *contextx.Context, addr string) (*Response, error)
	Protocol() string
}

var (
	_ Protocoler = &HTTPProtocol{}
)

// NewProtocol 创建协议
func NewProtocol(ctx *contextx.Context, serv service.Service, req interface{}) (p Protocoler, err error) {
	protocolName := serv.GetProtocol()

	if protocolName == "http" || protocolName == "https" {
		tmp, ok := req.(HTTPRequest)
		if !ok {
			return nil, fmt.Errorf("%s: bad request type: %T", protocolName, req)
		}
		if tmp.Converter == "" {
			tmp.Converter = converter.JSON
		}
		return NewHTTPProtocol(ctx, serv, &tmp, protocolName == "https")
	}

	return nil, fmt.Errorf("unknow protocol: %s ", protocolName)
}

// Response 通用的返回
type Response struct {
	Body      interface{}
	Head      interface{}
	Request   interface{}
	OriginRsp interface{}
}
