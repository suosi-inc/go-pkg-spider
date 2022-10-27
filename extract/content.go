// Package extract 新闻要素抽取, 在 CEPF 算法基础上做了大量的优化
// Refer to: 基于标签路径特征融合新闻内容抽取的 CEPF 算法 (吴共庆等) http://www.jos.org.cn/jos/article/abstract/4868
package extract

import (
	"bytes"
	"log"
	"math"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
	"golang.org/x/net/html"
)

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
	RegexWxScriptTime = `(?i)ct[\t ]{0,4}=[\t ]{0,4}"(1[4-9]\d{8})"`

	// RegexContentUrlPublishDate 内容页URL中隐藏的时间
	RegexContentUrlPublishDate = `(20[2-3]\d{1}[/]?(0[1-9]|1[0-2]|[1-9])[/]?(0[1-9]|[1-2][0-9]|3[0-1]|[1-9]))`

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

type News struct {
	// 标题
	Title string
	// 标题提取依据
	TitlePos string
	// 发布时间
	TimeLocal string
	// 时间
	Time string
	// 时间提取依据
	TimePos string
	// 正文纯文本
	Content string
	// 正文 Node 节点
	ContentNode *html.Node
	// 响应毫秒
	Spend int64
	// 语种
	Lang string
}

type Content struct {
	// 原始 Doc
	OriginDoc *goquery.Document
	// Doc
	Doc *goquery.Document
	// 原始标题, 来自于上级页面
	OriginTitle string
	// 原始链接, 来自于上级页面
	OriginUrl string
	// 语种
	Lang string

	infoMap      map[*html.Node]countInfo
	bodyNode     *html.Node
	title        string
	titlePos     string
	titleSim     float64
	timePos      string
	timeEnFormat bool
}

type countInfo struct {
	// 文本长度, 如 <p> 标签的文本
	TextCount int
	// 带有链接的文本长度, 如 <a> 标签中的文本
	LinkTextCount int
	// 标签数量
	TagCount int
	// 带有链接的标签数量
	LinkTagCount int
	// 密度
	Density float64
	// 密度统计
	DensitySum float64
	// <p> 标签数量
	PCount int
	// 叶子列表
	LeafList []int
}

func NewContent(doc *goquery.Document, lang string, originTitle string, originUrl string) *Content {
	originDoc := goquery.CloneDocument(doc)
	doc.Find(ContentRemoveTags).Remove()

	// 标题相似度阈值判定
	titleSim := TitleSimZh
	if fun.SliceContains(wordLangs, lang) {
		titleSim = TitleSimWord
	}

	infoMap := make(map[*html.Node]countInfo, 0)

	return &Content{OriginDoc: originDoc, Doc: doc, OriginTitle: originTitle, OriginUrl: originUrl, Lang: lang, infoMap: infoMap, titleSim: titleSim}
}

func (c *Content) ExtractNews() *News {
	news := &News{}

	// 开始时间
	begin := fun.Timestamp(true)

	// 提取正文结点和正文
	contentNode := c.getContentNode()
	if contentNode != nil {
		news.ContentNode = contentNode

		content := c.formatContent(contentNode)
		news.Content = content
	}

	// 提取标题
	title := c.getTitle(contentNode)
	news.Title = title
	news.TitlePos = c.titlePos
	c.title = title

	// 提取发布时间
	time := c.getTime()
	if time != "" {
		// 格式化时间
		news.Time = time
		news.TimePos = c.timePos
		time = c.formatTime(time)
		ts := fun.StrToTime(time)
		if ts > 0 {
			news.TimeLocal = fun.Date(ts)
		}
	}

	news.Spend = fun.Timestamp(true) - begin
	news.Lang = c.Lang

	return news
}

