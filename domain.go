package spider

import (
	"errors"
	"net/url"
	"strings"

	"github.com/x-funs/go-fun"
	"golang.org/x/net/publicsuffix"
)

type Domain struct {
	Subdomain, Domain, TLD string
	ICANN                  bool
}

// TopDomain 返回的顶级域名
func TopDomain(d string) string {
	if d, err := DomainParse(d); err == nil {
		return d.Domain + "." + d.TLD
	}

	return ""
}

// TopDomainFromUrl 解析 URL 返回顶级域名
func TopDomainFromUrl(urlStr string) string {
	if d, err := DomainParseFromUrl(urlStr); err == nil {
		return d.Domain + "." + d.TLD
	}

	return ""
}

func DomainParse(domain string) (*Domain, error) {
	if fun.Blank(domain) {
		return nil, errors.New("domain is blank")
	}

	// etld+1
	etld1, err := publicsuffix.EffectiveTLDPlusOne(domain)
	_, icann := publicsuffix.PublicSuffix(strings.ToLower(domain))
	if err != nil {
		return nil, err
	}

	// convert to domain name, and tld
	i := strings.Index(etld1, ".")
	domName := etld1[0:i]
	tld := etld1[i+1:]

	// and subdomain
	sub := ""
	if rest := strings.TrimSuffix(domain, "."+etld1); rest != domain {
		sub = rest
	}
	return &Domain{
		Subdomain: sub,
		Domain:    domName,
		TLD:       tld,
		ICANN:     icann,
	}, nil
}

func DomainParseFromUrl(urlStr string) (*Domain, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.New("url parse error")
	}

	d := u.Hostname()

	return DomainParse(d)
}
