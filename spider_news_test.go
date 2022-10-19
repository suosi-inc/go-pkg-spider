package spider

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/x-funs/go-fun"
)

func TestNews_GetNews(t *testing.T) {
	// n := NewNews("https://eastday.com/", 2, true)
	n := NewNews("http://yoka.com/", nil, 1, false)
	// n := NewNews("http://www.cankaoxiaoxi.com/", 3, true)

	n.GetNews(n.GetContentNews)
	// n.GetNews(n.PrintContentNews)

	go goFunc(n, t)
	// goFunc(n, t)

	n.Wg.Wait()

	n.Close()
	t.Log("close chan")

	t.Log("crawl finish")
}

func goFunc(n *News, t *testing.T) {
	for {
		select {
		case data, ok := <-n.DataChan:
			if !ok {
				t.Log("dataChan closed")
				return
			}

			t.Log("dataChan:", (*data).Title, (*data).Lang)
			// case <-time.After(10 * time.Second):
			// 	t.Log("time select*****************")
			// 	return

			// default:
			// 	time.Sleep(1 * time.Second)
			// 	t.Log("default")
		}
	}
}

func TestNews_GetNews2(t *testing.T) {
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

	// n := NewNews("https://eastday.com/", 2, true)
	n := NewNews("http://yoka.com/", req, 1, false)
	// n := NewNews("http://www.cankaoxiaoxi.com/", 3, true)

	n.GetNews(n.GetContentNews)
	// n.GetNews(n.PrintContentNews)

	go goFunc(n, t)

	n.Wg.Wait()

	n.Close()
	t.Log("close chan")

	t.Log("crawl finish")
}