// formatTime 时间格式化清洗(尽可能的)
func (c *Content) formatTime(time string) string {
	if !c.timeEnFormat {
		// 当包含时区信息时格式化空格
		if fun.ContainsAny(time, "T", "t", "Z", "z") {
			time = strings.ReplaceAll(time, " ", "")
		}
		// 当包含时区T时又没有偏移, 按本地时间处理
		if fun.Contains(time, "T") && !fun.ContainsCase(time, "z") {
			if !regexZonePattern.MatchString(time) {
				time = strings.ReplaceAll(time, "T", " ")
			}
		}
	}

	// 错误的尾巴处理
	if fun.Contains(time, ":") && !fun.ContainsAny(time, "时", "点") {
		time = strings.TrimSuffix(time, "分")
	}
	return time
}

// formatContent 正文格式化, 处理 <p> 的换行, 最终将多个换行符和空格均合并为一个
func (c *Content) formatContent(contentNode *html.Node) string {
	// 先提取 HTML
	node := goquery.NewDocumentFromNode(contentNode)
	contentHtml, _ := node.Html()

	// 给 <p>  则增加换行 \n
	contentHtml = strings.ReplaceAll(contentHtml, "</p>", "</p>\n")
	n, _ := goquery.NewDocumentFromReader(strings.NewReader(contentHtml))
	str := n.Text()

	// 最后合并多余的换行
	lines := fun.SplitTrim(str, fun.LF)
	if len(lines) > 0 {
		for i, line := range lines {
			lines[i] = fun.NormaliseSpace(line)
		}
		str = strings.Join(lines, fun.LF)
	} else {
		str = fun.NormaliseSpace(str)
	}

	return str
}

func (c *Content) getContentNode() *html.Node {
	var maxScore float64
	var contentNode *html.Node

	// 取第一个 body 标签
	bodyNodes := c.Doc.Find("body").Nodes
	if len(bodyNodes) > 0 {
		bodyNode := bodyNodes[0]
		c.bodyNode = bodyNode

		// 递归遍历计算并统计, 最后找得分最高那个节点
		c.computeInfo(c.bodyNode)

		for node := range c.infoMap {
			if node.Data == "a" || node == bodyNode {
				continue
			}

			score := c.computeScore(node)
			if score > maxScore {
				maxScore = score
				contentNode = node
			}
		}
	}

	return contentNode
}

func (c *Content) getTime() string {
	// meta
	regexZhPatterns := []*regexp.Regexp{
		regexPublishDatePattern,
	}
	metaZhTime := c.getTimeByMeta(regexZhPatterns)
	if metaZhTime != "" {
		c.timePos = "meta"
		return metaZhTime
	}

	// meta En
	if c.Lang != "zh" {
		regexEnPatterns := []*regexp.Regexp{
			regexEnPublishDatePattern1,
			regexEnPublishDatePattern2,
		}
		metaEnTime := c.getTimeByMetaEn(regexEnPatterns)
		if metaEnTime != "" {
			c.timePos = "meta"
			c.timeEnFormat = true
			return metaEnTime
		}
	}

	// <time> 标签
	tagTime := c.getTimeByTag()
	if tagTime != "" {
		c.timePos = "tag"
		return tagTime
	}

	// <script> 标签, 必须包含时间
	scriptTime := c.getTimeByScript()
	if scriptTime != "" {
		c.timePos = "script"
		return scriptTime
	}

	bodyText := c.Doc.Find("body").Text()
	bodyText = fun.NormaliseSpace(bodyText)

	// <body>
	contentTime := c.getTimeByBody(bodyText)
	if contentTime != "" {
		c.timePos = "body"
		return contentTime
	}

	// Lang专属
	langTime := c.getTimeByLang(bodyText)
	if langTime != "" {
		c.timePos = "lang"
		return langTime
	}

	urlTime := c.getTimeByUrl()
	if urlTime != "" {
		c.timePos = "url"
		return urlTime
	}

	return ""
}

