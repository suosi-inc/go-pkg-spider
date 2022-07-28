package extractor

import (
	"github.com/PuerkitoBio/goquery"
)

type TitleRes struct {
	Title    string
	TitlePos string
}

type NewsRes struct {
	Title string

	TitlePos string
}

var (
	metaTitleSelectors = []string{
		"meta[property=og:title]",
		"meta[property=twitter:title]",
		"meta[name=twitter:title]",
	}
)

type News struct {
	Doc *goquery.Document
}

func NewNews() *News {
	return &News{}
}
