package extract

import (
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider"
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

func LinkTitles(doc *goquery.Document, urlStr string, strictDomain bool) map[string]string {
	var linkTitles = make(map[string]string, 0)

	// 当前请求的 urlStr
	baseUrl, err := fun.UrlParse(urlStr)
	if err != nil {
		return linkTitles
	}

	// 获取所有 a 链接
	aTags := doc.Find("a")
	if aTags.Size() > 0 {
		var tmpLinks = make(map[string]string, 0)

		// 提取所有的 a 链接
		aTags.Each(func(i int, s *goquery.Selection) {
			tmpLink, exists := s.Attr("href")
			if exists {
				tmpLink = fun.RemoveLines(tmpLink)
				tmpLink = strings.TrimSpace(tmpLink)

				tmpTitle := s.Text()
				tmpLink = fun.RemoveLines(tmpLink)
				tmpTitle = strings.ReplaceAll(tmpTitle, "  ", "")
				tmpTitle = strings.TrimSpace(tmpTitle)
				if tmpLink != "" && tmpTitle != "" {
					tmpLinks[tmpLink] = tmpTitle
				}
			}
		})

		// 返回链接
		if len(tmpLinks) > 0 {
			// 过滤掉非法链接
			for link, title := range tmpLinks {
				if a, err := filterUrl(link, baseUrl, strictDomain); err == nil {
					linkTitles[a] = title

				} else {
					log.Println("@@@", a, err)
				}
			}
		}
	}

	return linkTitles
}

func filterUrl(link string, baseUrl *url.URL, strictDomain bool) (string, error) {
	var urlStr string

	// 过滤掉不太正常的链接
	if fun.ContainsAny(link, "{", "}", "[", "]", "@", "$", "<", ">", "\"") {
		return link, errors.New("invalid url with illegal characters")
	}

	// 转换为绝对路径
	if !fun.HasPrefixCase(link, "http") && !fun.HasPrefixCase(link, "https") {
		if l, err := baseUrl.Parse(link); err == nil {

			absoluteUrl := l.String()

			// 验证连接是否合法
			if u, err := fun.UrlParse(absoluteUrl); err == nil {
				urlStr = u.String()
			} else {
				return absoluteUrl, errors.New("invalid url with parse error")
			}
		} else {
			return link, errors.New("invalid url with baseUrl parse")
		}
	} else {
		urlStr = link
	}

	// 验证 url 是否合法
	if !fun.IsAbsoluteUrl(urlStr) {
		return urlStr, errors.New("invalid url with absolute url")
	}

	// 限制链接为本站
	if strictDomain {
		if u, err := fun.UrlParse(urlStr); err == nil {
			hostname := u.Hostname()
			baseDomainTop := spider.DomainTop(baseUrl.Hostname())
			if hostname != baseDomainTop && !fun.HasSuffixCase(hostname, "."+baseDomainTop) {
				return urlStr, errors.New("invalid url with strict domain")
			}
		} else {
			return urlStr, errors.New("invalid url with url parse by strict domain")
		}
	}

	return urlStr, nil
}
