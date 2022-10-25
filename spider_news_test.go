package spider

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/x-funs/go-fun"
)

var (
	newUrl     = "http://www.cankaoxiaoxi.com/"
	overseaUrl = "https://www.bbc.com/news"
)

func TestNews_GetLinkRes_Noctx(t *testing.T) {
	n := NewNews(newUrl, nil, 2, false, processLink, nil)
	n.GetLinkRes()
}

func TestNews_GetLinkRes(t *testing.T) {
	ctx := "getLinkRes"
	n := NewNews(newUrl, nil, 2, false, processLink, ctx)
	n.GetLinkRes()
}

func processLink(data ...any) {
	dd := data[0].(*LinkData)
	for i := range dd.LinkRes.List {
		fmt.Println(data[1], i)
	}
}

func TestNews_GetContentNews(t *testing.T) {
	ctx := "getContentNews"
	n := NewNews(newUrl, nil, 1, false, processContent, ctx)
	n.GetContentNews()
}

func processContent(data ...any) {
	dd := data[0].(*NewsContent)
	fmt.Println(data[1], dd.Title, dd.Lang)
}

func TestNews_GetNewsWithProxy(t *testing.T) {
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	proxyString := "http://username:password@host:port"
	proxy, _ := url.Parse(proxyString)
	transport.Proxy = http.ProxyURL(proxy)

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: HttpDefaultMaxContentLength,
			MaxRedirect:      2,
			Transport:        transport,
		},
		ForceTextContentType: true,
	}

	ctx := "getNewsWithProxy"
	n := NewNews(overseaUrl, req, 1, false, processContent, ctx)
	n.GetContentNews()
}
