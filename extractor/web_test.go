package extractor

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	spider "github.com/suosi-inc/go-pkg-spider"
)

func TestLinkTitles(t *testing.T) {
	var urlStrs = []string{
		"https://www.qq.com",
		// "https://www.36kr.com",
	}

	for _, urlStr := range urlStrs {

		resp, err := spider.HttpGetResp(urlStr, nil, 30000)

		t.Log(urlStr)
		t.Log(err)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(spider.DefaultRemoveTags).Remove()

		linkTitles := LinkTitles(doc, urlStr, true)
		// fmt.Println(len(linkTitles))
		for link, title := range linkTitles {
			if utf8.RuneCountInString(title) > 6 {
				fmt.Println(link + "	->	" + title)
			}

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
