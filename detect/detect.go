package detect

import (
	"net/http"
	"strings"

	"github.com/suosi-inc/chardet"
	"github.com/x-funs/go-fun"
)

// CharsetLang 解析 HTTP body、http.Header 中的编码和语言, 如果未解析成功则尝试进行猜测
func CharsetLang(body []byte, headers *http.Header) (string, string) {
	var charset string
	var lang string

	var guessCharset string
	var guessLang string

	// 根据 Content-Type、Body Html 标签探测编码和语言
	// charset = Charset(body, headers)
	lang = Lang(body, charset)

	// 未识别到 charset 则使用 guess
	if charset == "" {
		guessCharset, guessLang = GuessHtmlCharsetLang(body)
		charset = guessCharset

		if (lang == "" || lang == "en") && guessLang != "" {
			lang = guessLang
		}

	} else {
		// if strings.HasPrefix(charset, "iso-8859-") || strings.HasPrefix(charset, "windows-") {
		if lang == "" || lang == "en" {
			guessCharset, guessLang = GuessHtmlCharsetLang(body)
			charset = guessCharset
			// }
		}
	}

	return charset, lang
}

// GuessHtmlCharsetLang 根据 HTTP body 猜测编码和语言 (benchmark 3ms)
func GuessHtmlCharsetLang(body []byte) (string, string) {
	var guessCharset string
	var guessLang string

	detector := chardet.NewHtmlDetector()
	guess, err := detector.DetectBest(body)
	if err == nil {
		guessCharset = strings.ToLower(guess.Charset)
		guessLang = strings.ToLower(guess.Language)
		guessLang = fun.SubString(guessLang, 0, 2)
	}

	return guessCharset, guessLang
}
