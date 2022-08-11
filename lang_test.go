package spider

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/pemistahl/lingua-go"
	"github.com/x-funs/go-fun"
)

func TestLinguaText(t *testing.T) {
	text := "BEIJING, 10 août (Xinhua) -- Un porte-parole du Bureau du Travail du Comité central du Parti communiste chinois pour les affaires de Taiwan a fait mercredi des remarques sur un livre blanc nouvellement publié intitulé \"La question de Taiwan et la réunification de la Chine dans la nouvelle ère\"."

	start := fun.Timestamp(true)
	languages := []lingua.Language{
		lingua.French,
		lingua.Spanish,
		lingua.Portuguese,
		lingua.German,
	}
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	if language, exists := detector.DetectLanguageOf(text); exists {
		t.Log(text)
		t.Log(language.IsoCode639_1())
		fmt.Println(fun.Timestamp(true) - start)
	}
}

func BenchmarkLinguaTest(b *testing.B) {

	text := "BEIJING"

	languages := []lingua.Language{
		lingua.French,
		lingua.Spanish,
		lingua.Portuguese,
		lingua.German,
		lingua.English,
	}
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectLanguageOf(text)
	}
}

func TestLingua(t *testing.T) {

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
		"http://hanoimoi.com.vn",
		// "https://www.commerce.gov.mm",
		// "https://www.rrdmyanmar.gov.mm",
		// "https://czql.gov.cn/",
	}

	for _, urlStr := range urlStrs {
		resp, _ := HttpGetResp(urlStr, nil, 10000)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))

		doc.Find(DefaultRemoveTags).Remove()

		u, _ := fun.UrlParse(urlStr)

		// 语言
		start := fun.Timestamp(true)
		langRes := Lang(doc, resp.Charset.Charset, u.Hostname())

		t.Log(urlStr)
		t.Log(langRes)
		t.Log(fun.Timestamp(true) - start)
	}

}
