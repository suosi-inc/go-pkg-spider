package spider

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestTopDomain(t *testing.T) {
	t.Log(TopDomain("hi.chinanews.com"))
	t.Log(TopDomain("a.wh.cn"))
	t.Log(TopDomain("siat.ac.cn"))
	t.Log(TopDomain("abc.spring.io"))
	t.Log(TopDomain("abc.spring.ai"))
	t.Log(TopDomain("www.china-embassy.or.jp"))
	t.Log(TopDomain("whszdj.wh.cn"))
	t.Log(TopDomain("gk.wh.cn"))
	t.Log(TopDomain("xwxc.mwr.cn"))
	t.Log(TopDomain("legismac.safp.gov.mo"))
	t.Log(TopDomain("dezhou.rcsd.cn"))
	t.Log(TopDomain("www.gov.cn"))
	t.Log(TopDomain("scopsr.gov.cn"))
}

func TestTopDomain2(t *testing.T) {
	domains := []string{
		"spartanswire.usatoday.com",
		"badgerswire.usatoday.com",
		"wolverineswire.usatoday.com",
		"volswire.usatoday.com",
		"buckeyeswire.usatoday.com",
		"ugawire.usatoday.com",
		"guce.yahoo.com",
		"pageviewer.yomiuri.co.jp",
		"partner.buy.yahoo.com",
		"tw.edit.yahoo.com",
		"tw.security.yahoo.com",
		"tw.knowledge.yahoo.com",
		"travel.m.pchome.com.tw",
		"blogs.reuters.com",
		"reuters.com",
		"tw.money.yahoo.com",
		"tw.mobile.yahoo.com",
		"asia.adspecs.yahoo.com",
		"learngerman.dw.com",
		"conference.udn.com",
		"mediadirectory.economist.com",
		"eventsregistration.economist.com",
		"eventscustom.economist.com",
		"technologyforchange.economist.com",
		"sustainabilityregistration.economist.com",
		"learn-french.lemonde.fr",
		"jungeleute.sueddeutsche.de",
		"jetzt.sueddeutsche.de",
		"coupons.cnn.com",
		"www.cnn.com",
		"www.khmer.voanews.com",
		"www.burmese.voanews.com",
		"www.tigrigna.voanews.com",
		"nkpos.nikkei.co.jp",
		"nvs.nikkei.co.jp",
		"simonglazin.dailymail.co.uk",
		"adweb.nikkei.co.jp",
		"broganblog.dailymail.co.uk",
		"pclub.nikkei.co.jp",
		"araward.nikkei.co.jp",
		"blend.nikkei.co.jp",
		"esf.nikkei.co.jp",
		"hoshiaward.nikkei.co.jp",
		"marketing.nikkei.com",
		"www.now.com",
		"jp.wsj.com",
		"subscribenow.economist.com",
		"sportsawards.usatoday.com",
		"cooking.nytimes.com",
	}

	for _, domain := range domains {
		t.Log(TopDomain(domain))
	}
}

func TestTopDomainUrl(t *testing.T) {
	t.Log(TopDomainFromUrl("  "))
	t.Log(TopDomainFromUrl("https://www.google.com"))
	t.Log(TopDomainFromUrl("https://hi.baidu.com/news"))
	t.Log(TopDomainFromUrl("//hi.baidu.com/news"))
}

func BenchmarkTest(b *testing.B) {
	f, err := os.Open(filepath.Join("test/html", "sohu.com.html"))
	if err != nil {
		log.Println("open file error:", err)
	}
	defer f.Close()

	input, _ := io.ReadAll(f)
	log.Println("size:", len(input))
	for i := 0; i < b.N; i++ {
		DetectCharsetGuess(input)
	}
}
