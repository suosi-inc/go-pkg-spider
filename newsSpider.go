package spider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/suosi-inc/go-pkg-spider/extract"
)

type News struct {
	url        string
	subDomains []string
	contents   []string
	depth      uint8
	seen       map[string]bool
	isSub      bool
}

func (n *News) NewNews(domain string, depth uint8, isSub bool) *News {
	return &News{
		url:   domain,
		depth: depth,
		isSub: isSub,
	}
}

func (n *News) GetNews() []extract.News {
	// 获取首页url
	urlSlice := strings.Split(n.url, "/")
	domain := urlSlice[0] + "//" + urlSlice[2]

	if n.isSub {
		// 先探测出url主域名的所有子域名
		// 获取页面链接分组
		if subDomains, err := GetSubdomains(domain, 20000, 1); err == nil {
			i := 0
			for a, title := range subDomains {
				i += 1
				fmt.Println(i, "subDomain:"+a+"\t=>\t"+strconv.FormatBool(title))
			}
		}

	} else {
		// 直接获取list和content
		if linkRes, _, subDomains, err := GetLinkRes(domain, 20000, 1); err == nil {
			fmt.Println("subDomain:", len(subDomains))
			fmt.Println("content:", len(linkRes.Content))
			fmt.Println("list:", len(linkRes.List))

			i := 0
			for a, title := range subDomains {
				i = i + 1
				fmt.Println(i, "subDomain:"+a+"\t=>\t"+strconv.FormatBool(title))
			}
			i = 0
			for a, title := range linkRes.Content {
				i = i + 1
				fmt.Println(i, "content:"+a+"\t=>\t"+title)
			}
			i = 0
			for a, title := range linkRes.List {
				i = i + 1
				fmt.Println(i, "list:"+a+"\t=>\t"+title)
			}
		}
	}

	return nil
}
