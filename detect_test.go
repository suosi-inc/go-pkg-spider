package spider

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/extract"
	"github.com/x-funs/go-fun"
)

func TestDomainDetect(t *testing.T) {
	domains := []string{
		// "163.com",
		"cankaoxiaoxi.com",
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

func BenchmarkLinkTitles(b *testing.B) {
	urlStr := "http://www.163.com"

	resp, _ := HttpGetResp(urlStr, nil, 30000)

	// 解析 HTML
	u, _ := url.Parse(urlStr)
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	doc.Find(DefaultRemoveTags).Remove()

	// 语言
	langRes := Lang(doc, resp.Charset.Charset, u.Hostname())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 标题
		linkTitles := extract.WebLinkTitles(doc, urlStr, true)

		// 连接和子域名
		_, _ = extract.LinkTypes(linkTitles, langRes.Lang, "")
		//_, _ = extract.LinkTypes(linkTitles, langRes.Lang, `\d{7}\.shtml$`)
	}

	//fmt.Println(langRes.Lang)

}

func TestLinkTitles(t *testing.T) {
	var urlStrs = []string{
		// "https://www.qq.com",
		// "https://www.people.com.cn",
		// "https://www.36kr.com",
		//"https://www.163.com",
		//"http://jyj.suqian.gov.cn",
		// "http://www.news.cn",
		//"http://www.cankaoxiaoxi.com",
		// "http://www.bbc.com",
		// "https://www.ft.com",
		// "https://www.reuters.com/",
		// "https://nypost.com/",
		"http://www.mengcheng.gov.cn/",
	}

	for _, urlStr := range urlStrs {

		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(urlStr)
		t.Log(err)

		// 解析 HTML
		u, _ := url.Parse(urlStr)
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultRemoveTags).Remove()

		// 语言
		langRes := Lang(doc, resp.Charset.Charset, u.Hostname())

		// 标题
		linkTitles := extract.WebLinkTitles(doc, urlStr, true)

		// 连接和子域名
		linkRes, domainRes := extract.LinkTypes(linkTitles, langRes.Lang, "")
		//linkRes, domainRes := extract.LinkTypes(linkTitles, langRes.Lang, `\d{7}\.shtml$`)

		fmt.Println("all:", len(linkTitles))
		fmt.Println("content:", len(linkRes.Content))
		fmt.Println("list:", len(linkRes.List))
		fmt.Println("none:", len(linkRes.None))
		i := 0
		for subdomain, _ := range domainRes {
			i = i + 1
			fmt.Println(i, "domain:"+subdomain)
		}
		i = 0
		for a, title := range linkRes.Content {
			i = i + 1
			fmt.Println(i, "content:"+a+"\t=>\t"+title)
		}
		i = 0
		for a, title := range linkRes.List {
			i = i + 1
			fmt.Println(i, "list:"+a+"\t=>\t"+title)
		}
		i = 0
		for a, title := range linkRes.None {
			i = i + 1
			fmt.Println(i, "none:"+a+"\t=>\t"+title)
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

func TestLangFromUtf8Body(t *testing.T) {
	var urlStrs = []string{
		// "https://www.163.com",
		// "https://english.news.cn",
		// "https://jp.news.cn",
		// "https://kr.news.cn",
		// "https://arabic.news.cn",
		// "https://www.bbc.com",
		// "http://government.ru",
		// "https://french.news.cn",
		// "https://www.gouvernement.fr",
		// "http://live.siammedia.org/",
		// "http://hanoimoi.com.vn",
		// "https://www.commerce.gov.mm",
		// "https://www.rrdmyanmar.gov.mm",

	}

	for _, urlStr := range urlStrs {
		resp, _ := fun.HttpGetResp(urlStr, nil, 30000)
		u, _ := url.Parse(urlStr)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultRemoveTags).Remove()

		start := fun.Timestamp(true)
		lang, pos := LangFromUtf8Body(doc, u.Hostname())
		t.Log(urlStr)
		t.Log(lang)
		t.Log(pos)
		t.Log(fun.Timestamp(true) - start)

	}
}
