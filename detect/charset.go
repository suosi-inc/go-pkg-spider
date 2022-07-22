package detect

import (
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

const (
	RegexMetaPattern      = "(?i)<meta\\s+([^>]*http-equiv=(\"|')?content-type(\"|')?[^>]*)>"
	RegexMetaHtml5Pattern = "(?i)<meta\\s+charset\\s*=\\s*[\"']?([a-z][_\\-0-9a-z]*)[^>]*>"
	RegexCharsetPattern   = "(?i)charset=\\s*([a-z][_\\-0-9a-z]*)"
)

// CharsetFromHeader 解析 HTTP header 中的 charset
func CharsetFromHeader(headers *http.Header) string {
	var charset string
	contentType := headers.Get("Content-Type")
	if !fun.Blank(contentType) {
		matches := regexp.MustCompile(RegexCharsetPattern).FindStringSubmatch(contentType)
		if len(matches) > 1 {
			charset = matches[1]
		}
	}

	return formatCharset(charset)
}

func CharsetFromBody(body []byte) string {
	valid := utf8.Valid(body)
	if valid {
		return "utf-8"
	}

	var charset string
	html := fun.BytesToString(body)

	matches := regexp.MustCompile(RegexMetaPattern).FindStringSubmatch(html)
	if len(matches) > 1 {
		matches = regexp.MustCompile(RegexCharsetPattern).FindStringSubmatch(matches[1])
		if len(matches) > 1 {
			charset = matches[1]
		}
	}

	if charset == "" {
		matches = regexp.MustCompile(RegexMetaHtml5Pattern).FindStringSubmatch(html)
		if len(matches) > 1 {
			charset = matches[1]
		}
	}

	return formatCharset(charset)
}

func formatCharset(charset string) string {
	c := strings.ToLower(strings.TrimSpace(charset))

	if c == "utf8" {
		return "utf-8"
	}

	if strings.HasPrefix(c, "gb") {
		return "gbk"
	}

	return c
}
