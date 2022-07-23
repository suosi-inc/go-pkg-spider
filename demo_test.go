package spider

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGoquery(t *testing.T) {
	body, _ := HttpGet("https://www.163.com/news/article/HCB7Q3LA000189FH.html")
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	// lang, exist := doc.Find("html").Attr("id")
	text := doc.Find("body").Text()
	fmt.Println(text)
}
