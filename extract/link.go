package extract

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

var (
	zhPuncs   = []string{"，", "。", "；", "：", "？", "！", "（", "）", "《", "》", "“", "”"}
	wordLangs = []string{"en", "ru", "ar", "de", "fr", "it", "es", "pt"}
)

type LinkRes struct {
	Content map[string]string
	List    map[string]string
	None    map[string]string
}

type LinkTypeRule map[string][]string

type LinkType int

const (
	LinkTypeNone    LinkType = 0
	LinkTypeContent LinkType = 1
	LinkTypeList    LinkType = 2
)

// LinkTypes 返回链接分类结果
func LinkTypes(linkTitles map[string]string, lang string, rules LinkTypeRule) (*LinkRes, fun.StringSet) {
	linkRes := &LinkRes{
		Content: make(map[string]string),
		List:    make(map[string]string),
		None:    make(map[string]string),
	}

	subDomains := make(map[string]bool)

	for link, title := range linkTitles {
		u, err := fun.UrlParse(link)
		if err == nil {
			hostname := u.Hostname()
			domainTop := DomainTop(hostname)
			if hostname != domainTop {
				subDomains[hostname] = true
			}
		}

		if rules == nil {
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
			if LinkIsContentByRegex(link, rules) {
				linkRes.Content[link] = title
			} else {
				linkRes.List[link] = title
			}
		}
	}

	return linkRes, subDomains
}

func LinkIsContentByRegex(link string, rules LinkTypeRule) bool {
	u, err := fun.UrlParse(link)
	if err == nil {
		hostname := u.Hostname()
		domainTop := DomainTop(hostname)

		if _, exist := rules[hostname]; exist {
			for _, regex := range rules[hostname] {
				if fun.Matches(link, regex) {
					return true
				}
			}
		} else if _, exist := rules[domainTop]; exist {
			for _, regex := range rules[domainTop] {
				if fun.Matches(link, regex) {
					return true
				}
			}
		}
	}

	return false
}

func LinkIsContentByLang(link string, title string, lang string) LinkType {
	if lang == "zh" {
		// 中文

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
					if fun.ContainsAny(title, zhPuncs...) {
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

	} else if fun.SliceContains(wordLangs, lang) {
		// 英语等单词类的语种

		// 去掉所有标点
		m := regexp.MustCompile(`\pP`)
		title = m.ReplaceAllString(title, "")

		// 按照空格切分计算长度
		words := fun.SplitTrim(title, fun.SPACE)
		if len(words) >= 5 {
			return LinkTypeContent
		} else {
			return LinkTypeList
		}
	} else {
		// 其他语种，去除标点，计算长度
		m := regexp.MustCompile(`[\pP]`)
		title = m.ReplaceAllString(title, "")

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
