package spider

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/x-funs/go-fun"
)

const (
	HttpDefaultTimeOut          = 10000
	HttpDefaultMaxContentLength = 10 * 1024 * 1024
	HttpDefaultUserAgent        = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
	HttpDefaultAcceptEncoding   = "gzip, deflate"
)

var (
	textContentTypes = []string{
		"text/plain",
		"text/html",
		"text/xml",
		"application/xml",
		"application/xhtml+xml",
		"application/json",
	}
)

type HttpReq struct {
	// 嵌入 fun.HttpReq
	*fun.HttpReq

	// 禁止自动探测字符集和转换字符集
	DisableCharset bool

	// 强制 ContentType 为文本类型
	ForceTextContentType bool
}

type HttpResp struct {
	*fun.HttpResp

	// 字符集
	Charset CharsetRes
}

// HttpDefaultTransport 默认全局使用的 http.Transport
var HttpDefaultTransport = &http.Transport{
	DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
	DisableKeepAlives:     true,
	IdleConnTimeout:       60 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
}

// HttpGet 参数为请求地址 (HttpReq, 超时时间)
// HttpGet(url)、HttpGet(url, HttpReq)、HttpGet(url, timeout)、HttpGet(url, HttpReq, timeout)
// 返回 body, 错误信息
func HttpGet(urlStr string, args ...any) ([]byte, error) {
	l := len(args)

	switch l {
	case 0:
		return HttpGetDo(urlStr, nil, 0)
	case 1:
		switch v := args[0].(type) {
		case int:
			timeout := fun.ToInt(args[0])
			return HttpGetDo(urlStr, nil, timeout)
		case *HttpReq:
			return HttpGetDo(urlStr, v, 0)

		}
	case 2:
		timeout := fun.ToInt(args[1])
		switch v := args[0].(type) {
		case *HttpReq:
			return HttpGetDo(urlStr, v, timeout)
		}

	}

	return nil, errors.New("http get params error")
}

// HttpGetDo Http Get 请求, 参数为请求地址, HttpReq, 超时时间(毫秒)
// 返回 body, 错误信息
func HttpGetDo(urlStr string, r *HttpReq, timeout int) ([]byte, error) {
	resp, err := HttpGetResp(urlStr, r, timeout)
	if err != nil {
		return nil, err
	} else {
		return resp.Body, nil
	}
}

// HttpGetResp Http Get 请求, 参数为请求地址, HttpReq, 超时时间(毫秒)
// 返回 HttpResp, 错误信息
func HttpGetResp(urlStr string, r *HttpReq, timeout int) (*HttpResp, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	return HttpDoResp(req, r, timeout)
}

// HttpDo Http 请求, 参数为 http.Request, HttpReq, 超时时间(毫秒)
// 返回 body, 错误信息
func HttpDo(req *http.Request, r *HttpReq, timeout int) ([]byte, error) {
	resp, err := HttpDoResp(req, r, timeout)
	if err != nil {
		return nil, err
	} else {
		return resp.Body, nil
	}
}

// HttpDoResp Http 请求, 参数为 http.Request, HttpReq, 超时时间(毫秒)
// 返回 HttpResp, 错误信息
func HttpDoResp(req *http.Request, r *HttpReq, timeout int) (*HttpResp, error) {
	// 处理 Transport
	if r == nil {
		r = &HttpReq{
			HttpReq: &fun.HttpReq{
				Transport: HttpDefaultTransport,
			},
		}
	} else if r.HttpReq == nil {
		r.HttpReq = &fun.HttpReq{
			Transport: HttpDefaultTransport,
		}
	} else if r.Transport == nil {
		r.Transport = HttpDefaultTransport
	}

	// 强制文本类型
	if r != nil && r.ForceTextContentType {
		r.AllowedContentTypes = textContentTypes
	}

	// HttpResp
	var charset CharsetRes
	httpResp := &HttpResp{
		Charset: charset,
	}

	resp, err := fun.HttpDoResp(req, r.HttpReq, timeout)
	httpResp.HttpResp = resp
	if err != nil {
		return httpResp, err
	}

	// 默认会自动进行探测编码和转码, 除非手动禁用
	if r == nil || !r.DisableCharset {
		charsetRes := Charset(httpResp.Body, httpResp.Headers)
		httpResp.Charset = charsetRes

		if charsetRes.Charset != "" && charsetRes.Charset != "UTF-8" {
			utf8Body, e := fun.ToUtf8(httpResp.Body, charsetRes.Charset)
			if e != nil {
				return httpResp, errors.New("ErrorCharset")
			} else {
				httpResp.Body = utf8Body
			}
		}
	}

	return httpResp, nil
}
