package spider

import (
	"regexp"
	"strings"
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
		m := regexp.MustCompile(`\p{Han}`)
		zhs := m.FindAllString(title, -1)
		hanCount := len(zhs)

		if hanCount > 0 {
			if hanCount > 4 {
				title = strings.ReplaceAll(title, fun.SPACE, "")
				titleLen := utf8.RuneCountInString(title)
				if titleLen > zhMinTitle {
					return true
				}
			}
		}
	}

	return false
}
