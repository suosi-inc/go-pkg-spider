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
		"it": "意大利语",
		"th": "泰语",
		"vi": "越南语",
		"my": "缅甸语",
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
		"意大利语": "it",
		"泰语":   "th",
		"越南语":  "vi",
		"缅甸语":  "my",
	}

	langMetaSelectors = []string{
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
	LangPosTitleZh = "title"
	BodyChunkSize  = 2048
	BodyMinSize    = 64
)

const (
	RegexLangHtml = "^(?i)([a-z]{2}|[a-z]{2}\\-[a-z]+)$"
)

var (
	regexLangHtmlPattern = regexp.MustCompile(RegexLangHtml)
	regexPuncsPattern    = regexp.MustCompile(`[\pP\pS]`)
	regexEnPattern       = regexp.MustCompile(`[a-zA-Z]`)
	regexLatinPattern    = regexp.MustCompile("[\u0080-\u00ff]")
	regexZhPattern       = regexp.MustCompile(`\p{Han}`)
	regexJaPattern       = regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
	regexKoPattern       = regexp.MustCompile(`\p{Hangul}`)
)

type LangRes struct {
	Lang    string
	LangPos string
}

// LangText 探测纯文本语种
func LangText(text string) (string, string) {
	return langFromText(text)
}

// Lang 探测 HTML 语种
func Lang(doc *goquery.Document, charset string, listMode bool) LangRes {
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

	// 优先判断Title是否包含中文, 辅助内容排除日韩
	titleLang, pos := LangFromTitle(doc, listMode)
	if titleLang != "" {
		res.Lang = titleLang
		res.LangPos = pos
		return res
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
		bodyLang, pos := LangFromUtf8Body(doc, listMode)
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
	for _, selector := range langMetaSelectors {
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
func LangFromTitle(doc *goquery.Document, listMode bool) (string, string) {
	var lang string
	var text string

	// 获取 Title
	title := extract.WebTitle(doc, 0)
	text = fun.RemoveSign(title)
	text = strings.TrimSpace(text)

	if text != "" {
		// 首先判断标题是否包含汉字
		han := regexZhPattern.FindAllString(text, -1)
		if han != nil {
			hanCount := len(han)

			// 汉字数量 >=2
			if hanCount >= 2 {

				// 需要抽取内容验证包含有日语韩语, 如(日本語_新華網)
				bodyText := bodyTextForLang(doc, listMode)

				// 去除所有符号
				bodyText = fun.RemoveSign(bodyText)

				// 最大截取 BodyChunkSize 个字符
				bodyText = fun.SubString(bodyText, 0, BodyChunkSize)
				bodyText = strings.TrimSpace(bodyText)

				bodyTextCount := utf8.RuneCountInString(bodyText)

				// 包含一定的日语
				ja := regexJaPattern.FindAllString(bodyText, -1)
				if ja != nil {
					jaCount := len(ja)
					jaRate := float64(jaCount) / float64(bodyTextCount)

					// 日语出现比例
					if jaRate > 0.2 {
						lang = "ja"
						return lang, LangPosTitleZh
					}
				}

				// 包含一定的韩语
				ko := regexKoPattern.FindAllString(bodyText, -1)
				if ko != nil {
					koCount := len(ko)
					koRate := float64(koCount) / float64(bodyTextCount)

					// 韩语出现比例
					if koRate > 0.2 {
						lang = "ko"
						return lang, LangPosTitleZh
					}
				}

				lang = "zh"
				return lang, LangPosTitleZh
			}
		}
	}

	return lang, ""
}

func LangFromUtf8Body(doc *goquery.Document, listMode bool) (string, string) {
	var text string

	// 抽取内容
	text = bodyTextForLang(doc, listMode)

	return langFromText(text)
}

func langFromText(text string) (string, string) {
	var lang string

	// 去除换行(为了保留语义只替换多余的空格)
	text = fun.RemoveLines(text)
	text = strings.ReplaceAll(text, fun.TAB, "")
	text = strings.ReplaceAll(text, "  ", "")

	// 去除符号
	text = regexPuncsPattern.ReplaceAllString(text, "")

	// 最大截取 BodyChunkSize 个字符
	text = fun.SubString(text, 0, BodyChunkSize)
	text = strings.TrimSpace(text)

	// 截取后的字符长度
	textCount := utf8.RuneCountInString(text)

	// 内容太少不足以判断语言, 放弃
	if textCount < BodyMinSize {
		return "", ""
	}

	// 首先判断是否包含汉字, 中文和日语
	han := regexZhPattern.FindAllString(text, -1)
	if han != nil {
		hanCount := len(han)
		hanRate := float64(hanCount) / float64(textCount)

		// 汉字比例
		if hanRate >= 0.3 {
			ja := regexJaPattern.FindAllString(text, -1)
			if ja != nil {
				jaCount := len(ja)
				jaRate := float64(jaCount) / float64(hanCount)

				// 日语在汉字中的占比
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
	english := regexEnPattern.FindAllString(text, -1)
	if english != nil {
		englishCount := len(english)
		englishRate := float64(englishCount) / float64(textCount)
		if englishRate > 0.618 {

			// 包含拉丁补充字符集, 使用 lingua 分析主要的非英语拉丁语种
			latin := regexLatinPattern.FindAllString(text, -1)
			if latin != nil {
				latinCount := len(latin)

				if latinCount > 5 {
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

func bodyTextForLang(doc *goquery.Document, listMode bool) string {
	var text string

	// 列表页模式
	if listMode {
		// 优先获取网页中最多 64 个 a 标签, 如果没有 a 标签或过少，放弃
		aTag := doc.Find("a")
		aTagSize := aTag.Size()
		if aTagSize >= 16 {
			sliceMax := fun.Min(aTagSize, 64)
			text = aTag.Slice(0, sliceMax).Text()

			// 如果 a 标签中包含过多的 {} 可能是动态渲染, 放弃
			if strings.Count(text, "{") >= 5 && strings.Count(text, "}") >= 5 {
				return ""
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
		if textCount < BodyMinSize {
			text = doc.Find("body").Text()
		}
	}

	return text
}
