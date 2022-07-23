package detect

import (
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

const (
	RegexCharsetPattern   = "(?i)charset=\\s*([a-z][_\\-0-9a-z]*)"
	RegexMetaPattern      = "(?i)<meta\\s+([^>]*http-equiv=(\"|')?content-type(\"|')?[^>]*)>"
	RegexMetaHtml5Pattern = "(?i)<meta\\s+charset\\s*=\\s*[\"']?([a-z][_\\-0-9a-z]*)[^>]*>"
)

// Charset 解析 HTTP body、http.Header 中的 charset
func Charset(body []byte, headers *http.Header) string {
	var c string
	c = CharsetFromHeader(headers)
	if c != "" {
		return c
	}

	c = CharsetFromBody(body)
	if c != "" {
		return c
	}

	return c
}

// CharsetFromHeader 解析 HTTP header 中的 charset
func CharsetFromHeader(headers *http.Header) string {
	var charset string
	if headers != nil {
		contentType := headers.Get("Content-Type")
		if !fun.Blank(contentType) {
			matches := regexp.MustCompile(RegexCharsetPattern).FindStringSubmatch(contentType)
			if len(matches) > 1 {
				charset = matches[1]
			}
		}
	}

	return formatCharset(charset)
}

// CharsetFromBody 解析 HTTP body 中的 charset
func CharsetFromBody(body []byte) string {
	var charset string

	if len(body) >= 0 {
		// 监测是否是 UTF-8
		valid := utf8.Valid(body)
		if valid {
			return "utf-8"
		}

		// 监测 HTML 标签
		html := fun.String(body)
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
	}

	return formatCharset(charset)
}

// formatCharset 格式化 charset
func formatCharset(charset string) string {
	c := strings.ToLower(strings.TrimSpace(charset))

	if c != "" {
		if c == "utf8" {
			return "utf-8"
		}

		if strings.HasPrefix(c, "gb") {
			return "gb18030"
		}
	}

	return c
}
