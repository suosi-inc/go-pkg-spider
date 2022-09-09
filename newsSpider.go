package spider

import (
	"fmt"
	"strings"

	"github.com/x-funs/go-fun"
)

const (
	timeOut     = 20000
	retryTime   = 1
	retryAlways = 99
)

type News struct {
	url        string
	subDomains []string
	contents   []string
	depth      uint8
	seen       map[string]bool
	isSub      bool
	data       []*NewsData
	dataChan   chan *NewsData
}

type NewsData struct {
	// 链接
	Url string
	// 标题
	Title string
	// 发布时间
	Time string
	// 正文纯文本
	Content string
}

func NewNews(domain string, depth uint8, isSub bool) *News {
	return &News{
		url:      domain,
		depth:    depth,
		seen:     map[string]bool{},
		isSub:    isSub,
		data:     []*NewsData{},
		dataChan: make(chan *NewsData),
	}
}

func (n *News) GetNews(contentHandleFunc func(content map[string]string)) {
	// 初始化列表页和内容页切片
	listSlice := []string{}
	listSliceTemp := []string{}
	contentSlice := []map[string]string{}
	subDomainSlice := []string{}

	// 获取首页url和协议
	scheme, indexUrl := GetIndexUrl(n.url)

	if n.isSub {
		// 先探测出首页url的所有子域名
		subDomains, err := GetSubdomains(indexUrl, timeOut, retryAlways)
		if err != nil {
			fmt.Println("subDomain extract", err)
		}

		for subDomain := range subDomains {
			subDomainSlice = append(subDomainSlice, subDomain)
		}

		// 首次获取subDomain页
		listSliceTemp = subDomainSlice
	} else {
		listSliceTemp = append(listSliceTemp, n.url)
	}

	// 深度优先循环获取页面列表页和内容页
	for i := 0; i < int(n.depth); i++ {
		listS, contentS, _ := n.GetNewsLinkRes(contentHandleFunc, scheme, listSliceTemp, timeOut, retryTime)
		listSlice = append(listSlice, listS...)
		contentSlice = append(contentSlice, contentS...)

		// 重置循环列表页
		if len(listS) == 0 {
			break
		}
		listSliceTemp = listS
	}
}

// GetNewsLinkRes 获取news页面链接分组, 仅返回列表页和内容页
func (n *News) GetNewsLinkRes(contentHandleFunc func(content map[string]string), scheme string, urls []string, timeout int, retry int) ([]string, []map[string]string, error) {
	listSlice := []string{}
	contentSlice := []map[string]string{}

	for _, url := range urls {
		if !strings.Contains(url, "http") {
			url = scheme + url
		}
		if linkRes, _, _, err := GetLinkRes(url, timeout, retry); err == nil {
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
					// 内容页处理
					go contentHandleFunc(cc)
				}
			}

		} else {
			fmt.Println("GetNewsLinkRes", err)
		}
	}

	return listSlice, contentSlice, nil
}

func (n *News) GetData() []*NewsData {
	return n.data
}

// GetContentNews 获取内容页详情数据
func (n *News) GetContentNews(content map[string]string) {
	for url, title := range content {
		fmt.Println(url, title)
		if news, _, err := GetNews(url, title, timeOut, retryTime); err == nil {
			newsData := NewsData{}
			newsData.Url = url
			newsData.Title = news.Title
			newsData.Content = news.Content
			newsData.Time = news.TimeLocal
			n.data = append(n.data, &newsData)

			n.DataChanPush(&newsData)
		}
	}
}

// DataChanPush 推送data数据
func (n *News) DataChanPush(data *NewsData) {
	n.dataChan <- data
}

// DataChanPull 取出data数据
func (n *News) DataChanPull() NewsData {
	data := <-n.dataChan
	return *data
}

// GetSubdomains 获取subDomain
func GetSubdomains(url string, timeout int, retry int) (fun.StringSet, error) {
	if _, _, subDomains, err := GetLinkRes(url, timeout, retry); err == nil {
		return subDomains, nil
	} else {
		return nil, err
	}
}

// GetIndexUrl 获取首页url
func GetIndexUrl(url string) (string, string) {
	urlSlice := strings.Split(url, "/")
	scheme := urlSlice[0] + "//"
	indexUrl := scheme + urlSlice[2]
	return scheme, indexUrl
}
