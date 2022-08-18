package extract

import (
	"net/url"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

const (
	LinkTypeNone    LinkType = 0
	LinkTypeContent LinkType = 1
	LinkTypeList    LinkType = 2
	LinkTypeUnknown LinkType = 3

	RegexPublishDate = `20[2-3]\d{1}(0[1-9]|1[0-2]|[1-9])(0[1-9]|[1-2][0-9]|3[0-1]|[1-9])?`
)

var (
	zhPuncs   = []string{"，", "。", "；", "：", "？", "！", "（", "）", "“", "”"}
	wordLangs = []string{"en", "ru", "ar", "de", "fr", "es", "pt"}

	zhEnTitles = []string{"nba", "cba", "5g", "ai", "it", "ipo"}

	regexPublishDatePattern = regexp.MustCompile(RegexPublishDate)
	regexZhPattern          = regexp.MustCompile(`\p{Han}`)
	regexEnPattern          = regexp.MustCompile(`[a-zA-Z]`)
	regexPuncPattern        = regexp.MustCompile(`\pP`)
)

type LinkType int

type LinkTypeRule map[string][]string

type LinkRes struct {
	Content map[string]string
	List    map[string]string
	Unknown map[string]string
	None    map[string]string
}

// LinkTypes 返回链接分类结果
func LinkTypes(linkTitles map[string]string, lang string, rules LinkTypeRule) (*LinkRes, fun.StringSet) {
	linkRes := &LinkRes{
		Content: make(map[string]string),
		List:    make(map[string]string),
		Unknown: make(map[string]string),
		None:    make(map[string]string),
	}

	subDomains := make(map[string]bool)

	// 统计数据
	var contentPublishCount int
	contentTopPaths := make(map[string]int)

	for link, title := range linkTitles {
		if linkUrl, err := fun.UrlParse(link); err == nil {
			hostname := linkUrl.Hostname()
			domainTop := DomainTop(hostname)
			if hostname != domainTop {
				subDomains[hostname] = true
			}

			if rules == nil {
				linkType := LinkIsContentByTitle(linkUrl, title, lang)
				switch linkType {
				case LinkTypeContent:
					linkRes.Content[link] = title

					// 内容页 URL path 时间特征统计
					pathDir := path.Dir(strings.TrimSpace(linkUrl.Path))
					pathClean := pathDirClean(pathDir)
					if regexPublishDatePattern.MatchString(pathClean) {
						contentPublishCount++
					}

					// 内容页 URL path 统计
					paths := fun.SplitTrim(pathDir, fun.SLASH)
					if len(paths) > 0 {
						pathIndex := paths[0]
						contentTopPaths[pathIndex]++
					}
				case LinkTypeList:
					linkRes.List[link] = title
				case LinkTypeNone:
					linkRes.None[link] = title
				case LinkTypeUnknown:
					linkRes.Unknown[link] = title
				}
			} else {
				if LinkIsContentByRegex(linkUrl, rules) {
					linkRes.Content[link] = title
				} else {
					linkRes.List[link] = title
				}
			}
		}
	}

	// 基于内容页 URL path 特征统计与分类
	if rules == nil {
		linkRes = linkTypePathProcess(linkRes, contentTopPaths, contentPublishCount)
	}

	return linkRes, subDomains
}

func linkTypePathProcess(linkRes *LinkRes, contentTopPaths map[string]int, contentPublishCount int) *LinkRes {
	// 统计
	contentCount := len(linkRes.Content)
	listCount := len(linkRes.List)
	unknownCount := len(linkRes.Unknown)

	// 内容页 URL path 发布时间特征比例
	publishProb := float32(contentPublishCount) / float32(contentCount)

	// 内容页 URL path 占比较多的特征, 只取 Top 2
	topPaths := make([]string, 0)
	if contentCount >= 8 {
		for topPath, stat := range contentTopPaths {
			if stat > 1 {
				prob := float32(stat) / float32(contentCount)
				if prob > 0.4 {
					topPaths = append(topPaths, topPath)
				}
			}
		}
	}

	// 内容页 URL path 具有明显的发布时间特征比例, 处理 List、Unknown
	if publishProb > 0.7 {
		if listCount > 0 {
			for link, title := range linkRes.List {
				linkUrl, _ := fun.UrlParse(link)
				pathDir := path.Dir(strings.TrimSpace(linkUrl.Path))
				pathClean := pathDirClean(pathDir)
				if regexPublishDatePattern.MatchString(pathClean) {
					linkRes.Content[link] = title
					delete(linkRes.List, link)
				}
			}
		}
		if unknownCount > 0 {
			for link, title := range linkRes.Unknown {
				linkUrl, _ := fun.UrlParse(link)
				pathDir := path.Dir(strings.TrimSpace(linkUrl.Path))
				pathClean := pathDirClean(pathDir)
				if regexPublishDatePattern.MatchString(pathClean) {
					linkRes.Content[link] = title
				} else {
					linkRes.List[link] = title
				}
				delete(linkRes.Unknown, link)
			}
		}
	} else if len(topPaths) > 0 && unknownCount > 0 {
		// 内容页 URL path 具有前缀特征, 处理 Unknown
		for link, title := range linkRes.Unknown {
			linkUrl, _ := fun.UrlParse(link)

			pathDir := path.Dir(strings.TrimSpace(linkUrl.Path))
			paths := fun.SplitTrim(pathDir, fun.SLASH)
			if len(paths) > 0 {
				pathIndex := paths[0]
				if fun.SliceContains(topPaths, pathIndex) {
					linkRes.Content[link] = title
				} else {
					linkRes.List[link] = title
				}
				delete(linkRes.Unknown, link)
			}
		}
	}

	// path 具有特征, 清洗一下内容页中无 path 的
	if contentCount > 0 && (publishProb > 0.7 || len(topPaths) > 0) {
		for link, title := range linkRes.Content {
			linkUrl, _ := fun.UrlParse(link)
			pathStr := strings.TrimSpace(linkUrl.Path)
			pathDir := path.Dir(pathStr)
			paths := fun.SplitTrim(pathDir, fun.SLASH)
			if pathStr == "" || pathStr == "/" || len(paths) == 0 {
				linkRes.Unknown[link] = title
				delete(linkRes.Content, link)
			}
		}
	}

	return linkRes
}

func LinkIsContentByRegex(linkUrl *url.URL, rules LinkTypeRule) bool {
	hostname := linkUrl.Hostname()
	domainTop := DomainTop(hostname)

	if _, exist := rules[hostname]; exist {
		for _, regex := range rules[hostname] {
			if fun.Matches(linkUrl.String(), regex) {
				return true
			}
		}
	} else if _, exist := rules[domainTop]; exist {
		for _, regex := range rules[domainTop] {
			if fun.Matches(linkUrl.String(), regex) {
				return true
			}
		}
	}

	return false
}

func LinkIsContentByTitle(linkUrl *url.URL, title string, lang string) LinkType {
	link := linkUrl.String()

	if utf8.RuneCountInString(link) > 255 {
		return LinkTypeNone
	}

	// 无 path 或者默认 path, 应当由 domain 处理
	pathDir := strings.TrimSpace(linkUrl.Path)
	if pathDir == "" || pathDir == fun.SLASH || pathDir == "/index.html" || pathDir == "/index.htm" || pathDir == "/index.shtml" {
		return LinkTypeNone
	}

	if lang == "zh" {
		// 中文
		zhs := regexZhPattern.FindAllString(title, -1)
		hanCount := len(zhs)

		// 必须包含中文才可能是内容页
		if hanCount > 0 {
			// 内容页标题中文大于 5
			if hanCount > 5 {

				// 去掉空格
				title = strings.ReplaceAll(title, fun.SPACE, "")
				titleLen := utf8.RuneCountInString(title)

				// >= 8 判定为内容页 URL
				if titleLen >= 8 {
					return LinkTypeContent
				} else if titleLen < 8 {

					// 如果是中文, 判断是否包含常用标点
					if lang == "zh" {
						if fun.ContainsAny(title, zhPuncs...) {
							return LinkTypeContent
						}
					}
					return LinkTypeUnknown
				}
			} else {
				return LinkTypeList
			}
		} else {
			// 没有中文, 简单匹配英文字典
			if fun.SliceContains(zhEnTitles, strings.ToLower(title)) {
				return LinkTypeList
			}

			return LinkTypeNone
		}

	} else if fun.SliceContains(wordLangs, lang) {
		// 英语等单词类的语种

		// 去掉所有标点
		title = regexPuncPattern.ReplaceAllString(title, "")

		ens := regexEnPattern.FindAllString(title, -1)
		enCount := len(ens)

		// 必须包含英文字母
		if enCount > 0 {
			// 按照空格切分计算长度
			words := fun.SplitTrim(title, fun.SPACE)

			if len(words) >= 5 {
				return LinkTypeContent
			} else {
				return LinkTypeList
			}
		} else {
			return LinkTypeNone
		}
	} else {
		// 其他语种, 去除标点, 计算长度
		title = regexPuncPattern.ReplaceAllString(title, "")

		titleLen := utf8.RuneCountInString(title)
		if titleLen >= 10 {
			return LinkTypeContent
		} else if titleLen < 10 {
			// TODO 其他规则
			return LinkTypeList
		}
	}

	return LinkTypeNone
}

func pathDirClean(pathDir string) string {
	pathClean := strings.ReplaceAll(pathDir, fun.SLASH, "")
	pathClean = strings.ReplaceAll(pathClean, fun.DOT, "")
	pathClean = strings.ReplaceAll(pathClean, fun.DASH, "")
	pathClean = strings.ReplaceAll(pathClean, fun.UNDERSCORE, "")

	return pathClean
}
