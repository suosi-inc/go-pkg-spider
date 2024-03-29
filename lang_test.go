package spider

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/lingua-go"
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

func TestLang(t *testing.T) {

	var urlStrs = []string{

		"https://www.bbc.com",
		"https://www.ft.com/",

		"https://www.163.com/news/article/HEJGEVFT000189FH.html",
		"https://www.163.com",

		"https://english.news.cn",
		"https://jp.news.cn",
		"https://kr.news.cn",
		"https://german.news.cn/",
		"https://portuguese.news.cn/",
		"https://arabic.news.cn",
		"https://french.news.cn",

		"https://mn.cctv.com/",

		"http://government.ru",

		"https://www.gouvernement.fr",

		"http://live.siammedia.org/",
		"https://www.manchestereveningnews.co.uk/",

		"https://www.chinadaily.com.cn",
		"http://cn.chinadaily.com.cn/",
		"http://www.chinadaily.com.cn/chinawatch_fr/index.html",
		"https://d1ev.com/",
		"https://www.cngold.com.cn/",
		"https://china.guidechem.com/",
		"https://xdkb.net/",
		"https://www.lifeweek.com.cn/",
		"http://gxbsrd.gov.cn/",
		"https://defence24.com/",
		"http://www.gmp.or.kr/",
		"http://rdfmj.com/",
		"https://news.xmnn.cn/xmnn/2022/08/09/101067908.shtml",
	}

	for _, urlStr := range urlStrs {
		resp, _ := HttpGetResp(urlStr, nil, 10000)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))

		doc.Find(DefaultDocRemoveTags).Remove()

		// 语言
		start := fun.Timestamp(true)
		langRes := Lang(doc, resp.Charset.Charset, true)

		t.Log(urlStr)
		t.Log(resp.Charset)
		t.Log(langRes)
		t.Log(fun.Timestamp(true) - start)
	}

}

func TestLangText(t *testing.T) {
	start := fun.Timestamp(true)
	text := "中文"
	t.Log(fun.Timestamp(true) - start)
	t.Log(LangText(text))
}

func TestUnicode(t *testing.T) {
	text := "BEIJING, 9. August 2022 (Xinhuanet) -- In einem am Dienstag veröffentlichten Bericht über die Menschenrechtsverletzungen der USA wird darauf hingewiesen, dass die Vereinigten Staaten einen \"Konflikt der Zivilisationen\" geschaffen, Haft und Folter missbraucht sowie die Religionsfreiheit und Menschenwürde verletzt hätten.\n\nDer Bericht mit dem Titel ''Die USA begehen schwerwiegende Verbrechen der Menschenrechtsverletzungen im Nahen Osten und darüber hinaus'' wurde von der Chinesischen Gesellschaft für Menschenrechtsstudien veröffentlicht.\n\nIn dem Bericht heißt es, dass die Vereinigten Staaten keinen Respekt vor der Diversität der Zivilisationen zeigten, der islamischen Zivilisation feindlich gegenüberständen, das historische und kulturelle Erbe des Nahen Ostens zerstörten, Muslime rücksichtslos inhaftierten und folterten und die grundlegenden Menschenrechte der Bevölkerung im Nahen Osten und in anderen Gebieten schwer verletzten.\n\n\"Die Vereinigten Staaten haben die 'islamische Bedrohungstheorie' in der ganzen Welt verbreitet. Sie haben die Überlegenheit der westlichen und christlichen Zivilisation befürwortet, die nicht-westliche Zivilisation verachtet und die islamische Zivilisation stigmatisiert, indem sie sie als 'rückständig', 'terroristisch' und 'gewalttätig' bezeichneten\", heißt es in dem Bericht."
	// latinRex := regexp.MustCompile(`\p{Lo}`)
	latinRex := regexp.MustCompile("[\u0080-\u00ff]")
	latin := latinRex.FindAllString(text, -1)

	t.Log(latin)
}
