// Package extract 新闻要素抽取, 基于标签路径特征融合新闻内容抽取的 CEPF 算法, (吴共庆等)
// Refer to: http://www.jos.org.cn/jos/article/abstract/4868
package extract

import (
	"bytes"
	"log"
	"math"
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
	RegexPublishDate = "(((20[1-2]\\d{1}|2\\d{1})[-/年.])(0[1-9]|1[0-2]|[1-9])[-/月.](0[1-9]|[1-2][0-9]|3[0-1]|[1-9])[日T]?\x20{0,2}(([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)?((\\.\\d{3})?)(z|Z|[\\+-]\\d{2}[:]?\\d{2})?)?)"
	// RegexTime 仅时间正则
	RegexTime = "([0-9]|[0-1][0-9]|2[0-3]|[1-9])[:点时]([0-5][0-9]|[0-9])[:分]?(([0-5][0-9]|[0-9])[秒]?)"
	// RegexZhPublishPrefix 中文的发布时间前缀
	RegexZhPublishPrefix = "(?i)(发布|创建|出版|发表|编辑)?(时间|日期|于)"
)

var (
	metaTitleSelectors = []string{
		"meta[property='og:title' i]",
		"meta[property='twitter:title' i]",
		"meta[name='twitter:title' i]",
	}

	metaDatetimeDicts = []string{"publish", "pubdate", "pubtime"}

	regexPublishDatePattern = regexp.MustCompile(RegexPublishDate)

	regexTimePattern = regexp.MustCompile(RegexTime)
)

type News struct {
	// 标题
	Title string
	// 标题依据
	TitlePos string
	// 发布时间
	TimeLocal string
	// 时间
	Time string
	// 时间依据
	TimePos string
	// 正文纯文本
	Content string
	// 正文节点
	ContentNode *html.Node
}

type Content struct {
	Doc         *goquery.Document
	OriginTitle string
	Lang        string
	infoMap     map[*html.Node]CountInfo

	bodyNode *html.Node
	title    string
	titlePos string
	timePos  string
}

