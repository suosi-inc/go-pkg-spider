package spider

import (
	"testing"
)

func TestNews_GetNews(t *testing.T) {
	n := &News{
		url:   "https://www.163.com",
		depth: 1,
		isSub: true,
	}

	data := n.GetNews()
	t.Log(data)
}
