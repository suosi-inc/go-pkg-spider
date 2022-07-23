package detect

import (
	"testing"

	"github.com/x-funs/go-fun"
)

func TestGuessHtmlCharsetLang(t *testing.T) {
	var urlStr string

	urlStr = "http://government.ru/"

	resp, _ := fun.HttpGetResp(urlStr, nil, 10000)

	charset, lang := GuessHtmlCharsetLang(resp.Body)

	t.Log(charset)
	t.Log(lang)
}
