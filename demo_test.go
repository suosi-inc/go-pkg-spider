package spider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

func TestGoquery(t *testing.T) {
	body, _ := fun.HttpGet("https://www.163.com/news/article/HCB7Q3LA000189FH.html")
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(body))

	lang, exist := doc.Find("html").Attr("id")
	fmt.Println(lang, exist)
}
