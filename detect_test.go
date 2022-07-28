package spider

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/detect"
	"github.com/x-funs/go-fun"
)

func TestCharsetLang(t *testing.T) {
	var urlStrs = []string{
		"http://suosi.com.cn",
		"https://www.163.com",
		"https://english.news.cn",
		"https://jp.news.cn",
		"https://kr.news.cn",
		"https://www.donga.com/",
		"http://www.koreatimes.com/",
		"https://arabic.news.cn",
		"https://www.bbc.com",
		"http://government.ru",
		"https://french.news.cn",
		"https://www.gouvernement.fr",
		"http://live.siammedia.org/",
		"http://hanoimoi.com.vn",
		"https://www.commerce.gov.mm",
		"https://sanmarg.in/",
		"https://www.rrdmyanmar.gov.mm",
	}

	for _, urlStr := range urlStrs {
		resp, _ := fun.HttpGetResp(urlStr, nil, 30000)

		u, _ := url.Parse(urlStr)

		start := fun.Timestamp(true)
		charset := DetectCharset(resp.Body, resp.Headers)
		t.Log(urlStr)
		t.Log(charset)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()
		lang := DetectLang(doc, charset.Charset, u.Hostname())
		t.Log(lang)

		t.Log(fun.Timestamp(true) - start)
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
		"https://french.news.cn",
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
		doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()

		start := fun.Timestamp(true)
		lang, pos := detect.LangFromUtf8Body(doc, u.Hostname())
		t.Log(urlStr)
		t.Log(lang)
		t.Log(pos)
		t.Log(fun.Timestamp(true) - start)

	}
}

func TestDetectIcp(t *testing.T) {
	var urlStrs = []string{
		"http://suosi.com.cn",
		"https://www.163.com",
		"https://www.sohu.com",
		"https://www.qq.com",
		"https://www.hexun.com",
		"https://www.wfmc.edu.cn/",
		"https://www.cankaoxiaoxi.com/",
	}

	for _, urlStr := range urlStrs {

		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(err)
		t.Log(urlStr)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find("noscript,style,iframe,br,link,svg").Remove()
		icp, loc := detect.Icp(doc)
		t.Log(icp, loc)
	}
}
