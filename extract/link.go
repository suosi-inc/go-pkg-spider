package extract

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

const (
	ZhMinTitleLen = 8
	LatinTitleLen = 4
)

func IsContentByRegex(link string, regex string) bool {
	if !fun.Blank(regex) {
		if fun.Matches(link, regex) {
			return true
		}
	}
	return false
}

func IsContentByLang(link string, title string, lang string) bool {
	if lang == "zh" {
		m := regexp.MustCompile(`\p{Han}`)
		zhs := m.FindAllString(title, -1)
		hanCount := len(zhs)

		// 必须包含中文
		if hanCount > 0 {
			// 大于4个中文内容页标题
			if hanCount > 4 {
				title = strings.ReplaceAll(title, fun.SPACE, "")
				titleLen := utf8.RuneCountInString(title)
				// 大于 ZhMinTitleLen 判定为内容页 URL
				if titleLen > ZhMinTitleLen {
					return true
				}
			}
		}
	}

	return false
}