func (c *Content) getTimeByLang(bodyText string) string {
	if c.Lang == "zh" {
		allRegexs := regexZhPublishDatePattern.FindAllString(bodyText, -1)

		if allRegexs != nil {
			publishDates := make([]string, 0)
			for _, regex := range allRegexs {
				regexDate := regexPublishShortDatePattern.FindString(regex)
				if regexDate != "" {
					publishDates = append(publishDates, regexDate)
				}
			}

			if len(publishDates) > 0 {
				return c.pickPublishDates(bodyText, publishDates, false)
			}
		}
	} else {
		// 第一种格式
		allRegexs := regexEnPublishDatePattern1.FindAllString(bodyText, -1)
		if allRegexs != nil {
			publishDates := make([]string, 0)
			for _, regex := range allRegexs {
				dateStr := strings.TrimSpace(regex)
				dateStr = fun.NormaliseSpace(dateStr)
				dateStr = strings.ReplaceAll(dateStr, ",", " ")
				publishDates = append(publishDates, dateStr)

			}

			if len(publishDates) > 0 {
				c.timeEnFormat = true
				return c.pickPublishDates(bodyText, publishDates, false)
			}
		}

		// 第二种格式
		allRegexs = regexEnPublishDatePattern2.FindAllString(bodyText, -1)
		if allRegexs != nil {
			publishDates := make([]string, 0)
			for _, regex := range allRegexs {
				dateStr := strings.TrimSpace(regex)
				dateStr = fun.NormaliseSpace(dateStr)
				dateStr = strings.ReplaceAll(dateStr, ",", " ")
				publishDates = append(publishDates, dateStr)

			}

			if len(publishDates) > 0 {
				c.timeEnFormat = true
				return c.pickPublishDates(bodyText, publishDates, false)
			}
		}

		// 第三种格式, 美式时间, 如 8/30/2022 11:11:11
		allRegexs = regexEnUsPublishDatePattern.FindAllString(bodyText, -1)
		if allRegexs != nil {
			publishDates := make([]string, 0)
			for _, regex := range allRegexs {
				dateStr := strings.TrimSpace(regex)
				publishDates = append(publishDates, dateStr)
			}

			if len(publishDates) > 0 {
				return c.pickPublishDates(bodyText, publishDates, false)
			}
		}
	}

	return ""
}

func (c *Content) getTimeByBody(bodyText string) string {
	// 带有年份的完整匹配
	publishDates := regexPublishShortDatePattern.FindAllString(bodyText, -1)
	if (publishDates) != nil {
		return c.pickPublishDates(bodyText, publishDates, false)
	}

	// 不带年份的匹配, 仅处理中文并且必须有时间, 如  01-01 01:00
	if c.Lang == "zh" {
		publishNoYearDates := regexPublishDateNoYearPattern.FindAllString(bodyText, -1)
		if publishNoYearDates != nil {
			noYear := c.pickPublishDates(bodyText, publishNoYearDates, true)
			if noYear != "" {
				if strings.Contains(noYear, "月") {
					year := fun.Date("2006年")
					return year + noYear
				} else {
					noYear = strings.NewReplacer("/", "-", ".", "-").Replace(noYear)
					year := fun.Date("2006-")
					return year + noYear
				}
			}

			return noYear
		}
	}

	return ""
}

