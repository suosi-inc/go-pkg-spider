package spider

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/pemistahl/lingua-go"
	"github.com/x-funs/go-fun"
)

var (
	CharsetLangMap = map[string]string{
		"GBK":         "zh",
		"Big5":        "zh",
		"ISO-2022-CN": "zh",
		"SHIFT_JIS":   "ja",
		"KOI8-R":      "ru",
		"EUC-JP":      "ja",
		"EUC-KR":      "ko",
		"EUC-CN":      "zh",
		"ISO-2022-JP": "ja",
		"ISO-2022-KR": "ko",
	}

	LangEnZhMap = map[string]string{
		"zh": "中文",
		"en": "英语",
		"ja": "日语",
		"ru": "俄语",
		"ko": "韩语",
		"ar": "阿拉伯语",
		"hi": "印地语",
		"th": "泰语",
		"vi": "越南语",
		"de": "德语",
		"fr": "法语",
		"it": "意大利语",
		"es": "西班牙语",
		"pt": "葡萄牙语",
	}

	LangZhEnMap = map[string]string{
		"中文":   "zh",
		"英语":   "en",
		"日语":   "ja",
		"俄语":   "ru",
		"韩语":   "ko",
		"阿拉伯语": "ar",
		"印地语":  "hi",
		"泰语":   "th",
		"越南语":  "vi",
		"德语":   "de",
		"法语":   "fr",
		"意大利语": "it",
		"西班牙语": "es",
		"葡萄牙语": "pt",
	}

	metaLangSelectors = []string{
		"meta[http-equiv=content-language]",
		"meta[name=lang]",
	}

	linguaLanguages = []lingua.Language{
		lingua.Arabic,
		lingua.Russian,
		lingua.Hindi,
		lingua.Vietnamese,
		lingua.Thai,
		lingua.Korean,
	}

	linguaMap = map[string]string{
		"arabic":     "ar",
		"russian":    "ru",
		"hindi":      "hi",
		"vietnamese": "vi",
		"thai":       "th",
		"korean":     "ko",
	}
)

const (
	LangPosCharset = "charset"
	LangPosHtml    = "html"
	LangPosBody    = "body"
	LangPosLingua  = "lingua"
	LangPosHost    = "host"
	BodyChunkSize  = 1024
)

type LangRes struct {
	Lang    string
	LangPos string
}

func Lang(doc *goquery.Document, charset string, host string) LangRes {
	var res LangRes
	var lang string

	// 如果存在特定语言的 charset 对照表，则直接返回
	if charset != "" {
		if _, exist := CharsetLangMap[charset]; exist {
			res.Lang = CharsetLangMap[charset]
			res.LangPos = LangPosCharset
			return res
		}
	}

	// 解析 Html 语言属性，当不为空或 en 时可信度高，直接返回
	lang = LangFromHtml(doc)
	if lang != "" && lang != "en" {
		res.Lang = lang
		res.LangPos = LangPosHtml
		return res
	}

	// 当 utf 编码时，lang 为空或 en 不可信，进行基于内容、域名的语种的检测
	if strings.HasPrefix(charset, "UTF") && (lang == "" || lang == "en") {
		bodyLang, pos := LangFromUtf8Body(doc, host)
		if bodyLang != "" {
			res.Lang = bodyLang
			res.LangPos = pos
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

func LangFromUtf8Body(doc *goquery.Document, host string) (string, string) {
	var lang string
	var text string

	// 获取网页中最多 128 个 a 标签，如果没有 a 标签或过少，则获取 body
	aTag := doc.Find("a")
	aTagSize := aTag.Size()
	if aTagSize >= 16 {
		sliceMax := fun.Min(aTagSize, 128)
		text = aTag.Slice(0, sliceMax).Text()
	} else {
		text = doc.Find("body").Text()
	}

	// 去除换行、符号(为了保留语义只替换多余的空格)
	text = fun.RemoveLines(text)
	text = strings.ReplaceAll(text, fun.TAB, "")
	text = strings.ReplaceAll(text, "  ", "")

	m := regexp.MustCompile(`[\pP\pS]`)
	text = m.ReplaceAllString(text, "")

	// 最大截取 BodyChunkSize 个字符
	text = fun.SubString(text, 0, BodyChunkSize)
	textCount := utf8.RuneCountInString(text)

	// 是否包含汉字
	hanRegex := regexp.MustCompile(`\p{Han}`)
	han := hanRegex.FindAllString(text, -1)
	if han != nil {
		hanCount := len(han)
		hanRate := float64(hanCount) / float64(textCount)

		// 汉字比例
		if hanRate >= 0.3 {
			jaRegex := regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
			ja := jaRegex.FindAllString(text, -1)
			if ja != nil {
				jaCount := len(ja)
				jaRate := float64(jaCount) / float64(hanCount)

				// 日语占比
				if jaRate > 0.1 {
					lang = "ja"
					return lang, LangPosBody
				}
			}

			lang = "zh"
			return lang, LangPosBody
		}
	}

	// 英语占比很高, 不一定是英语, 可能是拉丁语系, 妥协的办法进行域名后缀再判定(不一定准确)
	englishRegexp := regexp.MustCompile(`[a-zA-Z]`)
	english := englishRegexp.FindAllString(text, -1)
	if english != nil {
		englishCount := len(english)
		englishRate := float64(englishCount) / float64(textCount)
		if englishRate > 0.7 {
			hostLang := LangFromHost(host)
			if hostLang != "" {
				return hostLang, LangPosHost
			}

			return "en", LangPosBody
		}
	}

	// 不是英、中、日，尝试小语种的域名特征
	lang = LangFromHost(host)
	if lang != "" {
		return lang, LangPosHost
	}

	// 域名没有特征，最后使用 lingua 分析指定小语种
	detector := lingua.NewLanguageDetectorBuilder().FromLanguages(linguaLanguages...).Build()
	if language, exists := detector.DetectLanguageOf(text); exists {

		key := strings.ToLower(language.String())
		linguaLang := linguaMap[key]
		return linguaLang, LangPosLingua
	}

	return lang, ""
}

func LangFromHost(host string) string {
	var lang string

	host = strings.ToLower(host)
	if strings.HasSuffix(host, ".fr") {
		lang = "fr"
	} else if strings.HasSuffix(host, ".de") {
		lang = "de"
	} else if strings.HasSuffix(host, ".it") {
		lang = "it"
	} else if strings.HasSuffix(host, ".es") {
		lang = "es"
	} else if strings.HasSuffix(host, ".pt") {
		lang = "pt"
	} else if strings.HasSuffix(host, ".in") {
		lang = "in"
	} else if strings.HasSuffix(host, ".vn") {
		lang = "vi"
	} else if strings.HasSuffix(host, ".mm") {
		lang = "my"
	} else if strings.HasSuffix(host, ".th") {
		lang = "th"
	}

	return lang
}