type CountInfo struct {
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

func NewContent(doc *goquery.Document, lang string, originTitle string) *Content {
	infoMap := make(map[*html.Node]CountInfo, 0)
	return &Content{Doc: doc, OriginTitle: originTitle, Lang: lang, infoMap: infoMap}
}

func (c *Content) News() *News {
	news := &News{}

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
	news.Time = time
	news.TimePos = c.timePos
	if news.Time != "" {
		news.TimeLocal = fun.Date(fun.StrToTime(time))
	}

	return news
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

	// 最后合并多余的换行符
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
	c.Doc.Find(ContentRemoveTags).Remove()
	bodyNodes := c.Doc.Find("body").Nodes
	if len(bodyNodes) > 0 {
		bodyNode := bodyNodes[0]
		c.bodyNode = bodyNode

		// 递归遍历计算并统计, 最后找得分最高节点
		c.computeInfo(c.bodyNode)

		// c.debug()

		for node, _ := range c.infoMap {
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
	metaTime := c.getTimeByMeta()

	if metaTime != "" {
		c.timePos = "meta"
		return metaTime
	}

	bodyText := c.Doc.Find("body").Text()
	bodyText = fun.NormaliseSpace(bodyText)

	langTime := c.getTimeByLang(bodyText)
	if langTime != "" {
		c.timePos = "lang"
		return langTime
	}

	contentTime := c.getTimeByBody(bodyText)
	if contentTime != "" {
		c.timePos = "body"
		return contentTime
	}

	return ""
}

func (c *Content) getTimeByLang(bodyText string) string {
	if c.Lang == "zh" {
		regexStr := RegexZhPublishPrefix + ".{0,32}" + RegexPublishDate
		allRegexs := regexp.MustCompile(regexStr).FindAllString(bodyText, -1)

		if allRegexs != nil {
			publishDates := make([]string, 0)
			for _, regex := range allRegexs {
				regexDate := regexPublishDatePattern.FindString(regex)
				if regexDate != "" {
					publishDates = append(publishDates, regexDate)
				}
			}

			if len(publishDates) > 0 {
				return c.pickPublishDates(bodyText, publishDates)
			}
		}
	}

	return ""
}

func (c *Content) getTimeByBody(bodyText string) string {
	// 带有年份的完整匹配
	publishDates := regexPublishDatePattern.FindAllString(bodyText, -1)
	if (publishDates) != nil {
		return c.pickPublishDates(bodyText, publishDates)
	}

	return ""
}

func (c *Content) pickPublishDates(bodyText string, publishDates []string) string {
	// 根据是否有时间进行分组
	hasTimes := make([]string, 0)
	noTimes := make([]string, 0)
	for _, date := range publishDates {
		dateStr := strings.TrimSpace(date)
		if regexTimePattern.MatchString(dateStr) {
			hasTimes = append(hasTimes, dateStr)
		} else {
			noTimes = append(noTimes, dateStr)
		}
	}

	// 有时间的情况
	if len(hasTimes) > 0 {
		if len(hasTimes) == 1 {
			return hasTimes[0]
		}

		// 判断第一个是不是最长的, 如果最长就直接返回
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
		if c.title != "" && (c.titlePos == "selector" || c.titlePos == "headline") {
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
	}

	// 没有时间的情况
	if len(noTimes) > 0 {
		if len(noTimes) == 1 {
			return noTimes[0]
		}

		// 判断第一个是不是最长的, 如果最长就直接返回
		var maxLen int
		var maxIndex int
		for i, date := range noTimes {
			length := utf8.RuneCountInString(date)
			if length > maxLen {
				maxLen = length
				maxIndex = i
			}
		}

		if maxIndex == 0 {
			return noTimes[0]
		}

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
	}

	return ""
}

func (c *Content) getTimeByMeta() string {
	metaDates := make([]string, 0)
	metas := c.Doc.Find("meta")
	if metas.Size() > 0 {
		metas.Each(func(i int, meta *goquery.Selection) {
			content := meta.AttrOr("content", "")
			if regexPublishDatePattern.MatchString(content) {
				name := meta.AttrOr("name", "")
				property := meta.AttrOr("property", "")
				name = strings.ReplaceAll(name, "_", "")
				name = strings.ReplaceAll(name, "-", "")
				property = strings.ReplaceAll(property, "_", "")
				property = strings.ReplaceAll(property, "-", "")

				if fun.ContainsAny(property, metaDatetimeDicts...) {
					dateStr := strings.TrimSpace(content)
					metaDates = append(metaDates, dateStr)
				}

				if fun.ContainsAny(name, metaDatetimeDicts...) {
					dateStr := strings.TrimSpace(content)
					metaDates = append(metaDates, dateStr)
				}
			}
		})
	}

	// 多个返回最长的并且包含时间的
	metaDatesLen := len(metaDates)
	if metaDatesLen > 0 {
		// 根据是否有时间进行分组
		hasTimes := make([]string, 0)
		for _, date := range metaDates {
			if regexTimePattern.MatchString(date) {
				hasTimes = append(hasTimes, date)
			}
		}

		// 有时间的情况
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
	}

	return ""
}

// getTitleByOrigin 获取页面的 H[1-2] 标题, 找出 OriginTitle 最像的
func (c *Content) getTitleByOrigin() string {
	if !fun.Blank(c.OriginTitle) {
		headlines := c.Doc.Find("h1,h2,h3")
		if headlines.Size() > 0 {
			titles := make([]string, 0)
			titleSim := make([]float64, 0)
			headlines.Each(func(i int, headline *goquery.Selection) {
				text := fun.NormaliseSpace(headline.Text())
				sim := fun.SimilarityText(c.OriginTitle, text)
				if sim > 0.3 {
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

	// 优先使用 originTitle 判定 headline
	title = c.getTitleByOrigin()
	if title != "" {
		c.titlePos = "headline"
		return title
	}

	// 页面 Title
	titleList := make([]*html.Node, 0)
	titleSim := make([]float64, 0)
	originMetaTitle := WebTitle(c.Doc, 255)

	// 去除原始 metaTitle 最后一个尾巴（一般是站点名称），再进行相似判断
	metaTitle := WebContentTitleClean(originMetaTitle, c.Lang)

	if !fun.Blank(originMetaTitle) && contentNode != nil {

		// 从 Meta 中提取相似 <title> 的标题，优先级较高，返回短的那个
		titleByMeta := c.getTitleByMeta(metaTitle)
		if titleByMeta != "" {
			return titleByMeta
		}

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

			// 这里他设置了一个 (i+1) 的权重因子，大意是越靠近后面权重越高
			for i := 0; i < index; i++ {
				score := titleSim[i] * float64(i+1)
				if score > maxScore {
					maxScore = score
					maxIndex = i
				}
			}

			if maxIndex != -1 {
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
		if sim > 0.3 {
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

// getTitleByEditDistance 从正文中找最相似 metaTitle 的片段，最坏的情况，返回页面标题
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
				if sim > 0.3 && sim > max[0] {
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
	for _, metaSelector := range metaTitleSelectors {
		title := strings.TrimSpace(c.Doc.Find(metaSelector).AttrOr("content", ""))
		if !fun.Blank(title) {
			titles = append(titles, title)
		}
	}

	if len(titles) > 0 {
		for _, title := range titles {
			sim := fun.SimilarityText(title, metaTitle)
			if sim > 0.3 {
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
	}

	return ""
}

func (c *Content) computeInfo(node *html.Node) CountInfo {
	if node.Type == html.ElementNode {
		countInfo := CountInfo{}
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
		countInfo := CountInfo{}

		text := fun.NormaliseSpace(node.Data)
		textLen := utf8.RuneCountInString(text)
		countInfo.TextCount = textLen
		countInfo.LeafList = append(countInfo.LeafList, textLen)

		return countInfo
	} else {
		return CountInfo{}
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
