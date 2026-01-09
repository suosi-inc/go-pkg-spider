package extract

import "regexp"

const (
	ContentRemoveTags = "script,noscript,style,iframe,br,link,svg,textarea"

	// RegexPublishDate 完整的发布时间正则
	RegexPublishDate = "(((20[1-3]\\d{1})[-/年.])(0[1-9]|1[0-2]|[1-9])[-/月.](0[1-9]|[1-2][0-9]|3[0-1]|[1-9])[日Tt]?[ ]{0,3}(([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)?((\\.\\d{3})?)(z|Z|[\\+-]\\d{2}[:]?\\d{2})?)?)"

	// RegexPublishShortDate 年份缩写发布时间正则, 如 22-09-02 11:11:11
	RegexPublishShortDate = "(((20[1-3]\\d{1}|[1-3]\\d{1})[-/年.])(0[1-9]|1[0-2]|[1-9])[-/月.](0[1-9]|[1-2][0-9]|3[0-1]|[1-9])[日Tt]?[ ]{0,3}(([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)?((\\.\\d{3})?)(z|Z|[\\+-]\\d{2}[:]?\\d{2})?)?)"

	// RegexPublishDateNoYear 不包含年的发布时间(优先级低), 09-02
	RegexPublishDateNoYear = "((0[1-9]|1[0-2]|[1-9])[-/月.](0[1-9]|[1-2][0-9]|3[0-1]|[1-9])[日Tt]?[ ]{0,3}(([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)?)?)"

	// RegexEnPublishDate1 英文格式的正则1, 如 02 Sep 2022 11:40:53 pm
	RegexEnPublishDate1 = "(?i)((?:(0[1-9]|[1-2][0-9]|3[0-1]|[1-9])(?:st|nd|rd|th)?)[, ]{0,4}(january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|may|jun|jul|aug|sept?|oct|nov|dec)[, ]{0,4}(20[1-3]\\d{1})([, ]{0,4}([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:]([0-5][0-9]|[0-9])([:]([0-5][0-9]|[0-9]))?([, ]{0,4}(am|pm))?)?)"

	// RegexEnPublishDate2 英文格式的正则2, 如 Sep 02 2022 11:40:53 pm
	RegexEnPublishDate2 = "(?i)((january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|may|jun|jul|aug|sept?|oct|nov|dec)[, ]{0,4}(?:(0[1-9]|[1-2][0-9]|3[0-1]|[1-9])(?:st|nd|rd|th)?)[, ]{0,4}(20[1-3]\\d{1})([, ]{0,4}([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:]([0-5][0-9]|[0-9])([:]([0-5][0-9]|[0-9]))?([, ]{0,4}(am|pm))?)?)"

	// RegexEnUsPublishDate 英文美式格式的正则3, 如 8/30/2022 11:11:11
	RegexEnUsPublishDate = "((0[1-9]|1[0-2]|[1-9])[-/.](0[1-9]|[1-2][0-9]|3[0-1]|[1-9])[-/.](20[1-3]\\d{1}|[1-3]\\d{1})[ ]{0,3}(([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:]([0-5][0-9]|[0-9])[:]?(([0-5][0-9]|[0-9]))?)?)"

	// RegexTime 仅时间正则
	RegexTime = "([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)?"

	// RegexZhPublishPrefix 中文的发布时间前缀
	RegexZhPublishPrefix = "(?i)(发布|创建|出版|发表|编辑)?(时间|日期)"

	// RegexZhPublishDate 中文的固定格式, 如 发布时间: xxx
	RegexZhPublishDate = RegexZhPublishPrefix + "[\\pP ]{1,8}" + RegexPublishShortDate

	// RegexScriptTitle Script 中的标题
	RegexScriptTitle = `(?i)"title"[\t ]{0,4}:[\t ]{0,4}"(.*)"`

	// RegexScriptTime Script 中的发布时间
	RegexScriptTime = `(?i)"[\w_\-]*pub.*"[\t ]{0,4}:[\t ]{0,4}"(((20[1-3]\d{1})[-/年.])(0[1-9]|1[0-2]|[1-9])[-/月.](0[1-9]|[1-2][0-9]|3[0-1]|[1-9])[日Tt]?[ ]{0,3}(([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)?((\.\d{3})?)(z|Z|[\+-]\d{2}[:]?\d{2})?))"`

	// RegexWxScriptTime 微信 Script 中的发布时间
	RegexWxScriptTime = `(?i)ct[\t ]{0,4}=[\t ]{0,4}"(1[2-9]\d{8})"`

	// RegexContentUrlPublishDate 内容页URL中隐藏的时间, 必须是非常完整标准的时间 20221003
	RegexContentUrlPublishDate = `(20[2-3]\d{1}[/]?(0[1-9]|1[0-2])[/]?(0[1-9]|[1-2][0-9]|3[0-1]))`

	// RegexFormatTime3 错误的时间格式, 用于过滤
	RegexFormatTime3 = `[:分]\d{3}$`

	// RegexFormatTime4 错误的时间格式, 用于过滤
	RegexFormatTime4 = `[:分]\d{4}$`

	// RegexZone 错误的时区格式, 用于过滤
	RegexZone = `(([\+-]\d{2})[:]?\d{2})$`

	// TitleSimZh 中文相似度阈值
	TitleSimZh = 0.3

	// TitleSimWord 单词相似度阈值
	TitleSimWord = 0.5
)

var (
	contentMetaTitleSelectors = []string{
		"meta[property='og:title' i]",
		"meta[property='twitter:title' i]",
		"meta[name='twitter:title' i]",
	}

	contentMetaDatetimeDicts = []string{"publish", "pubdate", "pubtime", "release", "dctermsdate"}

	regexPublishDatePattern = regexp.MustCompile(RegexPublishDate)

	regexPublishShortDatePattern = regexp.MustCompile(RegexPublishShortDate)

	regexPublishDateNoYearPattern = regexp.MustCompile(RegexPublishDateNoYear)

	regexZhPublishDatePattern = regexp.MustCompile(RegexZhPublishDate)

	regexEnPublishDatePattern1 = regexp.MustCompile(RegexEnPublishDate1)

	regexEnPublishDatePattern2 = regexp.MustCompile(RegexEnPublishDate2)

	regexEnUsPublishDatePattern = regexp.MustCompile(RegexEnUsPublishDate)

	regexTimePattern = regexp.MustCompile(RegexTime)

	regexScriptTitlePattern = regexp.MustCompile(RegexScriptTitle)

	regexScriptTimePattern = regexp.MustCompile(RegexScriptTime)

	regexWxScriptTimePattern = regexp.MustCompile(RegexWxScriptTime)

	regexContentUrlPublishDatePattern = regexp.MustCompile(RegexContentUrlPublishDate)

	regexFormatTime3 = regexp.MustCompile(RegexFormatTime3)

	regexFormatTime4 = regexp.MustCompile(RegexFormatTime4)

	regexZonePattern = regexp.MustCompile(RegexZone)
)
