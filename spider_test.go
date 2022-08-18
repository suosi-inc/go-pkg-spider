package spider

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

func BenchmarkHtmlParse(b *testing.B) {

	resp, _ := fun.HttpGetResp("https://www.163.com", nil, 30000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultDocRemoveTags).Remove()
	}
}

func TestGoquery(t *testing.T) {
	body, _ := HttpGet("https://jp.news.cn/index.htm")
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	// lang, exist := doc.Find("html").Attr("id")

	doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()
	text := doc.Find("body").Text()
	text = fun.RemoveSign(text)

	fmt.Println(text)
}

func TestRegex(t *testing.T) {
	str := ",.!，，D_NAME。！；‘’”“《》**dfs#%^&()-+我1431221     中国123漢字かどうかのjavaを<決定>$¥"
	r := regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
	s := r.FindAllString(str, -1)
	t.Log(str)
	t.Log(s)
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

func TestCount(t *testing.T) {
	fmt.Println(regexLangHtmlPattern.MatchString("zh"))
	fmt.Println(regexLangHtmlPattern.MatchString("en"))
	fmt.Println(regexLangHtmlPattern.MatchString("zh-cn"))
	fmt.Println(regexLangHtmlPattern.MatchString("utf-8"))

	fmt.Println(utf8.RuneCountInString("https://khmers.cn/2022/05/23/%e6%b4%aa%e6%a3%ae%e6%80%bb%e7%90%86%ef%bc%9a%e6%9f%ac%e5%9f%94%e5%af%a8%e7%b4%af%e8%ae%a1%e8%8e%b7%e5%be%97%e8%b6%85%e8%bf%875200%e4%b8%87%e5%89%82%e6%96%b0%e5%86%a0%e7%96%ab%e8%8b%97%ef%bc%8c/"))
}
