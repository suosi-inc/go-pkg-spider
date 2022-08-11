package extract

import (
	"github.com/PuerkitoBio/goquery"
)

var (
	metaTitleSelectors = []string{
		"meta[property=og:title]",
		"meta[property=twitter:title]",
		"meta[name=twitter:title]",
	}
)

type Article struct {
	Title string

	TitlePos string
}

type Content struct {
	Doc  *goquery.Document
	Lang string
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
	pCount int

	// 叶子列表
	LeafList []int
}

func NewContent(doc *goquery.Document, lang string) *Content {
	return &Content{Doc: doc, Lang: lang}
}

func (c *Content) Content() (string, error) {
	return c.Doc.Text(), nil
}
