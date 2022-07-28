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

const (
	CharsetPosHeader = "header"
	CharsetPosHtml   = "html"
	CharsetPosGuess  = "guess"
)

type CharsetRes struct {
	Charset    string
	CharsetPos string
}

// Charset 解析 HTTP body、http.Header 中的 charset, 准确性高
func Charset(h []byte, headers *http.Header) CharsetRes {
	var res CharsetRes
	var c string

	c = CharsetFromHeader(headers)
	if c != "" {
		res.Charset = c
		res.CharsetPos = CharsetPosHeader
		return res
	}

	c = CharsetFromHtml(h)
	if c != "" {
		res.Charset = c
		res.CharsetPos = CharsetPosHtml
		return res
	}

	return res
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

	return convertCharset(charset)
}

// CharsetFromHtml 解析 Html 中的 charset
func CharsetFromHtml(h []byte) string {
	var charset string

	if len(h) >= 0 {
		// 检测是否是 UTF-8
		valid := utf8.Valid(h)
		if valid {
			return "utf-8"
		}

		// 检测 HTML 标签
		html := fun.String(h)

		// 优先判断 HTML5 标签
		matches := regexp.MustCompile(RegexMetaHtml5Pattern).FindStringSubmatch(html)
		if len(matches) > 1 {
			charset = matches[1]
		}

		// HTML5 标签
		if charset == "" {
			matches = regexp.MustCompile(RegexMetaPattern).FindStringSubmatch(html)
			if len(matches) > 1 {
				matches = regexp.MustCompile(RegexCharsetPattern).FindStringSubmatch(matches[1])
				if len(matches) > 1 {
					charset = matches[1]
				}
			}
		}
	}

	return convertCharset(charset)
}

// convertCharset 格式化 charset
func convertCharset(charset string) string {
	c := strings.ToLower(strings.TrimSpace(charset))

	if c != "" {
		// alias utf8, utf-16..
		if strings.HasPrefix(c, "utf") {
			return "utf-8"
		}

		// alias gb2312 gb18030..
		if strings.HasPrefix(c, "gb") {
			return "gbk"
		}

		// alias big5-hkscs..
		if strings.HasPrefix(c, "big5") {
			return "big5"
		}

		// alias shift-jis..
		if strings.HasPrefix(c, "shift") {
			return "shift_jis"
		}
	}

	return c
}
