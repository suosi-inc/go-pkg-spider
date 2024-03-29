package extract

import (
	"errors"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

const (
	RegexHostnameIp = `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
)

var (
	filterUrlSuffix = []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".txt", ".xml",
		".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx",
		".zip", ".rar", ".7z", ".gz", ".apk", ".cgi", ".exe", ".bz2", ".play",
		".rss", ".sig", ".sgf",
		".mp3", ".mp4", ".rm", ".rmvb", ".mov", ".ogv", ".flv",
	}

	invalidUrlCharsets = []string{"{", "}", "[", "]", "@", "$", "<", ">", "\""}

	titleZhSplits = []string{"_", "|", "-", "－", "｜", "—", "＊", "：", ",", "，", ":", "·", ">>", "="}

	titleZhContentSplits = []string{"_", "|", "-", "－", "｜", "—"}

	titleEnSplits = []string{" - ", " | ", ":"}

	RegexHostnameIpPattern = regexp.MustCompile(RegexHostnameIp)
)

// WebTitle 返回网页标题, 最大 128 个字符
func WebTitle(doc *goquery.Document, maxLength int) string {
	var title string
	titleNode := doc.Find("title")
	if titleNode.Size() > 1 {
		// 竟然有多个 title, 只取第一个
		title = titleNode.First().Text()
	} else {
		title = titleNode.Text()
	}

	title = fun.RemoveLines(title)
	title = strings.TrimSpace(title)

	if maxLength > 0 && maxLength < 128 {
		return fun.SubString(title, 0, maxLength)
	} else {
		return fun.SubString(title, 0, 128)
	}
}

// WebTitleClean 返回尽量清洗后的网页标题
func WebTitleClean(title string, lang string) string {
	// 中文网站, 查找中文网站的分割标记, 找到任意一个, 从尾部循环删除后返回
	if lang == "zh" {

		for _, split := range titleZhSplits {
			if fun.HasPrefixCase(title, split) {
				title = fun.RemovePrefix(title, split)
			}
		}

		// 去除首页开头
		if fun.HasPrefixCase(title, "首页") {
			title = regexp.MustCompile("首页([ |\\-_－—｜])*").ReplaceAllString(title, "")
		}

		titleClean := title
		for _, split := range titleZhSplits {
			var exists bool
			end := strings.LastIndex(titleClean, split)
			if end != -1 {
				exists = true
				for {
					titleClean = strings.TrimSpace(titleClean[:end])
					end = strings.LastIndex(titleClean, split)

					if end == -1 {
						break
					}
				}
				if exists {
					break
				}
			}
		}

		// 去除尾巴
		if titleClean != "首页" {
			titleClean = fun.RemoveSuffix(titleClean, "首页")
		}

		titleClean = fun.RemoveSign(titleClean)

		return titleClean

	} else {
		// 其他, 查找英文分割标记, 如果找到, 从尾部删除一次返回
		for _, split := range titleEnSplits {
			end := strings.LastIndex(title, split)
			if end != -1 {
				titleClean := strings.TrimSpace(title[:end])
				return titleClean
			}
		}
	}

	return title
}

// WebContentTitleClean 返回内容页尽量清洗后的网页标题
func WebContentTitleClean(title string, lang string) string {
	// 中文网站, 查找中文网站的分割标记, 找到任意一个, 从尾部循环删除后返回
	if lang == "zh" {
		for _, split := range titleZhContentSplits {
			if fun.HasPrefixCase(title, split) {
				title = fun.RemovePrefix(title, split)
			}
		}

		titleClean := title
		for _, split := range titleZhContentSplits {
			var exists bool
			end := strings.LastIndex(titleClean, split)
			if end != -1 {
				exists = true
				for {
					titleClean = strings.TrimSpace(titleClean[:end])
					end = strings.LastIndex(titleClean, split)

					if end == -1 {
						break
					}
				}
				if exists {
					break
				}
			}
		}

		return titleClean

	} else {
		// 其他, 查找英文分割标记, 如果找到, 从尾部删除一次返回
		for _, split := range titleEnSplits {
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
	keywords := doc.Find("meta[name='keywords' i]").AttrOr("content", "")
	keywords = fun.RemoveLines(keywords)
	keywords = strings.TrimSpace(keywords)

	return keywords
}

// WebDescription 返回网页描述, 最大 384 个字符
func WebDescription(doc *goquery.Document, maxLength int) string {
	description := doc.Find("meta[name='description' i]").AttrOr("content", "")
	description = fun.RemoveLines(description)
	description = strings.TrimSpace(description)

	if maxLength > 0 && maxLength < 384 {
		return fun.SubString(description, 0, maxLength)
	} else {
		return fun.SubString(description, 0, 384)
	}
}

// WebLinkTitles 返回网页链接和锚文本
func WebLinkTitles(doc *goquery.Document, baseUrl *url.URL, strictDomain bool) (map[string]string, map[string]string) {
	var linkTitles = make(map[string]string)
	var filters = make(map[string]string)

	// 当前请求的 urlStr
	if baseUrl == nil {
		return linkTitles, filters
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
				tmpTitle = fun.NormaliseSpace(tmpTitle)
				tmpTitle = strings.TrimSpace(tmpTitle)
				if tmpLink != "" && tmpTitle != "" {
					// 如果链接已存在, 保留长标题
					if _, exists := tmpLinks[tmpLink]; exists {
						oldTitle := tmpLinks[tmpLink]
						if len(oldTitle) < len(tmpTitle) {
							tmpLinks[tmpLink] = tmpTitle
						}
					} else {
						tmpLinks[tmpLink] = tmpTitle
					}
				}
			}
		})

		// 过滤链接
		tmpLinkLen := len(tmpLinks)
		if tmpLinkLen > 0 {
			for link, title := range tmpLinks {
				if a, err := filterUrl(link, baseUrl, strictDomain); err == nil {
					linkTitles[a] = title
				} else {
					filters[a] = err.Error()
				}
			}
		}
	}

	return linkTitles, filters
}

// filterUrl 过滤 url
func filterUrl(link string, baseUrl *url.URL, strictDomain bool) (string, error) {
	var urlStr string

	// 过滤掉链接中包含特殊字符的
	if fun.ContainsAny(link, invalidUrlCharsets...) {
		return link, errors.New("invalid url with illegal characters")
	}

	// 转换为绝对路径
	if !fun.HasPrefixCase(link, "http") && !fun.HasPrefixCase(link, "https") {
		if l, err := baseUrl.Parse(link); err == nil {
			urlStr = l.String()
		} else {
			return link, errors.New("invalid url with baseUrl parse error")
		}
	} else {
		urlStr = link
	}

	// 解析验证
	u, err := fun.UrlParse(urlStr)
	if err != nil {
		return urlStr, errors.New("invalid url with parse error")
	}

	// 验证转换后是否是绝对路径
	if !u.IsAbs() {
		return urlStr, errors.New("invalid url with not absolute url")
	}

	// 验证非常规端口
	if u.Port() != "" {
		return urlStr, errors.New("invalid url with not 80 port")
	}

	// 验证主机名
	if RegexHostnameIpPattern.MatchString(u.Hostname()) {
		return urlStr, errors.New("invalid url with ip hostname")
	}

	// 过滤掉明显错误的后缀
	ext := path.Ext(u.Path)
	if strings.Contains(ext, ".") {
		ext = strings.ToLower(ext)
		if fun.SliceContains(filterUrlSuffix, ext) {
			return urlStr, errors.New("invalid url with suffix")
		}
	}

	// 过滤掉站外链接
	if strictDomain {
		hostname := u.Hostname()
		domainTop := DomainTop(hostname)
		baseDomainTop := DomainTop(baseUrl.Hostname())
		if domainTop != baseDomainTop {
			return urlStr, errors.New("invalid url with strict domain")
		}
	}

	return urlStr, nil
}
