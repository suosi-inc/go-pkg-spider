package detect

import (
	"bytes"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

var (
	metaLangSelectors = []string{
		"meta[http-equiv=content-language]",
		"meta[name=lang]",
	}

	langMaps = map[string]string{
		"gbk":         "zh", // 中文
		"big5":        "zh", // 中文
		"iso-2022-cn": "zh", // 中文
		"shift_jis":   "ja", // 日语
		"koi8-r":      "ru", // 俄语
		"koi8-u":      "ua", // 乌克兰语
		"euc-jp":      "ja", // 日语
		"euc-kr":      "ko", // 韩语
		"iso-2022-jp": "ja", // 日语
		"iso-2022-kr": "ko", // 韩语
	}
)

const (
	LangPosCharset = "charset"
	LangPosHtml    = "html"
	LangPosBody    = "body"
	LangPosGuess   = "guess"
	BodyChunkSize  = 2048
)

type LangRes struct {
	Lang    string
	LangPos string
}

func Lang(h []byte, charset string) LangRes {
	var res LangRes
	var lang string

	// 如果存在 charset 对照表，则直接返回
	if charset != "" {
		if _, exist := langMaps[charset]; exist {
			res.Lang = langMaps[charset]
			res.LangPos = LangPosCharset
			return res
		}
	}

	// 解析 Html
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(h))
	doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()
	lang = LangFromHtml(doc)
	if lang != "" {
		res.Lang = lang
		res.LangPos = LangPosHtml
		return res
	}

	// 当 utf-8 编码时，lang 为空或 en 不可信，进行少量语种的最后检测
	if charset == "utf-8" && (lang == "" || lang == "en") {
		bodyLang := LangFromUtf8Body(doc)
		if bodyLang != "" {
			res.Lang = bodyLang
			res.LangPos = LangPosBody
			return res
		}
	}

	return res
}

func LangFromHtml(doc *goquery.Document) string {
	var lang string

	// html lang
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

func LangFromUtf8Body(doc *goquery.Document) string {
	var lang string
	var text string

	// 获取网页中最多 128 个 a 标签，如果没有 a 标签，则获取 body
	aTag := doc.Find("a")
	aTagSize := aTag.Size()
	if aTagSize >= 10 {
		sliceMax := fun.Min(aTagSize, 128)
		text = aTag.Slice(0, sliceMax).Text()
	} else {
		text = doc.Find("body").Text()
	}

	// 去除换行、空格、符号
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "\t", "")
	text = strings.ReplaceAll(text, " ", "")
	m := regexp.MustCompile(`[\pP\pS\pZ]`)
	text = m.ReplaceAllString(text, "")

	// 最大截取 BodyChunkSize 个字符
	text = fun.SubString(text, 0, BodyChunkSize)
	textCount := utf8.RuneCountInString(text)

	// 英语占比很高
	englishRegexp := regexp.MustCompile(`[a-zA-Z]`)
	english := englishRegexp.FindAllString(text, -1)
	englishCount := len(english)
	englishRate := float64(englishCount) / float64(textCount)
	if englishRate > 0.8 {
		lang = "en"
		return lang
	}

	// 是否包含汉字
	hanRegex := regexp.MustCompile(`\p{Han}`)
	han := hanRegex.FindAllString(text, -1)
	hanCount := len(han)
	hanRate := float64(hanCount) / float64(textCount)

	// 汉字比例
	if hanRate >= 0.2 {
		jaRegex := regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
		ja := jaRegex.FindAllString(text, -1)
		jaCount := len(ja)
		jaRate := float64(jaCount) / float64(hanCount)

		// 日语占比
		if jaRate > 0.1 {
			lang = "ja"
			return lang
		} else {
			lang = "zh"
			return lang
		}
	} else if englishRate > 0.8 {
		lang = "en"
		return lang
	}

	return lang
}
