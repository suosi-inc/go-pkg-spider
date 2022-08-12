package spider

import (
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/suosi-inc/chardet"
	"github.com/x-funs/go-fun"
)

const (
	CharsetPosHeader = "header"
	CharsetPosHtml   = "html"
	CharsetPosGuess  = "guess"
)

const (
	RegexCharset      = "(?i)charset=\\s*([a-z][_\\-0-9a-z]*)"
	RegexCharsetHtml4 = "(?i)<meta\\s+([^>]*http-equiv=(\"|')?content-type(\"|')?[^>]*)>"
	RegexCharsetHtml5 = "(?i)<meta\\s+charset\\s*=\\s*[\"']?([a-z][_\\-0-9a-z]*)[^>]*>"
)

var (
	regexCharsetPattern      = regexp.MustCompile(RegexCharset)
	regexCharsetHtml4Pattern = regexp.MustCompile(RegexCharsetHtml4)
	regexCharsetHtml5Pattern = regexp.MustCompile(RegexCharsetHtml5)
)

type CharsetRes struct {
	Charset    string
	CharsetPos string
}

// Charset 解析 HTTP body、http.Header 中的编码和语言, 如果未解析成功则尝试进行猜测
func Charset(body []byte, headers *http.Header) CharsetRes {
	var charsetRes CharsetRes
	var guessCharset string

	// 根据 Content-Type、Body Html 标签探测编码
	charsetRes = CharsetFromHeaderHtml(body, headers)

	// 未识别到 charset 则使用 guess
	if charsetRes.Charset == "" {
		guessCharset = CharsetGuess(body)

		if guessCharset != "" {
			charsetRes.Charset = guessCharset
			charsetRes.CharsetPos = CharsetPosGuess
		}
	}

	return charsetRes
}

// CharsetFromHeaderHtml 解析 HTTP body、http.Header 中的 charset, 准确性高
func CharsetFromHeaderHtml(h []byte, headers *http.Header) CharsetRes {
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
			matches := regexCharsetPattern.FindStringSubmatch(contentType)
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
			return "UTF-8"
		}

		// 检测 HTML 标签
		html := fun.String(h)

		// 优先判断 HTML4 标签
		matches := regexCharsetHtml4Pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			matches = regexCharsetPattern.FindStringSubmatch(matches[1])
			if len(matches) > 1 {
				charset = matches[1]
			}
		}

		// HTML4 标签
		if charset == "" {
			matches = regexCharsetHtml5Pattern.FindStringSubmatch(html)
			if len(matches) > 1 {
				charset = matches[1]
			}

		}
	}

	return convertCharset(charset)
}

// CharsetGuess 根据 HTTP body 猜测编码
func CharsetGuess(body []byte) string {
	var guessCharset string

	detector := chardet.NewHtmlDetector()
	guess, err := detector.DetectBest(body)
	if err == nil {
		guessCharset = strings.ToLower(guess.Charset)
	}

	return guessCharset
}

// convertCharset 格式化 charset
func convertCharset(charset string) string {
	c := strings.ToUpper(strings.TrimSpace(charset))

	if c != "" {
		// alias utf8
		if c == "UTF8" || c == "UTF_8" {
			return "UTF-8"
		}

		// alias gb2312, gb18030
		if strings.HasPrefix(c, "GB") {
			return "GBK"
		}

		// alias big5-hkscs..
		if strings.HasPrefix(c, "BIG5") {
			return "Big5"
		}

		// alias shift-jis
		if strings.HasPrefix(c, "SHIFT") {
			return "SHIFT_JIS"
		}
	}

	return c
}
