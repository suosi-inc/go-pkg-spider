package spider

import (
	"testing"
)

func TestNews_GetNews(t *testing.T) {

	// n := NewNews("https://eastday.com/", 2, true)
	// n := NewNews("http://yoka.com/", 2, true)
	n := NewNews("http://www.cankaoxiaoxi.com/", 2, true)

	go func() {
		for {
			select {
			case data := <-n.dataChan:
				t.Log(*data)
				t.Log("\n")
			// case <-time.After(time.Second):
			// 	t.Log("time select")
			default:
			}
		}
	}()

	n.GetNews(n.GetContentNews)

	t.Log(n.GetData())
}
