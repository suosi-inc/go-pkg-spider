package spider

import (
	"testing"

	"github.com/x-funs/go-fun"
)

const (
	TestUrl = "http://localhost:8080"
)

func TestHttpGet(t *testing.T) {
	var urlStrs = []string{
		"http://suosi.com.cn",
		"https://www.163.com",
		"https://english.news.cn",
		"https://jp.news.cn",
		"https://kr.news.cn",
		"https://www.donga.com/",
		"http://www.koreatimes.com/",
		"https://arabic.news.cn",
		"https://www.bbc.com",
		"http://government.ru",
		"https://french.news.cn",
		"https://www.gouvernement.fr",
		"http://live.siammedia.org/",
		"http://hanoimoi.com.vn",
		"https://www.commerce.gov.mm",
		"https://sanmarg.in/",
		"https://www.rrdmyanmar.gov.mm",
	}

	for _, urlStr := range urlStrs {
		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(urlStr)
		t.Log(err)
		t.Log(resp.Success)
		t.Log(resp.Charset)
		t.Log(resp.Lang)
		t.Log(resp.ContentLength)
		t.Log(resp.Headers)
	}
}

func TestHttpGetPublic(t *testing.T) {
	var urlStr string

	urlStr = "http://www.163.com"
	// urlStr = "http://www.qq.com"

	resp, err := HttpGetResp(urlStr, nil, 10000)

	t.Log(urlStr)
	t.Log(err)
	t.Log(resp.Success)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(resp.ContentLength)
	t.Log(resp.Headers)
	t.Log(fun.String(resp.Body))
}

func TestHttpGetContentType(t *testing.T) {
	var urlStr string

	urlStr = "https://mirrors.163.com/mysql/Downloads/MySQL-8.0/libmysqlclient-dev_8.0.27-1debian10_amd64.deb"

	req := &HttpReq{
		ForceTextContentType: true,
	}
	resp, err := HttpGetResp(urlStr, req, 10000)

	t.Log(urlStr)
	t.Log(err)
	t.Log(resp.Success)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(resp.ContentLength)
	t.Log(resp.Headers)
	t.Log(fun.String(resp.Body))
}

func TestHttpGetContentLength(t *testing.T) {
	var urlStr string

	urlStr = "http://suosi.com.cn"

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 1000,
		},
	}
	resp, err := HttpGetResp(urlStr, req, 10000)

	t.Log(urlStr)
	t.Log(err)
	t.Log(resp.Success)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(resp.ContentLength)
	t.Log(resp.Headers)
	t.Log(fun.String(resp.Body))
}
