package spider

import (
	"testing"
	"time"
)

func TestNews_GetNews(t *testing.T) {

	// n := NewNews("https://eastday.com/", 2, true)
	n := NewNews("http://yoka.com/", 1, false)
	// n := NewNews("http://www.cankaoxiaoxi.com/", 1, false)

	n.GetNews(n.GetContentNews)
	// n.GetNews(n.PrintContentNews)

	go goFunc(n, t)

	// go func() {
	// 	for {
	// 		select {
	// 		case data := <-n.DataChan:
	// 			t.Log("dataChan:", data)
	// 			t.Log("\n")
	// 		case <-time.After(time.Second):
	// 			t.Log("time select")
	// 		default:
	// 			t.Log("default")
	// 		}
	// 	}
	// }()

	// t.Log(n.GetData())
}

func goFunc(n *News, t *testing.T) {
	for {
		select {
		case data := <-n.DataChan:
			t.Log("dataChan:", data)
			t.Log("\n")
		case <-time.After(time.Second):
			t.Log("time select")
		default:
			t.Log("default")
		}
	}
}
