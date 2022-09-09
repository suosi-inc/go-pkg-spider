package spider

import (
	"testing"
)

func TestNews_GetNews(t *testing.T) {

	// n := NewNews("https://eastday.com/", 2, true)
	// n := NewNews("http://yoka.com/", 2, true)
	n := NewNews("http://www.cankaoxiaoxi.com/", 2, true)

	n.GetNews(n.GetContentNews)

	t.Log(n.GetData())
}
