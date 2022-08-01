package extract

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

type LinkType int

const (
	LinkTypeNone    LinkType = 0
	LinkTypeContent LinkType = 1
	LinkTypeList    LinkType = 2
)

type LinkRes struct {
	Content map[string]string
	List    map[string]string
	None    map[string]string
}

type SubDomainRes map[string]bool

// LinkTypes 返回链接分组
func LinkTypes(linkTitles map[string]string, lang string, regex string) (*LinkRes, SubDomainRes) {
	linkRes := &LinkRes{
		Content: make(map[string]string),
		List:    make(map[string]string),
		None:    make(map[string]string),
	}

	subDomains := make(map[string]bool)

	for link, title := range linkTitles {
		u, err := fun.UrlParse(link)
		if err == nil {
			subDomain := u.Hostname()
			domainTop := DomainTop(subDomain)
			if subDomain != domainTop {
				subDomains[subDomain] = true
			}
		}

		if regex == "" {
			linkType := LinkIsContentByLang(link, title, lang)
			switch linkType {
			case LinkTypeContent:
				linkRes.Content[link] = title
			case LinkTypeList:
				linkRes.List[link] = title
			case LinkTypeNone:
				linkRes.None[link] = title
			}
		} else {
			if LinkIsContentByRegex(link, regex) {
				linkRes.Content[link] = title
			} else {
				linkRes.List[link] = title
			}
		}
	}

	return linkRes, subDomains
}

func LinkIsContentByRegex(link string, regex string) bool {
	if !fun.Blank(regex) {
		if fun.Matches(link, regex) {
			return true
		}
	}

	return false
}

func LinkIsContentByLang(link string, title string, lang string) LinkType {
	tokenLangs := []string{"ja", "ko"}
	wordLangs := []string{"en", "ru", "ar", "de", "fr", "it", "es", "pt"}

	if lang == "zh" {
		m := regexp.MustCompile(`\p{Han}`)
		zhs := m.FindAllString(title, -1)
		hanCount := len(zhs)

		// 必须包含中文
		if hanCount > 0 {
			// 内容页标题中文大于4
			if hanCount > 4 {
				// 去掉空格
				title = strings.ReplaceAll(title, fun.SPACE, "")
				titleLen := utf8.RuneCountInString(title)
				// >= 8 判定为内容页 URL
				if titleLen >= 8 {
					return LinkTypeContent
				} else if titleLen < 8 {
					// 包含常用标点
					if fun.ContainsAny(title, "，", "。", "；", "：", "？", "！", "（", "）", "《", "》", "“", "”") {
						return LinkTypeContent
					} else {
						// TODO: 根据 URL 特征判断
						return LinkTypeList
					}
				}
			} else {
				return LinkTypeList
			}
		} else {
			return LinkTypeNone
		}

		// 类似中文的语种
	} else if fun.SliceContains(tokenLangs, lang) {
		// 去掉空格
		title = strings.ReplaceAll(title, fun.SPACE, "")
		titleLen := utf8.RuneCountInString(title)
		// >= 8 判定为内容页 URL
		if titleLen >= 8 {
			return LinkTypeContent
		} else if titleLen < 8 {
			// TODO 其他规则
			return LinkTypeList
		} else if titleLen < 2 {
			return LinkTypeNone
		}
		// 单词类的语种
	} else if fun.SliceContains(wordLangs, lang) {
		// 去掉所有标点
		m := regexp.MustCompile(`\pP`)
		title = m.ReplaceAllString(title, "")

		words := fun.SplitTrim(title, fun.SPACE)
		if len(words) >= 5 {
			return LinkTypeContent
		} else {
			return LinkTypeList
		}
	} else {
		titleLen := utf8.RuneCountInString(title)
		if titleLen >= 8 {
			return LinkTypeContent
		} else if titleLen < 8 {
			// TODO 其他规则
			return LinkTypeList
		}
	}

	return LinkTypeNone
}
