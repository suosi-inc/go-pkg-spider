// Package extract 正文抽取, 针对论文 CEPF 算法实现做了一些优化
// 基于标签路径特征融合的在线 Web 新闻内容抽取, (吴共庆等), Refer to: http://www.jos.org.cn/jos/article/abstract/4868
package extract

import (
	"bytes"
	"log"
	"math"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
	"golang.org/x/net/html"
)

const (
	ContentRemoveTags = "textarea"
)

var (
	metaTitleSelectors = []string{
		"meta[property='og:title' i]",
		"meta[property='twitter:title' i]",
		"meta[name='twitter:title' i]",
	}
)

type News struct {
	// 标题
	Title string
	// 标题依据
	TitlePos string
	// 发布时间
	Time string
	// 发布时间依据
	TimePos string
	// 正文纯文本
	Content string
	// 正文节点
	ContentNode *html.Node
}

type Content struct {
	Doc     *goquery.Document
	Lang    string
	InfoMap map[*html.Node]CountInfo

	BodyNode *html.Node
	TitlePos string
	TimePos  string
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

func NewContent(doc *goquery.Document, lang string) *Content {
	infoMap := make(map[*html.Node]CountInfo, 0)
	return &Content{Doc: doc, Lang: lang, InfoMap: infoMap}
}

func (c *Content) News() *News {
	news := &News{}

	// 提取内容根结点
	contentNode := c.getContentNode()
	if contentNode != nil {
		news.ContentNode = contentNode

		node := goquery.NewDocumentFromNode(contentNode)
		content := c.normaliseText(node.Selection)
		news.Content = content
	}

	title := c.getTitle(contentNode)
	news.Title = title
	news.TitlePos = c.TitlePos

	return news
}

func (c *Content) getContentNode() *html.Node {
	var maxScore float64
	var contentNode *html.Node

	// 取第一个 body 标签
	c.Doc.Find(ContentRemoveTags).Remove()
	bodyNodes := c.Doc.Find("body").Nodes
	if len(bodyNodes) > 0 {
		bodyNode := bodyNodes[0]
		c.BodyNode = bodyNode

		// 遍历计算统计最后找得分最高节点
		c.computeInfo(c.BodyNode)

		// c.debug()

		for node, _ := range c.InfoMap {
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

	return ""
}

func (c *Content) getTitle(contentNode *html.Node) string {
	var title string
	var contentIndex int32
	titleList := make([]*html.Node, 0)
	titleSim := make([]float64, 0)
	originMetaTitle := WebTitle(c.Doc, 255)

	if !fun.Blank(originMetaTitle) {
		// 去除原始 metaTitle 最后一个尾巴（一般是站点名称），再进行相似判断
		metaTitle := WebContentTitleClean(originMetaTitle, c.Lang)

		// 从 Meta 中提取相似 <title> 的标题，优先级较高，返回短的那个
		titleByMeta := c.getTitleByMeta(metaTitle)
		if titleByMeta != "" {
			return titleByMeta
		}

		// 从 body 开始遍历，一直遍历到内容区域，收集 h1->h2，并计算与 metaTitle 的相似度
		var traverse func(*html.Node)
		traverse = func(n *html.Node) {

			if n.FirstChild != nil {
				if n.Type == html.ElementNode {
					if contentNode != nil && n == contentNode {
						atomic.StoreInt32(&contentIndex, int32(len(titleList)))
						return
					}

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
		traverse(c.BodyNode)

		// 从 h 标签中获取
		index := int(atomic.LoadInt32(&contentIndex))
		if index > 0 {
			var maxScore float64
			var maxIndex int
			maxIndex = -1

			// 这里他设置了一个 (i+1) 的权重因子，大意是越靠近内容区域权重越高
			for i := 0; i < index; i++ {
				score := titleSim[i] * float64(i+1)
				if score > maxScore {
					maxScore = score
					maxIndex = i
				}
			}

			if maxIndex != -1 {
				c.TitlePos = "headline"
				tagNode := goquery.NewDocumentFromNode(titleList[maxIndex])
				headTitle := c.normaliseText(tagNode.Selection)
				return headTitle
			}
		}
	}

	// 如果没有相似的，则尝试从包含开头结尾 title 选择器中获取一个
	titles := c.Doc.Find("body").Find("*[id^=title],*[id$=title],*[class^=title],*[class$=title]")
	if titles.Size() > 0 {
		first := titles.First()
		selectorTitle := c.normaliseText(first)
		titleLen := utf8.RuneCountInString(selectorTitle)
		if titleLen > 5 && titleLen < 40 {
			c.TitlePos = "selector"
			return selectorTitle
		}
	}

	// 最后从正文中找最相似 metaTitle 的片段，最坏的情况，返回页面标题
	title = c.getTitleByEditDistance(originMetaTitle)

	return title
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
				if sim > 0 {
					if sim > max[0] {
						max[0] = sim
						buf.Reset()
						buf.WriteString(text)
					}
				}
			}

			for child := n.FirstChild; child != nil; child = child.NextSibling {
				traverse(child)
			}
		}
	}
	traverse(c.BodyNode)

	if len(buf.String()) > 0 {
		c.TitlePos = "text"
		return buf.String()
	}

	c.TitlePos = "title"
	return originMetaTitle
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
			if sim > 0.382 {
				titleLen := utf8.RuneCountInString(title)
				metaTitleLen := utf8.RuneCountInString(metaTitle)
				if titleLen < metaTitleLen {
					c.TitlePos = "title"
					return title
				} else {
					c.TitlePos = "metaTitle"
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

		c.InfoMap[node] = countInfo

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
	countInfo := c.InfoMap[node]
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

func (c *Content) debug() {
	for node, info := range c.InfoMap {
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
