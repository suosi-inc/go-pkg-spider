package spider

import (
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

const (
	zhMinTitle = 7
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
		// 至少含有中文
		if fun.Matches(title, `\p{Han}`) {
			title = fun.RemoveSign(title)
			titleLen := utf8.RuneCountInString(title)
			if titleLen > zhMinTitle {
				return true
			}
		}
	}

	return false
}
