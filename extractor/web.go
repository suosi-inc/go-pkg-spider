package extractor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

func Title(doc *goquery.Document, length int) string {
	title := doc.Find("title").Text()
	title = strings.TrimSpace(title)

	if length == 0 {
		return title
	} else {
		return fun.SubString(title, 0, length)
	}
}

func Keywords(doc *goquery.Document) string {
	keywords := doc.Find("meta[name=keywords]").AttrOr("content", "")
	keywords = strings.TrimSpace(keywords)
	return keywords
}

func Description(doc *goquery.Document) string {
	description := doc.Find("meta[name=description]").AttrOr("content", "")
	description = strings.TrimSpace(description)
	return description
}

// func LinkTitles(doc *goquery.Document, domain string, url string) map[string]string {
// 	var linkTitles = make(map[string]string, 0)
//
// 	aTags := doc.Find("a")
// 	if aTags.Size() > 0 {
// 		var tmpLinks map[string]string
//
// 		aTags.Each(func(i int, s *goquery.Selection) {
// 			tmpLink1, exists := s.Attr("href")
// 			if exists {
// 				tmpTitle1 := s.Text()
// 				href = strings.TrimSpace(href)
//
// 				if href != "" {
// 					if strings.HasPrefix(href, "http") {
// 						tmpLinks[href] = s.Text()
// 					} else {
// 						tmpLinks[fun.JoinUrl(domain, href)] = s.Text()
// 					}
// 				}
// 			}
// 		}
// 	}
//
// 	return linkTitles
// }
