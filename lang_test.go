package spider

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pemistahl/lingua-go"
	"github.com/x-funs/go-fun"
)

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

		text := doc.Find("a").Text()
		text = fun.RemoveLines(text)
		text = strings.ReplaceAll(text, fun.TAB, "")
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
