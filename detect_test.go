package spider

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/extract"
)

func TestDomainDetect(t *testing.T) {
	domains := []string{
		"163.com",
	}

	for _, domain := range domains {
		domainRes, err := DetectDomain(domain)
		if err == nil {
			t.Log(domainRes)
		} else {
			t.Log(err)
		}
	}
}

func TestLinkTitles(t *testing.T) {
	var urlStrs = []string{
		"https://www.qq.com",
		// "https://www.36kr.com",
		// "https://www.163.com",
		// "http://jyj.suqian.gov.cn",
		// "http://www.news.cn",
		// "http://www.cankaoxiaoxi.com",
	}

	for _, urlStr := range urlStrs {

		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(urlStr)
		t.Log(err)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultRemoveTags).Remove()

		linkTitles := extract.LinkTitles(doc, urlStr, true)
		// fmt.Println(len(linkTitles))

		fmt.Println(len(linkTitles))

		var contentLinks = make(map[string]string, 0)
		for a, title := range linkTitles {
			if extract.IsContentByLang(a, title, "zh") {
				contentLinks[a] = title
			}
		}

		fmt.Println(len(contentLinks))
		i := 1
		for a, title := range contentLinks {
			i = i + 1
			fmt.Println(i, ":"+a+"\t=>"+title)
		}

	}
}

func TestDetectIcp(t *testing.T) {
	var urlStrs = []string{
		// "http://suosi.com.cn",
		"https://www.163.com",
		// "https://www.sohu.com",
		// "https://www.qq.com",
		// "https://www.hexun.com",
		// "https://www.wfmc.edu.cn/",
		// "https://www.cankaoxiaoxi.com/",
	}

	for _, urlStr := range urlStrs {

		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(err)
		t.Log(urlStr)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultRemoveTags).Remove()
		icp, loc := extract.Icp(doc)
		t.Log(icp, loc)
	}
}

func TestIcpFromText(t *testing.T) {
	texts := []string{
		"粤ICP备17055554号",
		"粤ICP备17055554-34号",
		"沪ICP备05018492",
		"粤B2-20090059",
		"京公网安备31010402001073号",
		"京公网安备-31010-4020010-73号",
		"鲁ICP备05002386鲁公网安备37070502000027号",
	}

	for _, text := range texts {
		icp, loc := extract.IcpFromText(text)
		t.Log(icp, loc)
	}
}
