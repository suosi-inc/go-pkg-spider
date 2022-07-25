package detect

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/PuerkitoBio/goquery"
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
	var urlStr string

	// utf-8,
	// urlStr = "http://government.ru/"

	// utf-8,无 (中文)
	// urlStr = "http://www.news.cn"
	// utf-8,无 (英语)
	// urlStr = "https://english.news.cn"
	// utf-8,无 (日语)
	// urlStr = "https://jp.news.cn/"
	// utf-8,无 (韩语)
	// urlStr = "https://kr.news.cn/"
	// utf-8,zh-CN (中文)
	// urlStr = "http://www.qq.com"
	// utf-8,无 (中文)
	urlStr = "https://www.163.com/news/article/HD4PT9KO000189FH.html"
	// utf-8,无 (英语)
	// urlStr = "https://www.bbc.com"

	resp, _ := fun.HttpGetResp(urlStr, nil, 10000)

	charset, lang := CharsetLang(resp.Body, resp.Headers)

	t.Log(charset)
	t.Log(lang)
}

func BenchmarkCharsetLang(b *testing.B) {
	var urlStr string

	// utf-8,zh-CN (中文)
	// urlStr = "http://www.qq.com"
	// utf-8,无 (中文)
	// urlStr = "https://www.163.com/news/article/HD4PT9KO000189FH.html"
	// utf-8,无 (英语)
	urlStr = "https://english.news.cn"

	resp, _ := fun.HttpGetResp(urlStr, nil, 10000)

	// 重制定时器
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CharsetLang(resp.Body, resp.Headers)
	}
}

func BenchmarkLangFromUtf8Body(b *testing.B) {
	var urlStr string

	// utf-8,无 (中文)
	urlStr = "https://www.163.com"

	resp, _ := fun.HttpGetResp(urlStr, nil, 10000)

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
	doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()

	// 重制定时器
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LangFromUtf8Body(doc)
	}
}
