package spider

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/suosi-inc/go-pkg-spider/detect"
	"github.com/x-funs/go-fun"
)

const (
	HttpDefaultTimeOut        = 10000
	HttpDefaultUserAgent      = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
	HttpDefaultAcceptEncoding = "gzip, deflate"
)

var (
	textContentTypes = []string{
		"text/html",
		"text/xml",
		"application/xml",
		"application/xhtml+xml",
		"application/json",
		"application/javascript",
	}
)

type HttpReq struct {
	*fun.HttpReq

	// 禁止自动探测字符集和语言
	DisableCharsetLang bool

	// 强制 ContentType 为文本类型
	ForceTextContentType bool
}

type HttpResp struct {
	*fun.HttpResp

	// 语言 (en/zh/...)
	Lang detect.LangRes

	// 字符集
	Charset detect.CharsetRes
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
	if r == nil || r.HttpReq == nil || r.HttpReq.Transport == nil {
		r = &HttpReq{
			HttpReq: &fun.HttpReq{
				Transport: HttpDefaultTransport,
			},
		}
	}

	// 强制文本类型
	if r != nil && r.ForceTextContentType {
		r.AllowedContentTypes = textContentTypes
	}

	resp, err := fun.HttpDoResp(req, r.HttpReq, timeout)

	// HttpResp
	var lang detect.LangRes
	var charset detect.CharsetRes
	httpResp := &HttpResp{
		HttpResp: resp,
		Lang:     lang,
		Charset:  charset,
	}

	if err != nil {
		return httpResp, err
	}

	body := httpResp.Body
	// 编码语言探测与自动转码
	if err != nil {
		httpResp.Success = false
		return httpResp, err
	} else {
		// 默认会自动进行编码和语种探测，除非手动禁用
		if r == nil || !r.DisableCharsetLang {
			charsetRes, langRes := CharsetLang(body, httpResp.Headers, req.URL.Hostname())
			httpResp.Lang = langRes
			httpResp.Charset = charsetRes

			if charsetRes.Charset != "" && charsetRes.Charset != "utf-8" {
				utf8Body, err := fun.ToUtf8(body, charsetRes.Charset)
				if err != nil {
					return httpResp, errors.New("charset detect to utf-8 error")
				} else {
					httpResp.Body = utf8Body
				}
			} else {
				httpResp.Body = body
			}
		} else {
			httpResp.Body = body
		}
	}

	return httpResp, nil
}
