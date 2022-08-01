package extract

import (
	"fmt"
	"regexp"
	"testing"
)

func TestMatch(t *testing.T) {
	m := regexp.MustCompile(`\p{Han}`)
	allString := m.FindAllString("123你好，世界asdf", -1)
	fmt.Println(allString)
}
