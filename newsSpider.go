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
	url        string          // 根链接
	subDomains []string        // 子域名
	depth      uint8           // 采集页面深度
	seen       map[string]bool // 是否已采集
	isSub      bool            // 是否采集子域名
	data       []*NewsData     // newsData切片
	DataChan   chan *NewsData  // newsData通道共享
}

type NewsData struct {
	Url string // 链接

	Title string // 标题

	Time string // 发布时间

	Content string // 正文纯文本
}

// NewNews 初始化newsSpider
func NewNews(domain string, depth uint8, isSub bool) *News {
	return &News{
		url:      domain,
		depth:    depth,
		seen:     map[string]bool{},
		isSub:    isSub,
		data:     []*NewsData{},
		DataChan: make(chan *NewsData),
	}
}

// GetNews 开始获取news
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

			// linkRes.Content = map[string]string{
			// 	"http://yoga.com/1":  "666",
			// 	"http://yoga.com/2":  "666",
			// 	"http://yoga.com/3":  "666",
			// 	"http://yoga.com/4":  "666",
			// 	"http://yoga.com/5":  "666",
			// 	"http://yoga.com/6":  "666",
			// 	"http://yoga.com/7":  "666",
			// 	"http://yoga.com/8":  "666",
			// 	"http://yoga.com/9":  "666",
			// 	"http://yoga.com/10": "666",
			// 	"http://yoga.com/11": "666",
			// 	"http://yoga.com/12": "666",
			// 	"http://yoga.com/13": "666",
			// 	"http://yoga.com/14": "666",
			// 	"http://yoga.com/15": "666",
			// 	"http://yoga.com/16": "666",
			// 	"http://yoga.com/17": "666",
			// 	"http://yoga.com/18": "666",
			// 	"http://yoga.com/19": "666",
			// 	"http://yoga.com/20": "666",
			// 	"http://yoga.com/21": "666",
			// 	"http://yoga.com/22": "666",
			// 	"http://yoga.com/23": "666",
			// 	"http://yoga.com/24": "666",
			// 	"http://yoga.com/25": "666",
			// 	"http://yoga.com/26": "666",
			// 	"http://yoga.com/27": "666",
			// 	"http://yoga.com/28": "666",
			// 	"http://yoga.com/29": "666",
			// 	"http://yoga.com/30": "666",
			// }

			for c, v := range linkRes.Content {
				fmt.Println("handle news:", c)
				if !n.seen[c] {
					n.seen[c] = true
					cc := map[string]string{}
					cc[c] = v
					contentSlice = append(contentSlice, cc)
					// 内容页处理
					// go contentHandleFunc(cc)
					contentHandleFunc(cc)
				} else {
					fmt.Println("same news")
				}
			}

		} else {
			fmt.Println("GetNewsLinkRes", err)
		}
	}

	return listSlice, contentSlice, nil
}

// GetData　通过内存来通信
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
		} else {
			fmt.Println("getContentNews err:" + err.Error())
		}
	}
}

// PrintContentNews 打印内容页
func (n *News) PrintContentNews(content map[string]string) {
	for url, title := range content {
		fmt.Println("print news:", url, title)
	}
}

// DataChanPush 推送data数据
func (n *News) DataChanPush(data *NewsData) {
	n.DataChan <- data
}

// DataChanPull 取出data数据
func (n *News) DataChanPull() NewsData {
	data := <-n.DataChan
	return *data
}

// Close 关闭dataChan
func (n *News) Close() {
	close(n.DataChan)
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
