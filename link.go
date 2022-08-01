package spider

import (
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

const (
	TokenTitleLen = 8
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
	langSlices := []string{"zh", "ja", "ko", "ar", "hi", "th", "vi", "id"}

	if fun.SliceContains(langSlices, lang) {
		title = fun.RemoveSign(title)
		titleLen := utf8.RuneCountInString(title)
		if lang == "zh" {
			// 至少含有中文
			if fun.Matches(title, `\p{Han}`) {
				if titleLen < TokenTitleLen {
					return true
				}
			}
		} else {
			if titleLen >= TokenTitleLen {
				return true
			}
		}
	}

	return false
}
