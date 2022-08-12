package spider

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/extract"
	"github.com/x-funs/go-fun"
)

func TestDomainDetect(t *testing.T) {
	domains := []string{
		// "china-nengyuan.com",
		// "suosi.com.cn",
		"srzc.com",
	}

	for _, domain := range domains {
		domainRes, err := DetectDomain(domain, 10000, 1)
		if err == nil {
			t.Log(domainRes)
		} else {
			t.Log(err)
			t.Log(domainRes)
		}
	}
}

func BenchmarkLinkTitles(b *testing.B) {
	urlStr := "http://www.qq.com/"

	resp, _ := HttpGetResp(urlStr, nil, 30000)

	// 解析 HTML
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	doc.Find(DefaultDocRemoveTags).Remove()

	// 语言

	langRes := Lang(doc, resp.Charset.Charset, true)

	fmt.Println(langRes)

	var linkTitles map[string]string

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 标题
		linkTitles, _ = extract.WebLinkTitles(doc, resp.RequestURL, true)

		// 连接和子域名
		_, _ = extract.LinkTypes(linkTitles, langRes.Lang, nil)

		// rules := map[string][]string{
		// 	"163.com": []string{
		// 		"`\\w{16}\\.html`",
		// 	},
		// }
		// _, _ = extract.LinkTypes(linkTitles, langRes.Lang, rules)
	}

	b.StopTimer()

	fmt.Println(langRes.Lang)
	fmt.Println(len(linkTitles))

}

func TestLinkTitles(t *testing.T) {
	var urlStrs = []string{
		"https://www.qq.com",
		// "https://www.people.com.cn",
		// "https://www.36kr.com",
		// "https://www.163.com",
		// "https://news.163.com/",
		// "http://jyj.suqian.gov.cn",
		// "https://www.huxiu.com/",
		// "http://www.news.cn/politicspro/",
		// "http://www.cankaoxiaoxi.com",
		// "http://www.bbc.com",
		// "https://www.ft.com",
		// "https://www.reuters.com/",
		// "https://nypost.com/",
		// "http://www.mengcheng.gov.cn/",
		// "https://www.chunichi.co.jp",
		// "https://www.donga.com/",
		// "https://people.com/",
		// "https://czql.gov.cn/",
		// "https://qiye.163.com/",
		// "https://www.washingtontimes.com/",
		// "https://www.gamersky.com/",
	}

	for _, urlStr := range urlStrs {

		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(urlStr)
		t.Log(err)

		// 解析 HTML
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultDocRemoveTags).Remove()

		// 语言
		langRes := Lang(doc, resp.Charset.Charset, true)

		fmt.Println(resp.Charset)
		fmt.Println(langRes)

		// 标题
		linkTitles, filters := extract.WebLinkTitles(doc, resp.RequestURL, true)

		// 分类链接和子域名列表
		linkRes, domainRes := extract.LinkTypes(linkTitles, langRes.Lang, nil)

		// 分类链接和子域名列表, 规则
		// rules := map[string][]string{
		// 	"cankaoxiaoxi.com": []string{
		// 		"\\d{7}\\.shtml$",
		// 	},
		// }
		// linkRes, domainRes := extract.LinkTypes(linkTitles, langRes.Lang, rules)

		fmt.Println("all:", len(linkTitles))
		fmt.Println("content:", len(linkRes.Content))
		fmt.Println("list:", len(linkRes.List))
		fmt.Println("unknown:", len(linkRes.Unknown))
		fmt.Println("none:", len(linkRes.None))

		i := 0
		for a, title := range filters {
			i = i + 1
			fmt.Println(i, "filter:"+a+"\t=>\t"+title)
		}
		i = 0
		for subdomain := range domainRes {
			i = i + 1
			fmt.Println(i, "domain:"+subdomain)
		}
		i = 0
		for a, title := range linkRes.Content {
			i = i + 1
			fmt.Println(i, "content:"+a+"\t=>\t"+title)
		}
		i = 0
		for a, title := range linkRes.Unknown {
			i = i + 1
			fmt.Println(i, "unknown:"+a+"\t=>\t"+title)
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
		doc.Find(DefaultDocRemoveTags).Remove()
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
		"https://czql.gov.cn/",
	}

	for _, urlStr := range urlStrs {
		resp, _ := fun.HttpGetResp(urlStr, nil, 30000)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultDocRemoveTags).Remove()

		start := fun.Timestamp(true)
		lang, pos := LangFromUtf8Body(doc, false)
		t.Log(urlStr)
		t.Log(lang)
		t.Log(pos)
		t.Log(fun.Timestamp(true) - start)

	}
}
