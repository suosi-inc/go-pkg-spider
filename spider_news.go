package spider

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/x-funs/go-fun"
)

const (
	timeOut     = 20000
	retryTime   = 2
	retryAlways = 99
)

type News struct {
	url      string          // 根链接
	depth    uint8           // 采集页面深度
	seen     map[string]bool // 是否已采集
	isSub    bool            // 是否采集子域名
	data     []*NewsData     // newsData 切片
	DataChan chan *NewsData  // newsData 通道共享
	Wg       *sync.WaitGroup // 同步等待组
	Req      *HttpReq        // 请求体
}

type NewsData struct {
	Url     string // 链接
	Title   string // 标题
	Time    string // 发布时间
	Content string // 正文纯文本
	Lang    string // 语种
}

// NewNews 初始化
func NewNews(url string, req *HttpReq, depth uint8, isSub bool) *News {
	return &News{
		url:      url,
		depth:    depth,
		seen:     map[string]bool{},
		isSub:    isSub,
		data:     []*NewsData{},
		DataChan: make(chan *NewsData),
		Wg:       &sync.WaitGroup{},
		Req:      req,
	}
}

// GetNews 开始获取news
func (n *News) GetNews(contentHandleFunc func(content map[string]string)) {
	// 初始化列表页和内容页切片
	var listSlice []string
	var listSliceTemp []string
	var contentSlice []map[string]string
	var subDomainSlice []string

	// 获取首页url和协议
	scheme, indexUrl := GetIndexUrl(n.url)

	if n.isSub {
		// 先探测出首页url的所有子域名
		subDomains, err := GetSubdomains(indexUrl, n.Req, timeOut, retryAlways)
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

		if linkData, err := GetLinkDataWithReq(url, true, n.Req, timeout, retry); err == nil {
			for l := range linkData.LinkRes.List {
				if !n.seen[l] {
					n.seen[l] = true
					listSlice = append(listSlice, l)
				}
			}

			for c, v := range linkData.LinkRes.Content {
				if !n.seen[c] {
					n.seen[c] = true
					cc := map[string]string{}
					cc[c] = v
					contentSlice = append(contentSlice, cc)

					n.Wg.Add(1)
					go contentHandleFunc(cc)
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
	defer n.Wg.Done()

	time.Sleep(time.Duration(fun.RandomInt(10, 100)) * time.Millisecond)

	for url, title := range content {
		if news, _, err := GetNews(url, title, timeOut, retryTime); err == nil {
			newsData := NewsData{}
			newsData.Url = url
			newsData.Title = news.Title
			newsData.Content = news.Content
			newsData.Time = news.TimeLocal
			newsData.Lang = news.Lang
			n.data = append(n.data, &newsData)

			n.DataChanPush(&newsData)
		} else {
			fmt.Println("getContentNews err:" + err.Error())
		}
	}
	return
}

// PrintContentNews 打印内容页
func (n *News) PrintContentNews(content map[string]string) {
	defer n.Wg.Done()

	time.Sleep(1 * time.Second)
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
func GetSubdomains(url string, req *HttpReq, timeout int, retry int) (map[string]bool, error) {
	if linkData, err := GetLinkDataWithReq(url, true, req, timeout, retry); err == nil {
		return linkData.SubDomains, nil
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
