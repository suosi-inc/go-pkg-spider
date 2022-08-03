package extract

import (
	"fmt"
	"net/url"
	"path"
	"testing"
	"unicode/utf8"

	"github.com/x-funs/go-fun"
)

func TestTitleClean(t *testing.T) {
	strs := map[string]string{
		"“暴徒试图杀死他！”阿拉木图市长在1月5日的暗杀企图中幸存_网易订阅":                                                   "zh",
		"“暴徒试图杀死他！”阿拉木图市长在1月5日的暗杀企图中幸存 - 网易订阅":                                                 "zh",
		"北极圈内最高温达到38℃ 北极熊还好吗？南极情况怎么样？_科技频道_中国青年网":                                              "zh",
		"About the Project on Nuclear Issues | Center for Strategic and International Studies": "en",
	}

	for str, l := range strs {
		t.Log(WebTitleClean(str, l))
	}
}

func TestUrlQuery(t *testing.T) {
	// urlStr := "https://people.com/tag/stories-to-make-you-smile/a/b/abc.html?a=1&b=2&c=3#ddd"
	urlStr := "https://vipmail.163.com/abc"
	u, err := url.Parse(urlStr)

	fmt.Println(err)
	fmt.Println(u.Path)
	fmt.Println(u.RawQuery)
	fmt.Println(path.Dir(u.Path))
	// fmt.Println(path.Base(u.Path))

	fmt.Println(utf8.RuneCountInString("https://adx.36kr.com/api/ad/click?sign=2eda7665240cec93f902311eb10c195a&param.redirectUrl=aHR0cHM6Ly8zNmtyLmNvbS9wLzE4NTM5NTQ2NzgxMzIzNTI&param.adsdk=Phid2i9VOob6U23ybkDx8q7cr1KbBDM4oiu1d_-C6gY5qf5SKxqBPsptEVMy_wtzqB5Yr08U7ioREUL7HLxIrQ"))
}

func TestFilterUrl(t *testing.T) {
	urlStr := "http://www.163.com/a/b/"
	baseUrl, _ := fun.UrlParse(urlStr)

	t.Log(filterUrl("./c/123.html", baseUrl, true))
	t.Log(filterUrl("../c/123.html", baseUrl, true))
	t.Log(filterUrl("/c/123.html", baseUrl, true))
	t.Log(filterUrl("//www.163.com/c/123.html", baseUrl, true))
	t.Log(filterUrl("//www.163.com/c/123.pdf?abc=1123", baseUrl, true))
}
