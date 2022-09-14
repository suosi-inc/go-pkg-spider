package spider

import (
	"testing"
)

func TestNews_GetNews(t *testing.T) {
	// n := NewNews("https://eastday.com/", 2, true)
	n := NewNews("http://yoka.com/", 1, false)
	// n := NewNews("http://www.cankaoxiaoxi.com/", 1, false)

	n.GetNews(n.GetContentNews)
	// n.GetNews(n.PrintContentNews)

	go goFunc(n, t)

	n.Wg.Wait()

	// t.Log(n.GetData())

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

			t.Log("dataChan:", *data)
			// case <-time.After(time.Second):
			// 	t.Log("time select")
			// default:
			// t.Log("default")
		}
	}
}
