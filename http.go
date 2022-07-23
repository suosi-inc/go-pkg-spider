package spider

import (
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/suosi-inc/go-pkg-spider/detect"
	"github.com/x-funs/go-fun"
)

const (
	HttpDefaultTimeOut          = 10000
	HttpDefaultUserAgent        = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
	HttpDefaultMaxContentLength = 10 * 1024 * 1024
)

var (
	// refer: https://www.iana.org/assignments/media-types/media-types.xhtml
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
	// UserAgent 优先于请求头 Headers 中的 User-Agent 字段
	UserAgent string

	// 请求头
	Headers map[string]string

	// 禁止
	DisableCharsetLang bool

	// 强制 ContentType 为文本类型
	ForceTextContentType bool

	// 限制最大返回大小
	MaxContentLength int64

	// 最大跳转次数
	MaxRedirects int

	// 限制允许访问 ContentType 列表
	AllowedContentTypes []string

	// http.Transport
	Transport http.RoundTripper
}

type HttpResp struct {
	// 是否成功 (200-299)
	Success bool

	// Http 状态码
	StatusCode int

	// 响应体
	Body []byte

	// 语言 (en/zh/...)
	Lang string

	// 字符集
	Charset string

	// ContentLength (字节数)
	ContentLength int64

	// ContentType
	ContentType string

	// 响应头
	Headers *http.Header
}

// HttpDefaultTransport 默认全局使用的 http.Transport
var HttpDefaultTransport = &http.Transport{
	DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
	ForceAttemptHTTP2:     true,
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

	return nil, errors.New("HttpGet params error")
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
	if timeout == 0 {
		timeout = HttpDefaultTimeOut
	}

	// NewClient
	var client *http.Client
	if r != nil && r.Transport != nil {
		client = &http.Client{
			Timeout:   time.Duration(timeout) * time.Millisecond,
			Transport: r.Transport,
		}
	} else {
		client = &http.Client{
			Timeout:   time.Duration(timeout) * time.Millisecond,
			Transport: HttpDefaultTransport,
		}
	}

	// 处理请求头
	headers := make(map[string]string)
	if r != nil && r.UserAgent != "" {
		r.Headers["User-Agent"] = r.UserAgent
	}
	if r != nil && r.Headers != nil && len(r.Headers) > 0 {
		headers = r.Headers
		if _, exist := headers["User-Agent"]; !exist {
			headers["User-Agent"] = HttpDefaultUserAgent
		}
	} else {
		headers["User-Agent"] = HttpDefaultUserAgent
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// HttpResp
	httpResp := &HttpResp{
		Success:       false,
		StatusCode:    0,
		Body:          nil,
		ContentLength: 0,
		Headers:       nil,
	}

	resp, err := client.Do(req)
	if err != nil {
		return httpResp, err
	}

	// 状态码
	httpResp.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		httpResp.Success = true
	} else {
		return httpResp, errors.New("http Status code error")
	}

	httpResp.Headers = &resp.Header

	// ContentType 限制
	if _, err := validContentType(r, httpResp.Headers); err != nil {
		return httpResp, err
	}

	// ContentLength 限制，并限制 Body 读取
	var body []byte
	httpResp.ContentLength = resp.ContentLength
	if r != nil && r.MaxContentLength > 0 {
		if resp.ContentLength != -1 {
			if resp.ContentLength > r.MaxContentLength {
				httpResp.Success = false
				return httpResp, errors.New("ContentLength > MaxContentLength ")
			}
			body, err = ioutil.ReadAll(resp.Body)
		} else {

			body, err = ioutil.ReadAll(io.LimitReader(resp.Body, r.MaxContentLength))
		}
	} else {
		body, err = ioutil.ReadAll(resp.Body)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if err != nil {
		return httpResp, err
	} else {
		// 编码语言探测与转换
		if r == nil || r.DisableCharsetLang {
			charset, lang := detect.CharsetLang(body, httpResp.Headers)

			httpResp.Lang = lang
			if charset != "" {
				httpResp.Charset = charset
				body, err := fun.ToUtf8(body, charset)
				if err != nil {
					return httpResp, errors.New("CharsetLang detect error")

				} else {
					httpResp.Body = body
				}
			}
		} else {
			httpResp.Body = body
		}

	}

	return httpResp, nil
}

func validContentType(r *HttpReq, headers *http.Header) (bool, error) {
	if r == nil {
		return true, nil
	}

	if r.ForceTextContentType || len(r.AllowedContentTypes) > 0 {
		valid := false

		ct := strings.TrimSpace(strings.ToLower(headers.Get("Content-Type")))

		// Text Content-Type
		if r.ForceTextContentType {

			for _, t := range textContentTypes {
				if strings.HasPrefix(ct, t) {
					valid = true
					break
				}
			}

			if valid {
				return valid, nil
			} else {
				return valid, errors.New("Content-Type ForceTextContentType invalid")
			}
		}

		// Custom Content-Type
		if len(r.AllowedContentTypes) > 0 {
			for _, t := range r.AllowedContentTypes {
				if strings.HasPrefix(ct, t) {
					valid = true
					break
				}
			}

			if valid {
				return valid, nil
			} else {
				return valid, errors.New("Content-Type AllowedContentTypes invalid")
			}
		}
	}

	return true, nil
}
