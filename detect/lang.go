package detect

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

var (
	metaLangSelectors = []string{
		"meta[http-equiv=content-language]",
		"meta[name=lang]",
	}

	langMaps = map[string]string{
		"gb18030":     "zh",
		"big5":        "zh",
		"shift_jis":   "ja",
		"koi8-r":      "ru",
		"euc-jp":      "ja",
		"euc-kr":      "ko",
		"iso-2022-jp": "ja",
		"iso-2022-kr": "ko",
		"iso-2022-cn": "zh",
	}
)

func Lang(body []byte, charset string) string {
	var lang string

	// 如果存在 charset，则直接返回
	if charset != "" {
		if _, exist := langMaps[charset]; exist {
			return langMaps[charset]
		}
	}

	// html lang
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if lang, exists := doc.Find("html").Attr("lang"); exists {
		lang = fun.SubString(lang, 0, 2)
		return lang
	}
	if lang, exists := doc.Find("html").Attr("xml:lang"); exists {
		lang = fun.SubString(lang, 0, 2)
		return lang
	}
	for _, selector := range metaLangSelectors {
		if lang, exists := doc.Find(selector).Attr("content"); exists {
			lang = fun.SubString(lang, 0, 2)
			return lang
		}
	}

	return lang
}