func (c *Content) pickPublishDates(bodyText string, publishDates []string, requireTime bool) string {
	// 根据是否有时间进行分组
	hasTimes := make([]string, 0)
	noTimes := make([]string, 0)
	for _, date := range publishDates {
		dateStr := strings.TrimSpace(date)
		if regexTimePattern.MatchString(dateStr) {
			// 去除非法的尾巴
			if regexFormatTime3.MatchString(dateStr) {
				timeRunes := []rune(dateStr)
				timeRunes = timeRunes[:len(timeRunes)-1]
				dateStr = string(timeRunes)
			}
			if regexFormatTime4.MatchString(dateStr) {
				timeRunes := []rune(dateStr)
				timeRunes = timeRunes[:len(timeRunes)-2]
				dateStr = string(timeRunes)
			}
			hasTimes = append(hasTimes, dateStr)
		} else {
			noTimes = append(noTimes, dateStr)
		}
	}

	// 有时间的情况优先
	if len(hasTimes) > 0 {
		if len(hasTimes) == 1 {
			return hasTimes[0]
		}

		// 判断第一个是不是最长的, 如果最长就优先返回
		var maxLen int
		var maxIndex int
		for i, date := range hasTimes {
			length := utf8.RuneCountInString(date)
			if length > maxLen {
				maxLen = length
				maxIndex = i
			}
		}

		if maxIndex == 0 {
			return hasTimes[0]
		}

		// 找最靠近标题的那一个
		if c.title != "" && (c.titlePos == "selector" || c.titlePos == "headline" || c.titlePos == "content") {
			titleIndex := strings.Index(bodyText, c.title)

			minDistance := float64(math.MaxInt)
			var minIndex int
			for i, date := range hasTimes {
				dateIndex := strings.Index(bodyText, date)
				abs := math.Abs(float64(dateIndex) - float64(titleIndex))
				if abs < minDistance {
					minDistance = abs
					minIndex = i
				}
			}

			return hasTimes[minIndex]
		}

		// 没找到或标题不是正文区域, 最后返回第一个
		return hasTimes[0]
	}

	// 没有时间的情况
	if !requireTime {
		if len(noTimes) > 0 {
			if len(noTimes) == 1 {
				return noTimes[0]
			}

			// 英文时间格式
			if c.timeEnFormat {
				// 找最靠近标题的那一个
				if c.title != "" && (c.titlePos == "selector" || c.titlePos == "headline") {
					titleIndex := strings.Index(bodyText, c.title)

					minDistance := float64(math.MaxInt)
					var minIndex int
					for i, date := range noTimes {
						dateIndex := strings.Index(bodyText, date)
						abs := math.Abs(float64(dateIndex) - float64(titleIndex))
						if abs < minDistance {
							minDistance = abs
							minIndex = i
						}
					}

					return noTimes[minIndex]
				}

				// 返回第一个
				return noTimes[0]
			} else {
				// 返回最近的一个日期
				var maxTimestamp int64
				var maxIndex int
				for i, date := range noTimes {
					timestamp := fun.StrToTime(date)
					if timestamp > maxTimestamp {
						maxIndex = i
					}
				}

				return noTimes[maxIndex]
			}
			// 不会直接返回没有时间的, 因为可靠性低
		}
	}

	return ""
}

func (c *Content) getTimeByTag() string {
	timeTags := c.Doc.Find("time")
	if timeTags.Size() > 0 {
		firstTimeTags := timeTags.First()
		dateTime := firstTimeTags.AttrOr("datetime", "")
		if dateTime != "" {
			// 先匹配标准格式
			find := regexPublishDatePattern.FindString(dateTime)
			if find != "" {
				return find
			}

			// 非英文再匹配其他格式
			if c.Lang != "zh" {
				find = regexEnPublishDatePattern1.FindString(dateTime)
				if find != "" {
					find = fun.NormaliseSpace(find)
					find = strings.ReplaceAll(find, ",", " ")
					c.timeEnFormat = true
					return find
				}

				find = regexEnPublishDatePattern2.FindString(dateTime)
				if find != "" {
					find = fun.NormaliseSpace(find)
					find = strings.ReplaceAll(find, ",", " ")
					c.timeEnFormat = true
					return find
				}
			}
		}
	}

	return ""
}

