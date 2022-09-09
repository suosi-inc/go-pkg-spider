package spider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/x-funs/go-fun"
)

type News struct {
	url        string
	subDomains []string
	contents   []string
	depth      uint8
	seen       map[string]bool
	isSub      bool
}

type NewsData struct {
	// 标题
	Title string
	// 发布时间
	TimeLocal string
	// 时间
	Time string
	// 正文纯文本
	Content string
}

func (n *News) NewNews(domain string, depth uint8, isSub bool) *News {
	return &News{
		url:   domain,
		depth: depth,
		isSub: isSub,
	}
}

func (n *News) GetNews() []NewsData {
	// 获取首页url
	urlSlice := strings.Split(n.url, "/")
	domain := urlSlice[0] + "//" + urlSlice[2]

	if n.isSub {
		// 先探测出url主域名的所有子域名
		// 获取页面链接分组
		subDomains, err := GetSubdomains(domain, 20000, 1)
		if err != nil {
			fmt.Println("subDomain extract", err)
		}

		listSlice := []string{}
		contentSlice := []map[string]string{}
		subDomainSlice := []string{}

		for subDomain := range subDomains {
			subDomainSlice = append(subDomainSlice, subDomain)
		}

		for i := 0; i < int(n.depth); i++ {
			listS, contentS, _ := n.GetNewsLinkRes(subDomainSlice, 2000, 1)
			listSlice = append(listSlice, listS...)
			contentS = append(contentSlice, contentS...)
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

// GetNewsLinkRes 获取news页面链接分组, 仅返回列表页和内容页
func (n *News) GetNewsLinkRes(urls []string, timeout int, retry int) ([]string, []map[string]string, error) {
	listSlice := []string{}
	contentSlice := []map[string]string{}

	for _, url := range urls {
		if linkRes, _, _, err := GetLinkRes(url, timeout, retry); err != nil {
			for l := range linkRes.List {
				if !n.seen[l] {
					n.seen[l] = true
					listSlice = append(listSlice, l)
				}
			}

			for c, v := range linkRes.Content {
				if !n.seen[c] {
					n.seen[c] = true
					cc := map[string]string{}
					cc[c] = v
					contentSlice = append(contentSlice, cc)
				}
			}

		} else {
			fmt.Println("GetNewsLinkRes", err)
		}
	}

	return listSlice, contentSlice, nil
}

// GetSubdomains 获取subDomain
func GetSubdomains(domain string, timeout int, retry int) (fun.StringSet, error) {
	if _, _, subDomains, err := GetLinkRes(domain, timeout, retry); err == nil {
		return subDomains, nil
	} else {
		return nil, err
	}
}
