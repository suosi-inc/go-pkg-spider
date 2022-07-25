package detect

import (
	"net/http"
	"strings"

	"github.com/suosi-inc/chardet"
	"github.com/x-funs/go-fun"
)

// CharsetLang 解析 HTTP body、http.Header 中的编码和语言, 如果未解析成功则尝试进行猜测
func CharsetLang(body []byte, headers *http.Header) (CharsetRes, LangRes) {
	var charsetRes CharsetRes
	var langRes LangRes

	var guessCharset string
	var guessLang string

	// 根据 Content-Type、Body Html 标签探测编码
	charsetRes = Charset(body, headers)

	// 未识别到 charset 则使用 guess
	if charsetRes.Charset == "" {
		guessCharset, guessLang = GuessHtmlCharsetLang(body)

		if guessCharset != "" {
			charsetRes.Charset = guessCharset
			charsetRes.CharsetPos = CharsetPosGuess
		}

		if guessLang != "" {
			langRes.Lang = guessLang
			langRes.LangPos = LangPosGuess
		}

	} else {
		if strings.HasPrefix(charsetRes.Charset, "iso") || strings.HasPrefix(charsetRes.Charset, "windows") {
			_, guessLang = GuessHtmlCharsetLang(body)
			if guessLang != "" {
				langRes.Lang = guessLang
				langRes.LangPos = LangPosGuess
			}
		}
	}

	// 探测语言
	if langRes.Lang == "" {
		langRes = Lang(body, charsetRes.Charset)
	}

	return charsetRes, langRes
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
