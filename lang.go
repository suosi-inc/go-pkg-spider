package spider

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/pemistahl/lingua-go"
	"github.com/suosi-inc/go-pkg-spider/extract"
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
		"de": "德语",
		"fr": "法语",
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
		"德语":   "de",
		"法语":   "fr",
		"西班牙语": "es",
		"葡萄牙语": "pt",
	}

	metaLangSelectors = []string{
		"meta[http-equiv='content-language' i]",
		"meta[name='lang' i]",
	}

	linguaLanguages = []lingua.Language{
		lingua.Arabic,
		lingua.Russian,
		lingua.Hindi,
		lingua.Korean,
	}

	linguaLatinLanguages = []lingua.Language{
		lingua.French,
		lingua.German,
		lingua.Spanish,
		lingua.Portuguese,
		lingua.English,
	}

	linguaMap = map[string]string{
		"arabic":     "ar",
		"russian":    "ru",
		"hindi":      "hi",
		"korean":     "ko",
		"french":     "fr",
		"german":     "de",
		"spanish":    "es",
		"portuguese": "pt",
		"english":    "en",
	}
)

const (
	LangPosCharset = "charset"
	LangPosHtmlTag = "html"
	LangPosBody    = "body"
	LangPosLingua  = "lingua"
	LangPosTd      = "td"
	LangPosHost    = "host"
	BodyChunkSize  = 2048
)

const (
	RegexLangHtml = "^(?i)([a-z]{2}|[a-z]{2}\\-[a-z]+)$"
)

var (
	regexLangHtmlPattern = regexp.MustCompile(RegexLangHtml)
)

type LangRes struct {
	Lang    string
	LangPos string
}

func Lang(doc *goquery.Document, charset string, list bool) LangRes {
	var res LangRes
	var lang string

	// 如果存在特定语言的 charset 对照表, 则直接返回
	if charset != "" {
		if _, exist := CharsetLangMap[charset]; exist {
			res.Lang = CharsetLangMap[charset]
			res.LangPos = LangPosCharset
			return res
		}
	}

	// 解析 Html 语言属性, 当不为空不为 en 时可信度比较高, 直接返回
	lang = LangFromHtml(doc)
	if lang != "" && lang != "en" {
		res.Lang = lang
		res.LangPos = LangPosHtmlTag
		return res
	}

	// 当 utf 编码时, lang 为空或 en 可信度比较低, 进行基于内容语种的检测
	if strings.HasPrefix(charset, "UTF") && (lang == "" || lang == "en") {
		// 列表模式, 优先判断TD是否是中文或日语

		tdLang, pos := LangFromTd(doc, list)
		if tdLang != "" {
			res.Lang = tdLang
			res.LangPos = pos
			return res
		}

		// 最后根据内容进行语种判断
		bodyLang, pos := LangFromUtf8Body(doc, list)
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
		lang = strings.TrimSpace(lang)
		if regexLangHtmlPattern.MatchString(lang) {
			lang = fun.SubString(lang, 0, 2)
			return lang
		}
	}
	if lang, exists := doc.Find("html").Attr("xml:lang"); exists {
		lang = strings.TrimSpace(lang)
		if regexLangHtmlPattern.MatchString(lang) {
			lang = fun.SubString(lang, 0, 2)
			return lang
		}

	}
	for _, selector := range metaLangSelectors {
		if lang, exists := doc.Find(selector).Attr("content"); exists {
			lang = strings.TrimSpace(lang)
			if regexLangHtmlPattern.MatchString(lang) {
				lang = fun.SubString(lang, 0, 2)
				return lang
			}
		}
	}

	return lang
}
func LangFromTd(doc *goquery.Document, list bool) (string, string) {
	var lang string
	var text string

	// 列表模式
	if list {
		// 获取 TD
		title := extract.WebTitle(doc, 0)
		description := extract.WebDescription(doc, 0)
		text = title + description

		text = fun.RemoveSign(text)

		text = strings.TrimSpace(text)

		// 截取后的字符长度
		textCount := utf8.RuneCountInString(text)

		if textCount >= 2 {
			// 首先判断是否包含汉字
			hanRegex := regexp.MustCompile(`\p{Han}`)
			han := hanRegex.FindAllString(text, -1)
			if han != nil {
				hanCount := len(han)
				hanRate := float64(hanCount) / float64(textCount)

				// 汉字比例
				if hanRate >= 0.38 {

					// 抽取内容验证是否有日语
					bodyText := bodyTextForLang(doc, list)

					jaRegex := regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
					ja := jaRegex.FindAllString(bodyText, -1)
					if ja != nil {
						jaCount := len(ja)
						jaRate := float64(jaCount) / float64(hanCount)

						// 日语占比
						if jaRate > 0.1 {
							lang = "ja"
							return lang, LangPosTd
						}
					}

					lang = "zh"
					return lang, LangPosTd
				}
			}
		}
	}

	return lang, ""
}

