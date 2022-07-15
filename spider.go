package spider

import (
	"github.com/suosi-inc/go-pkg-spider/domain"
)

// TopDomain 返回的顶级域名
func TopDomain(d string) string {
	if d, err := domain.Parse(d); err == nil {
		return d.Domain + "." + d.TLD
	}

	return ""
}

// TopDomainFromUrl 解析 URL 返回顶级域名
func TopDomainFromUrl(urlStr string) string {
	if d, err := domain.ParseFromUrl(urlStr); err == nil {
		return d.Domain + "." + d.TLD
	}

	return ""
}
