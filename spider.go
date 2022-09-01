package spider

import (
	"strings"

	"github.com/x-funs/go-fun"
)

var (
	DefaultDocRemoveTags = "script,noscript,style,iframe,br,link,svg"
)

func NormaliseLine(str string) string {
	lines := fun.SplitTrim(str, fun.LF)
	if len(lines) > 0 {
		str = strings.Join(lines, fun.LF)
	}

	return str
}