func LangFromUtf8Body(doc *goquery.Document, list bool) (string, string) {
	var lang string
	var text string

	// 抽取内容
	text = bodyTextForLang(doc, list)

	// 内容太少, 直接返回
	textCount := utf8.RuneCountInString(text)
	if textCount < 64 {
		return "", ""
	}

	// 去除换行(为了保留语义只替换多余的空格)
	text = fun.RemoveLines(text)
	text = strings.ReplaceAll(text, fun.TAB, "")
	text = strings.ReplaceAll(text, "  ", "")

	// 去除符号
	m := regexp.MustCompile(`[\pP\pS]`)
	text = m.ReplaceAllString(text, "")

	// 最大截取 BodyChunkSize 个字符
	text = fun.SubString(text, 0, BodyChunkSize)
	text = strings.TrimSpace(text)

	// 截取后的字符长度
	textCount = utf8.RuneCountInString(text)

	// 首先判断是否包含汉字, 中文和日语
	hanRegex := regexp.MustCompile(`\p{Han}`)
	han := hanRegex.FindAllString(text, -1)
	if han != nil {
		hanCount := len(han)
		hanRate := float64(hanCount) / float64(textCount)

		// 汉字比例
		if hanRate >= 0.38 {
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

	// 其次判断拉丁语系, 分析主要的一些语种
	englishRegexp := regexp.MustCompile(`[a-zA-Z]`)
	english := englishRegexp.FindAllString(text, -1)
	if english != nil {
		englishCount := len(english)
		englishRate := float64(englishCount) / float64(textCount)
		if englishRate > 0.38 {

			// 包含拉丁补充字符集, 使用 lingua 分析主要的非英语拉丁语种
			latinRegexp := regexp.MustCompile("[\u0080-\u00ff]")
			latin := latinRegexp.FindAllString(text, -1)
			if latin != nil {
				latinCount := len(latin)

				if latinCount > 3 {
					detector := lingua.NewLanguageDetectorBuilder().FromLanguages(linguaLatinLanguages...).Build()
					if language, exists := detector.DetectLanguageOf(text); exists {
						key := strings.ToLower(language.String())
						linguaLang := linguaMap[key]
						return linguaLang, LangPosLingua
					}
				}
			}

			return "en", LangPosBody
		}
	}

	// 最后, 使用 lingua 分析其他主要的非拉丁语种
	detector := lingua.NewLanguageDetectorBuilder().FromLanguages(linguaLanguages...).Build()
	if language, exists := detector.DetectLanguageOf(text); exists {

		key := strings.ToLower(language.String())
		linguaLang := linguaMap[key]
		return linguaLang, LangPosLingua
	}

	return lang, ""
}

func bodyTextForLang(doc *goquery.Document, list bool) string {
	var text string

	// 列表页模式
	if list {
		// 优先获取网页中最多 64 个 a 标签, 如果没有 a 标签或过少，放弃
		aTag := doc.Find("a")
		aTagSize := aTag.Size()
		if aTagSize >= 16 {
			sliceMax := fun.Min(aTagSize, 64)
			text = aTag.Slice(0, sliceMax).Text()

			// 如果 a 标签中包含过多的 {} 可能是动态渲染, 放弃
			if strings.Count(text, "{") >= 5 && strings.Count(text, "}") >= 5 {
				text = ""
			}
		}
	} else {
		// 内容页模式, 获取网页中最多 64 个 p 标签
		pTag := doc.Find("p")
		pTagSize := pTag.Size()
		sliceMax := fun.Min(pTagSize, 64)
		text = pTag.Slice(0, sliceMax).Text()

		// 如果内容太少, 获取全部 body 文本
		textCount := utf8.RuneCountInString(text)
		if textCount < 64 {
			text = doc.Find("body").Text()
		}
	}

	return text
}

func LangFromHost(host string) string {
	var lang string

	host = strings.ToLower(host)
	if strings.HasSuffix(host, ".fr") {
		lang = "fr"
	} else if strings.HasSuffix(host, ".de") {
		lang = "de"
	} else if strings.HasSuffix(host, ".es") {
		lang = "es"
	} else if strings.HasSuffix(host, ".ar") {
		lang = "es"
	} else if strings.HasSuffix(host, ".pt") {
		lang = "pt"
	} else if strings.HasSuffix(host, ".br") {
		lang = "pt"
	} else if strings.HasSuffix(host, ".ru") {
		lang = "ru"
	} else if strings.HasSuffix(host, ".pl") {
		lang = "pt"
	} else if strings.HasSuffix(host, ".in") {
		lang = "hi"
	}

	return lang
}