func (c *Content) getTimeByMeta(regexPatterns []*regexp.Regexp) string {
	metaDates := make([]string, 0)
	metas := c.Doc.Find("meta")
	if metas.Size() > 0 {
		metas.Each(func(i int, meta *goquery.Selection) {
			content := meta.AttrOr("content", "")
			for _, regexPattern := range regexPatterns {
				dateStr := regexPattern.FindString(content)
				if dateStr != "" {
					name := meta.AttrOr("name", "")
					property := meta.AttrOr("property", "")
					replacer := strings.NewReplacer("_", "", "-", "", ".", "")
					name = replacer.Replace(name)
					property = replacer.Replace(property)
					if fun.ContainsAny(property, contentMetaDatetimeDicts...) {
						dateStr = strings.TrimSpace(dateStr)
						metaDates = append(metaDates, dateStr)
					}

					if fun.ContainsAny(name, contentMetaDatetimeDicts...) {
						dateStr = strings.TrimSpace(dateStr)
						metaDates = append(metaDates, dateStr)
					}

					break
				}
			}
		})
	}

	metaDatesLen := len(metaDates)
	if metaDatesLen > 0 {
		// 根据是否有时间进行分组
		hasTimes := make([]string, 0)
		noTimes := make([]string, 0)
		for _, date := range metaDates {
			if regexTimePattern.MatchString(date) {
				// 去除非法的尾巴
				hasTimes = append(hasTimes, date)
			} else {
				noTimes = append(noTimes, date)
			}
		}

		// 有时间的情况, 返回最长的
		if len(hasTimes) > 0 {
			if len(hasTimes) == 1 {
				return hasTimes[0]
			}

			var maxLen int
			var maxLenDate string
			for _, date := range hasTimes {
				length := utf8.RuneCountInString(date)
				if length > maxLen {
					maxLen = length
					maxLenDate = date
				}
			}

			return maxLenDate
		}

		// 返回最长的, 非中文情况下才会返回没有时间的
		if c.Lang != "zh" {
			if len(noTimes) > 0 {
				if len(noTimes) == 1 {
					return noTimes[0]
				}

				var maxLen int
				var maxLenDate string
				for _, date := range noTimes {
					length := utf8.RuneCountInString(date)
					if length > maxLen {
						maxLen = length
						maxLenDate = date
					}
				}

				return maxLenDate
			}
		}
	}

	return ""
}

func (c *Content) getTimeByMetaEn(regexPatterns []*regexp.Regexp) string {
	metaDates := make([]string, 0)
	metas := c.Doc.Find("meta")
	if metas.Size() > 0 {
		metas.Each(func(i int, meta *goquery.Selection) {
			content := meta.AttrOr("content", "")
			for _, regexPattern := range regexPatterns {
				dateStr := regexPattern.FindString(content)
				if dateStr != "" {
					name := meta.AttrOr("name", "")
					property := meta.AttrOr("property", "")
					replacer := strings.NewReplacer("_", "", "-", "", ".", "")
					name = replacer.Replace(name)
					property = replacer.Replace(property)

					if fun.ContainsAny(property, contentMetaDatetimeDicts...) {
						dateStr = strings.TrimSpace(dateStr)
						dateStr = fun.NormaliseSpace(dateStr)
						dateStr = strings.ReplaceAll(dateStr, ",", " ")
						metaDates = append(metaDates, dateStr)
					}

					if fun.ContainsAny(name, contentMetaDatetimeDicts...) {
						dateStr = strings.TrimSpace(dateStr)
						dateStr = fun.NormaliseSpace(dateStr)
						dateStr = strings.ReplaceAll(dateStr, ",", " ")
						metaDates = append(metaDates, dateStr)
					}

					break
				}
			}
		})
	}

	metaDatesLen := len(metaDates)
	if metaDatesLen > 0 {
		// 根据是否有时间进行分组
		hasTimes := make([]string, 0)
		noTimes := make([]string, 0)
		for _, date := range metaDates {
			if regexTimePattern.MatchString(date) {
				// 去除非法的尾巴
				hasTimes = append(hasTimes, date)
			} else {
				noTimes = append(noTimes, date)
			}
		}

		// 有时间的情况, 返回最长的
		if len(hasTimes) > 0 {
			if len(hasTimes) == 1 {
				return hasTimes[0]
			}

			var maxLen int
			var maxLenDate string
			for _, date := range hasTimes {
				length := utf8.RuneCountInString(date)
				if length > maxLen {
					maxLen = length
					maxLenDate = date
				}
			}

			return maxLenDate
		}

		// 返回最长的, 非中文情况下才会返回没有时间的
		if c.Lang != "zh" {
			if len(noTimes) > 0 {
				if len(noTimes) == 1 {
					return noTimes[0]
				}

				var maxLen int
				var maxLenDate string
				for _, date := range noTimes {
					length := utf8.RuneCountInString(date)
					if length > maxLen {
						maxLen = length
						maxLenDate = date
					}
				}

				return maxLenDate
			}
		}
	}

	return ""
}

