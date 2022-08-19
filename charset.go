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
func CharsetFromHeaderHtml(body []byte, headers *http.Header) CharsetRes {
	var res CharsetRes

	cHeader := CharsetFromHeader(headers)

	cHtml := CharsetFromHtml(body)

	// 只有 Header 则使用 Header
	if cHeader != "" && cHtml == "" {
		res.Charset = cHeader
		res.CharsetPos = CharsetPosHeader
		return res
	}

	// 只有 Html 则使用 Html
	if cHeader == "" && cHtml != "" {
		res.Charset = cHtml
		res.CharsetPos = CharsetPosHtml
		return res
	}

	// 同时有 Header 和 Html, 根据情况使用 Header 或 Html
	if cHeader != "" && cHtml != "" {
		if cHeader == cHtml {
			res.Charset = cHeader
			res.CharsetPos = CharsetPosHeader
			return res
		}

		// Header 和 Html 不一致, 以下情况以 Html 为准
		if strings.HasPrefix(cHeader, "ISO") || strings.HasPrefix(cHeader, "WINDOWS") {
			res.Charset = cHtml
			res.CharsetPos = CharsetPosHtml
			return res
		}

		res.Charset = cHeader
		res.CharsetPos = CharsetPosHeader
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
func CharsetFromHtml(body []byte) string {
	var charset string

	if len(body) >= 0 {
		// 先检测 HTML 标签
		html := fun.String(body)

		// 匹配 HTML4 标签
		var charset4 string
		matches := regexCharsetHtml4Pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			matches = regexCharsetPattern.FindStringSubmatch(matches[1])
			if len(matches) > 1 {
				charset4 = matches[1]
			}
		}

		// 匹配 HTML5 标签
		var charset5 string
		matches = regexCharsetHtml5Pattern.FindStringSubmatch(html)
		if len(matches) > 1 {
			charset5 = matches[1]
		}

		// 只有其中一个
		if charset4 != "" && charset5 == "" {
			charset = charset4
		}

		if charset4 == "" && charset5 != "" {
			charset = charset5
		}

		if charset4 != "" && charset5 != "" {
			// 竟然两个都有, 以最先出现的为准
			if charset4 == charset5 {
				charset = charset5
			} else {
				charset4Index := strings.Index(html, charset4)
				charset5Index := strings.Index(html, charset5)

				if charset4Index < charset5Index {
					charset = charset4
				} else {
					charset = charset5
				}
			}

		}
	}

	return convertCharset(charset)
}

// CharsetGuess 根据 HTTP body 猜测编码
func CharsetGuess(body []byte) string {
	var guessCharset string

	// 检测是否是 UTF-8
	valid := utf8.Valid(body)
	if valid {
		return "UTF-8"
	}

	// 如果没有则 guess
	detector := chardet.NewHtmlDetector()
	guess, err := detector.DetectBest(body)
	if err == nil {
		guessCharset = strings.ToUpper(guess.Charset)
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
