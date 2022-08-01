package extract

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider"
)

func TestLinkTitles(t *testing.T) {
	var urlStrs = []string{
		// "https://www.qq.com",
		// "https://www.36kr.com",
		// "https://www.163.com",
		"https://www.sohu.com",
		// "http://jyj.suqian.gov.cn",
		// "http://www.news.cn",
		// "http://www.cankaoxiaoxi.com",
	}

	for _, urlStr := range urlStrs {

		resp, err := spider.HttpGetResp(urlStr, nil, 30000)

		t.Log(urlStr)
		t.Log(err)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(spider.DefaultRemoveTags).Remove()

		linkTitles := LinkTitles(doc, urlStr, true)
		// fmt.Println(len(linkTitles))

		fmt.Println(len(linkTitles))

		var contentLinks = make(map[string]string, 0)
		for a, title := range linkTitles {
			if spider.IsContentByLang(a, title, "zh") {
				contentLinks[a] = title
			}
		}

		fmt.Println(len(contentLinks))
		for a, title := range contentLinks {
			fmt.Println(a, title)
		}

	}
}

func TestUrlParse(t *testing.T) {
	var urlStrs = []string{
		"https://www.163.com",
		"https://www.163.com/",
		"https://www.163.com/a",
		"https://www.163.com/aa.html",
		"https://www.163.com/a/b",
		"https://www.163.com/a/bb.html",
		"https://www.163.com/a/b/",
		"https://www.163.com/a/b/c",
		"https://www.163.com/a/b/cc.html",
	}

	for _, urlStr := range urlStrs {
		u, _ := url.Parse(urlStr)
		link := "javascript:;"
		absolute, err := u.Parse(link)
		t.Log(err)

		_, err = url.Parse(absolute.String())
		if err != nil {
			t.Log(err)
		}

		t.Log(urlStr + "	+ " + link + " => " + absolute.String())
	}

}