// getTitleByOrigin 获取页面的 H[1-2] 标题, 找出与 OriginTitle 最像的
func (c *Content) getTitleByOrigin() string {
	if !fun.Blank(c.OriginTitle) {
		headlines := c.Doc.Find("h1,h2")
		if headlines.Size() > 0 {
			titles := make([]string, 0)
			titleSim := make([]float64, 0)
			headlines.Each(func(i int, headline *goquery.Selection) {
				text := fun.NormaliseSpace(headline.Text())
				sim := fun.SimilarityText(c.OriginTitle, text)
				if sim > c.titleSim {
					titleSim = append(titleSim, sim)
					titles = append(titles, text)
				}
			})

			if len(titles) > 0 {
				var title string
				var maxScore float64
				for i, t := range titles {
					if titleSim[i] > maxScore {
						title = t
					}
				}

				return title
			}
		}
	}

	return ""
}

func (c *Content) getTitle(contentNode *html.Node) string {
	var title string

	// 优先使用 originTitle 判定页面中的 H1-2
	title = c.getTitleByOrigin()
	if title != "" {
		c.titlePos = "headline"
		return title
	}

	// 页面 Title
	originMetaTitle := WebTitle(c.Doc, 255)

	// 去除原始 metaTitle 最后一个尾巴（一般是站点名称），再进行相似判断
	metaTitle := WebContentTitleClean(originMetaTitle, c.Lang)

	// 从 Meta 中提取相似 <title> 的标题，优先级较高，返回短的那个
	titleByMeta := c.getTitleByMeta(metaTitle)
	if titleByMeta != "" {
		c.titlePos = "meta"
		return titleByMeta
	}

	// 提取页面 Script 寻找是否包含有 title 的字段
	titleScript := c.getTitleByScript(metaTitle)
	if titleScript != "" {
		c.titlePos = "script"
		return titleScript
	}

	titleList := make([]*html.Node, 0)
	titleSim := make([]float64, 0)
	if !fun.Blank(originMetaTitle) && contentNode != nil {
		// 从 body 开始遍历，收集 h1->h2，并计算与 metaTitle 的相似度
		var traverse func(*html.Node)
		traverse = func(n *html.Node) {
			if n.FirstChild != nil {
				if n.Type == html.ElementNode {
					// 计算 h1->h2 的相似度
					tagName := n.Data
					if fun.Matches(tagName, "h[1-2]") {
						tagNode := goquery.NewDocumentFromNode(n)
						headTitle := c.normaliseText(tagNode.Selection)
						sim := fun.SimilarityText(headTitle, metaTitle)
						titleSim = append(titleSim, sim)
						titleList = append(titleList, n)
					}
				}

				for child := n.FirstChild; child != nil; child = child.NextSibling {
					traverse(child)
				}
			}
		}
		if c.bodyNode != nil {
			traverse(c.bodyNode)
		}

		// 从 h 标签中获取
		index := len(titleList)
		if index > 0 {
			var maxScore float64
			var maxIndex int
			maxIndex = -1

			// 找相似度最高的
			for i := 0; i < index; i++ {
				score := titleSim[i]
				if score > maxScore {
					maxScore = score
					maxIndex = i
				}
			}

			if maxIndex != -1 && maxScore > c.titleSim {
				c.titlePos = "headline"
				tagNode := goquery.NewDocumentFromNode(titleList[maxIndex])
				headTitle := c.normaliseText(tagNode.Selection)
				return headTitle
			}
		}
	}

	// 尝试从包含开头结尾 title 选择器中获取一个相似度高的
	titles := c.Doc.Find("body").Find("*[id^=title],*[id$=title],*[class^=title],*[class$=title]")
	if titles.Size() > 0 {
		first := titles.First()
		selectorTitle := c.normaliseText(first)
		sim := fun.SimilarityText(metaTitle, selectorTitle)
		if sim > c.titleSim {
			c.titlePos = "selector"
			return selectorTitle
		}
	}

	// 从正文中找最相似 metaTitle 的文本片段
	title = c.getTitleByEditDistance(metaTitle)
	if title != "" {
		c.titlePos = "content"
		return title
	}

	// 最坏的情况是, 直接返回页面标题
	c.titlePos = "title"
	return metaTitle
}

