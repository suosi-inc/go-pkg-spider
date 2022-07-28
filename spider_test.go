package spider

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pemistahl/lingua-go"
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
		charset := Charset(resp.Body, resp.Headers)
		t.Log(urlStr)
		t.Log(charset)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()
		lang := Lang(doc, charset.Charset, u.Hostname())
		t.Log(lang)

		t.Log(fun.Timestamp(true) - start)
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

func TestLingua(t *testing.T) {

	var urlStrs = []string{
		"https://www.163.com",
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
		resp, _ := fun.HttpGetResp(urlStr, nil, 10000)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()

		text := doc.Find("a").Text()
		text = strings.ReplaceAll(text, "\n", "")
		text = strings.ReplaceAll(text, "\t", "")
		text = strings.ReplaceAll(text, "  ", "")
		m := regexp.MustCompile(`[\pP\pS]`)
		text = m.ReplaceAllString(text, "")

		text = fun.SubString(text, 0, 1024)

		start := fun.Timestamp(true)
		languages := []lingua.Language{
			lingua.Arabic,
			lingua.Russian,
			lingua.Hindi,
			lingua.Vietnamese,
			lingua.Thai,
		}
		detector := lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			Build()

		if language, exists := detector.DetectLanguageOf(text); exists {
			t.Log(urlStr)
			t.Log(text)
			t.Log(language.IsoCode639_1())
			fmt.Println(fun.Timestamp(true) - start)
		}
	}

}
