package spider

import (
	"net/http"
	"strings"

	"github.com/suosi-inc/chardet"
	"github.com/suosi-inc/go-pkg-spider/detect"
)

// DetectCharset 解析 HTTP body、http.Header 中的编码和语言, 如果未解析成功则尝试进行猜测
func DetectCharset(body []byte, headers *http.Header) detect.CharsetRes {
	var charsetRes detect.CharsetRes
	var guessCharset string

	// 根据 Content-Type、Body Html 标签探测编码
	charsetRes = detect.Charset(body, headers)

	// 未识别到 charset 则使用 guess
	if charsetRes.Charset == "" {
		guessCharset = DetectCharsetGuess(body)

		if guessCharset != "" {
			charsetRes.Charset = guessCharset
			charsetRes.CharsetPos = detect.CharsetPosGuess
		}
	}

	return charsetRes
}

// DetectLang 根据 HTTP body、编码、域名后缀探测语言
func DetectLang(body []byte, charset string, host string) detect.LangRes {
	langRes := detect.Lang(body, charset, host)

	return langRes
}

// DetectCharsetGuess 根据 HTTP body 猜测编码
func DetectCharsetGuess(body []byte) string {
	var guessCharset string

	detector := chardet.NewHtmlDetector()
	guess, err := detector.DetectBest(body)
	if err == nil {
		guessCharset = strings.ToLower(guess.Charset)
	}

	return guessCharset
}
