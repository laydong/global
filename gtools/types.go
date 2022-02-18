package gtools

import (
	jsonIter "github.com/json-iterator/go"
)

const (
	XForwardedFor = "X-Forwarded-For" // 获取真实ip
	XRealIP       = "X-Real-IP"       // 获取真实ip
	RequestIdKey  = "request_id"      // 日志key
)

var (
	// CJson 全局json序列化和反序列化
	CJson = jsonIter.ConfigCompatibleWithStandardLibrary
)
