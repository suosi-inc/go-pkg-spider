package extract

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

var (
	regexUrlPattern = regexp.MustCompile(fun.RegexUrl)
)

// WebTitle 返回网页标题, 最大 255 个字符
func WebTitle(doc *goquery.Document, maxLength int) string {
	title := doc.Find("title").Text()
	title = strings.TrimSpace(title)

	if maxLength > 0 && maxLength < 255 {
		return fun.SubString(title, 0, maxLength)
	} else {
		return fun.SubString(title, 0, 255)
	}
}

// WebTitleClean 返回尽量清洗后的网页标题
func WebTitleClean(title string, lang string) string {
	zhSplits := []string{"_", "|", "-", "－", "｜"}
	enSplits := []string{" - ", " | "}

	// 中文网站，查找中文网站的分割标记，找到任意一个，从尾部循环删除后返回
	if lang == "zh" {
		for _, split := range zhSplits {
			end := strings.LastIndex(title, split)
			if end != -1 {
				titleClean := title

				for {
					titleClean = strings.TrimSpace(titleClean[:end])
					end = strings.LastIndex(titleClean, split)

					if end == -1 {
						break
					}
				}

				return titleClean
			}
		}

		// 其他，查找英文分割标记，如果找到，从尾部删除一次返回
	} else {
		for _, split := range enSplits {
			end := strings.LastIndex(title, split)
			if end != -1 {
				titleClean := strings.TrimSpace(title[:end])
				return titleClean
			}
		}
	}

	return title
}

// WebKeywords 返回网页 Keyword
func WebKeywords(doc *goquery.Document) string {
	keywords := doc.Find("meta[name=keywords]").AttrOr("content", "")
	keywords = strings.TrimSpace(keywords)
	return keywords
}

// WebDescription 返回网页描述
func WebDescription(doc *goquery.Document) string {
	description := doc.Find("meta[name=description]").AttrOr("content", "")
	description = strings.TrimSpace(description)
	return description
}

// WebLinkTitles 返回网页链接和锚文本
func WebLinkTitles(doc *goquery.Document, urlStr string, strictDomain bool) map[string]string {
	var linkTitles = make(map[string]string)

	// 当前请求的 urlStr
	baseUrl, err := fun.UrlParse(urlStr)
	if err != nil {
		return linkTitles
	}

	// 获取所有 a 链接
	aTags := doc.Find("a")
	if aTags.Size() > 0 {
		var tmpLinks = make(map[string]string)

		// 提取所有的 a 链接
		aTags.Each(func(i int, s *goquery.Selection) {
			tmpLink, exists := s.Attr("href")
			if exists {
				tmpLink = fun.RemoveLines(tmpLink)
				tmpLink = strings.TrimSpace(tmpLink)

				tmpTitle := s.Text()
				tmpTitle = fun.RemoveLines(tmpTitle)
				tmpTitle = strings.ReplaceAll(tmpTitle, fun.TAB, "")
				tmpTitle = strings.TrimSpace(tmpTitle)
				if tmpLink != "" && tmpTitle != "" {
					tmpLinks[tmpLink] = tmpTitle
				}
			}
		})

		// 返回链接
		if len(tmpLinks) > 0 {
			// 过滤链接
			for link, title := range tmpLinks {
				if a, err := filterUrl(link, baseUrl, strictDomain); err == nil {
					linkTitles[a] = title
				} else {
					//log.Println("@@@", a, err)
				}
			}
		}
	}

	return linkTitles
}

// filterUrl 过滤 url
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

	// 验证 url 是否合法 TODO: 性能问题
	//if !regexUrlPattern.MatchString(urlStr) {
	//	return urlStr, errors.New("invalid url with absolute url")
	//}

	// 限制链接为站内链接
	if strictDomain {
		if u, err := fun.UrlParse(urlStr); err == nil {
			hostname := u.Hostname()
			baseDomainTop := DomainTop(baseUrl.Hostname())
			if hostname != baseDomainTop && !fun.HasSuffixCase(hostname, "."+baseDomainTop) {
				return urlStr, errors.New("invalid url with strict domain")
			}
		} else {
			return urlStr, errors.New("invalid url with url parse by strict domain")
		}
	}

	return urlStr, nil
}
