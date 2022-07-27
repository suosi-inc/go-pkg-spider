package spider

import (
	"bytes"
	"net/url"
	"regexp"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/detect"
	"github.com/x-funs/go-fun"
)

func TestRegex(t *testing.T) {
	str := ",.!，，D_NAME。！；‘’”“《》**dfs#%^&()-+我1431221     中国123漢字かどうかのjavaを<決定>$¥"
	r := regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
	s := r.FindAllString(str, -1)
	t.Log(str)
	t.Log(s)
}

func TestCharsetLang(t *testing.T) {
	var urlStrs = []string{
		"https://www.163.com",
		"https://english.news.cn",
		"https://jp.news.cn",
		"https://kr.news.cn",
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
		resp, _ := fun.HttpGetResp(urlStr, nil, 10000)

		u, _ := url.Parse(urlStr)

		start := fun.Timestamp(true)
		charset, lang := CharsetLang(resp.Body, resp.Headers, u.Hostname())
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
		resp, _ := fun.HttpGetResp(urlStr, nil, 10000)
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
	// 	_, _ = CharsetLang(resp.Body, resp.Headers)
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
