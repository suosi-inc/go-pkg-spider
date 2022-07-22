package spider

import (
	"net/http"

	"github.com/suosi-inc/go-pkg-spider/detect"
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

func CharsetDetect(body []byte, headers *http.Header) string {
	c := detect.CharsetFromHeader(headers)
	if c != "" {
		return c
	}

	c = detect.CharsetFromHeader(headers)
	if c != "" {
		return c
	}

	return c
}

func LangDetect(body []byte) string {
	return ""
}
