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
	// GB2312,zh
	// urlStr = "http://www.changzhou.gov.cn/"
	// Shift_JIS
	urlStr = "https://chiba-shinbun.co.jp"
	// UTF-8,en
	// urlStr = "https://english.news.cn/"
	// UTF-8,ja
	// urlStr = "https://jp.news.cn/"
	// utf-8,ru
	// urlStr = "http://government.ru/"

	resp, err := HttpGetResp(urlStr, nil, 10000)

	t.Log(err)
	t.Log(resp.Charset)
	t.Log(resp.Lang)
	t.Log(fun.String(resp.Body))
}