// getTitleByEditDistance 从正文中找最相似 metaTitle 的片段
func (c *Content) getTitleByEditDistance(originMetaTitle string) string {
	max := []float64{0.0}
	var buf bytes.Buffer

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {

		if n.FirstChild != nil {
			if n.Type == html.TextNode {
				node := goquery.NewDocumentFromNode(n)
				text := c.normaliseText(node.Selection)
				sim := fun.SimilarityText(text, originMetaTitle)
				if sim > c.titleSim && sim > max[0] {
					max[0] = sim
					buf.Reset()
					buf.WriteString(text)
				}
			}

			for child := n.FirstChild; child != nil; child = child.NextSibling {
				traverse(child)
			}
		}
	}
	if c.bodyNode != nil {
		traverse(c.bodyNode)
	}

	if len(buf.String()) > 0 {
		return buf.String()
	}

	return ""
}

func (c *Content) getTitleByMeta(metaTitle string) string {
	var titles []string
	for _, metaSelector := range contentMetaTitleSelectors {
		title := strings.TrimSpace(c.Doc.Find(metaSelector).AttrOr("content", ""))
		if !fun.Blank(title) {
			titles = append(titles, title)
		}
	}

	if len(titles) > 0 {
		if metaTitle != "" {
			for _, title := range titles {
				sim := fun.SimilarityText(title, metaTitle)
				if sim > c.titleSim {
					titleLen := utf8.RuneCountInString(title)
					metaTitleLen := utf8.RuneCountInString(metaTitle)
					if titleLen < metaTitleLen {
						c.titlePos = "title"
						return title
					} else {
						c.titlePos = "metaTitle"
						return metaTitle
					}
				}
			}
		} else {
			return titles[0]
		}
	}

	return ""
}

func (c *Content) computeInfo(node *html.Node) countInfo {
	if node.Type == html.ElementNode {
		countInfo := countInfo{}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			childCountInfo := c.computeInfo(child)
			countInfo.TextCount += childCountInfo.TextCount
			countInfo.LinkTextCount += childCountInfo.LinkTextCount
			countInfo.TagCount += childCountInfo.TagCount
			countInfo.LinkTagCount += childCountInfo.LinkTagCount
			countInfo.DensitySum += childCountInfo.Density
			countInfo.PCount += childCountInfo.PCount
			countInfo.LeafList = append(countInfo.LeafList, childCountInfo.LeafList...)
		}

		countInfo.TagCount++
		if node.Data == "a" {
			countInfo.LinkTextCount = countInfo.TextCount
			countInfo.LinkTagCount++
		} else if node.Data == "p" {
			countInfo.PCount++
		}

		pureLen := countInfo.TextCount - countInfo.LinkTextCount
		tagLen := countInfo.TagCount - countInfo.LinkTagCount
		if pureLen == 0 || tagLen == 0 {
			countInfo.Density = 0
		} else {
			countInfo.Density = float64(pureLen) / float64(tagLen)
		}

		c.infoMap[node] = countInfo

		return countInfo
	} else if node.Type == html.TextNode {
		countInfo := countInfo{}

		text := fun.NormaliseSpace(node.Data)
		textLen := utf8.RuneCountInString(text)
		countInfo.TextCount = textLen
		countInfo.LeafList = append(countInfo.LeafList, textLen)

		return countInfo
	} else {
		return countInfo{}
	}
}

