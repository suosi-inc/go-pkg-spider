package spider

import (
	"fmt"
	"net/url"
	"testing"
)

func TestTopDomain(t *testing.T) {
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
		t.Log(DomainTop(domain))
	}
}

func TestUrlParse(t *testing.T) {
	u, _ := url.Parse("https://www.test.com:8080/a")
	fmt.Println(u.Host)
	fmt.Println(u.Hostname())
}

func TestDomainParse(t *testing.T) {
	fmt.Println(DomainParse("https://www.google.com"))
}
