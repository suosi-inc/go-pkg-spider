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
		lang := DetectLang(resp.Body, charset.Charset, u.Hostname())
		t.Log(urlStr)
		t.Log(charset)
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

func BenchmarkCharsetLang(b *testing.B) {
	// var urlStr string

	// utf-8,zh-CN (中文)
	// urlStr = "http://www.qq.com"
	// utf-8,无 (中文)
	// urlStr = "https://www.163.com/news/article/HD4PT9KO000189FH.html"
	// utf-8,无 (英语)
	// urlStr = "https://english.news.cn"

	// resp, _ := fun.HttpGetResp(urlStr, nil, 10000)
	//
	// // 重制定时器
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	_, _ = DetectCharset(resp.Body, resp.Headers)
	// }
}

func BenchmarkLangFromUtf8Body(b *testing.B) {
	// var urlStr string
	//
	// // utf-8,无 (中文)
	// // urlStr = "https://www.163.com"
	// urlStr = "https://english.news.cn"
	//
	// resp, _ := fun.HttpGetResp(urlStr, nil, 10000)
	//
	// doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	// doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()
	//
	// // 重制定时器
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	_ = LangFromUtf8Body(doc)
	// }
}