func (c *Content) computeScore(node *html.Node) float64 {
	countInfo := c.infoMap[node]
	value := c.computeVar(countInfo.LeafList) + 1
	value = math.Sqrt(value)

	scoreLog10 := math.Log10(float64(countInfo.PCount) + 1)
	scoreLog := math.Log(float64(countInfo.TextCount) - float64(countInfo.LinkTextCount) + 1)
	score := math.Log(value) * countInfo.DensitySum * scoreLog * scoreLog10

	return score
}

func (c *Content) computeVar(leafList []int) float64 {
	leafLen := len(leafList)

	if leafLen == 0 {
		return 0
	}

	if leafLen == 1 {
		return float64(leafList[0]) / float64(2)
	}

	var sum float64
	for _, i := range leafList {
		sum += float64(i)
	}

	ave := sum / float64(leafLen)
	sum = 0
	for _, i := range leafList {
		t := (float64(i) - ave) * (float64(i) - ave)
		sum += t
	}

	sum = sum / float64(leafLen)
	return sum
}

func (c *Content) normaliseText(s *goquery.Selection) string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := fun.NormaliseSpace(n.Data)

			buf.WriteString(text)
		}
		if n.FirstChild != nil {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				f(child)
			}
		}
	}
	for _, n := range s.Nodes {
		f(n)
	}

	return buf.String()
}

func (c *Content) Debug() {
	for node, info := range c.infoMap {
		if node.Data == "div" {
			for _, a := range node.Attr {
				if a.Key == "id" || a.Key == "class" {
					log.Println(node.Attr)
					log.Println(info)
				}
			}
		}
	}
}

func (c *Content) getTitleByScript(metaTitle string) string {
	scripts := c.OriginDoc.Find("script")
	if scripts.Size() > 0 {
		var title string
		scripts.Each(func(i int, script *goquery.Selection) {
			scriptText := fun.NormaliseLine(script.Text())
			titleStrs := regexScriptTitlePattern.FindStringSubmatch(scriptText)
			if titleStrs != nil {
				titleStr := strings.TrimSpace(titleStrs[1])
				sim := fun.SimilarityText(metaTitle, titleStr)
				if sim > c.titleSim {
					title = titleStr
					return
				}
			}
		})

		if title != "" {
			return title
		}
	}

	return ""
}

func (c *Content) getTimeByScript() string {
	scripts := c.OriginDoc.Find("script")
	if scripts.Size() > 0 {
		var time string
		scripts.Each(func(i int, script *goquery.Selection) {
			scriptText := fun.NormaliseLine(script.Text())
			dateStrs := regexScriptTimePattern.FindStringSubmatch(scriptText)
			if dateStrs != nil {
				dateStr := strings.TrimSpace(dateStrs[1])
				time = dateStr
				return
			}

			dateStrs = regexWxScriptTimePattern.FindStringSubmatch(scriptText)
			if dateStrs != nil {
				dateStr := strings.TrimSpace(dateStrs[1])
				dateTs := fun.ToInt(dateStr)
				time = fun.Date(dateTs)
				return
			}
		})

		if time != "" {
			return time
		}
	}

	return ""
}

func (c *Content) getTimeByUrl() string {
	if c.OriginUrl != "" {
		if linkUrl, err := fun.UrlParse(c.OriginUrl); err == nil {
			// 内容页 URL path 时间特征统计
			pathDir := path.Dir(strings.TrimSpace(linkUrl.Path))
			pathClean := pathDirClean(pathDir)
			dateStr := regexContentUrlPublishDatePattern.FindString(pathClean)
			if dateStr != "" {
				dateStr = strings.ReplaceAll(dateStr, fun.SLASH, "")
				return dateStr
			}
		}
	}

	return ""
}
