package extract

import (
	"fmt"
	"testing"
)

func TestDomainParse(t *testing.T) {
	domains := []string{
		"www.net.cn",
		"hi.chinanews.com",
		"a.wh.cn",
		"siat.ac.cn",
		"abc.spring.io",
		"abc.spring.ai",
		"www.china-embassy.or.jp",
		"whszdj.wh.cn",
		"gk.wh.cn",
		"xwxc.mwr.cn",
		"legismac.safp.gov.mo",
		"dezhou.rcsd.cn",
		"www.gov.cn",
		"scopsr.gov.cn",
		"usa.gov",
	}

	for _, domain := range domains {
		t.Log(DomainParse(domain))
	}
}

func TestDomainTop(t *testing.T) {
	domains := []string{
		"www.net.cn",
		"hi.chinanews.com",
		"a.wh.cn",
		"siat.ac.cn",
		"abc.spring.io",
		"abc.spring.ai",
		"www.china-embassy.or.jp",
		"whszdj.wh.cn",
		"gk.wh.cn",
		"xwxc.mwr.cn",
		"legismac.safp.gov.mo",
		"dezhou.rcsd.cn",
		"www.gov.cn",
		"scopsr.gov.cn",
		"usa.gov",
	}

	for _, domain := range domains {
		t.Log(DomainTop(domain))
	}
}

func TestDomainTopFromUrl(t *testing.T) {
	fmt.Println(DomainTopFromUrl("https://www.google.com"))
	fmt.Println(DomainTopFromUrl("https://www.baidu.com/news"))
}
