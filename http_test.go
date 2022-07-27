package spider

import (
	"testing"

	"github.com/x-funs/go-fun"
)

const (
	TestUrl = "http://localhost:8080"
)

func TestHttpGet(t *testing.T) {
	var urlStr string
	urlStr = "http://www.163.com"
	// GB2312,zh
	// urlStr = "http://www.changzhou.gov.cn/"
	// Shift_JIS
	// urlStr = "https://chiba-shinbun.co.jp"
	// UTF-8,en
	// urlStr = "https://english.news.cn/"
	// UTF-8,ja
	// urlStr = "https://jp.news.cn/"
	// utf-8,ru
	// urlStr = "http://government.ru/"

	req := &HttpReq{
		ForceTextContentType: true,
	}
	resp, err := HttpGetResp(urlStr, req, 10000)

	t.Log(err)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(fun.String(resp.Body))
}

func TestHttpGetContentType(t *testing.T) {
	var urlStr string

	urlStr = "https://www.qq.com"

	req := &HttpReq{
		ForceTextContentType: true,
	}
	resp, err := HttpGetResp(urlStr, req, 10000)

	t.Log(err)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(fun.String(resp.Body))
}

func TestHttpGetContentLength(t *testing.T) {
	var urlStr string

	// urlStr = "https://mirrors.163.com/mysql/Downloads/MySQL-8.0/mysql-8.0.27-macos11-x86_64.tar"
	urlStr = "http://www.163.com"

	req := &HttpReq{
		MaxContentLength: 1000,
		Headers: map[string]string{
			"Accept-Encoding": "",
		},
	}
	resp, err := HttpGetResp(urlStr, req, 10000)

	t.Log(err)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(resp.ContentLength)
	t.Log(resp.Headers)
	t.Log(fun.String(resp.Body))
}
