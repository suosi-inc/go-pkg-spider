package spider

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/x-funs/go-fun"
)

// 新闻采集器结构体
type News struct {
	url         string            // 根链接
	depth       uint8             // 采集页面深度
	seen        map[string]bool   // 是否已采集
	isSub       bool              // 是否采集子域名
	LinkChan    chan *LinkData    // LinkData 通道共享
	ContentChan chan *NewsContent // NewsContent 通道共享
	ProcessFunc func(...any)      // 处理函数
	RetryTime   int               // 请求重试次数
	TimeOut     int               // 请求响应时间
	Wg          *sync.WaitGroup   // 同步等待组
	Req         *HttpReq          // 请求体
	Ctx         any               // 任务详情上下文
}

// 新闻内容结构体
type NewsContent struct {
	Url     string // 链接
	Title   string // 标题
	Time    string // 发布时间
	Content string // 正文纯文本
	Lang    string // 语种
}

// NewNews 初始化
func NewNews(url string, req *HttpReq, depth uint8, isSub bool, pf func(...any), ctx any) *News {
	return &News{
		url:         url,
		depth:       depth,
		seen:        map[string]bool{},
		isSub:       isSub,
		LinkChan:    make(chan *LinkData),
		ContentChan: make(chan *NewsContent),
		ProcessFunc: pf,
		RetryTime:   2,
		TimeOut:     20000,
		Wg:          &sync.WaitGroup{},
		Req:         req,
		Ctx:         ctx,
	}
}

// GetNews 开始采集
func (n *News) GetNews(linksHandleFunc func(*LinkData)) {
	// 初始化列表页和内容页切片
	var listSlice []string
	var listSliceTemp []string
	var subDomainSlice []string

	// 获取首页url和协议
	scheme, indexUrl := GetIndexUrl(n.url)

	// 首次添加当前页
	listSliceTemp = append(listSliceTemp, n.url)

	if n.isSub {
		// 先探测出首页url的所有子域名
		subDomains, _ := GetSubdomains(indexUrl, n.Req, n.TimeOut, n.RetryTime*100)

		for subDomain := range subDomains {
			subDomainSlice = append(subDomainSlice, subDomain)
			listSliceTemp = append(listSliceTemp, subDomain)
		}
	}

	// 深度优先循环遍历获取页面列表页和内容页
	for i := 0; i < int(n.depth); i++ {
		listS, _ := n.GetNewsLinkRes(linksHandleFunc, scheme, listSliceTemp, n.TimeOut, n.RetryTime)
		listSlice = append(listSlice, listS...)
		// contentSlice = append(contentSlice, contentS...)

		// 重置循环列表页
		if len(listS) == 0 {
			break
		}
		listSliceTemp = listS
	}
}

// GetNewsLinkRes 获取news页面链接分组, 仅返回列表页和内容页
func (n *News) GetNewsLinkRes(linksHandleFunc func(*LinkData), scheme string, urls []string, timeout int, retry int) ([]string, error) {
	listSlice := []string{}

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

			n.Wg.Add(1)
			go linksHandleFunc(linkData)

		} else {
			return nil, errors.New("GetNewsLinkRes Err")
		}
	}

	return listSlice, nil
}

// CrawlLinkRes 直接推送列表页内容页
func (n *News) CrawlLinkRes(l *LinkData) {
	defer n.Wg.Done()
	defer n.sleep()

	n.PushLinks(l)
}

// GetContentNews 解析内容页详情数据
func (n *News) CrawlContentNews(l *LinkData) {
	defer n.Wg.Done()
	defer n.sleep()

	for c, v := range l.LinkRes.Content {
		if !n.seen[c] {
			n.seen[c] = true
			cc := map[string]string{}
			cc[c] = v

			n.Wg.Add(1)
			go n.ReqContentNews(cc)
		}
	}
}

// ReqContentNews 获取内容页详情数据
func (n *News) ReqContentNews(content map[string]string) {
	defer n.Wg.Done()

	time.Sleep(time.Duration(fun.RandomInt(10, 100)) * time.Millisecond)

	for url, title := range content {
		if news, _, err := GetNews(url, title, n.TimeOut, n.RetryTime); err == nil {
			newsData := &NewsContent{}
			newsData.Url = url
			newsData.Title = news.Title
			newsData.Content = news.Content
			newsData.Time = news.TimeLocal
			newsData.Lang = news.Lang

			n.PushContentNews(newsData)
		}
	}
}

// PushLinks 推送links数据
func (n *News) PushLinks(data *LinkData) {
	n.LinkChan <- data
}

// PushContentNews 推送详情页数据
func (n *News) PushContentNews(data *NewsContent) {
	n.ContentChan <- data
}

// Wait wg阻塞等待退出
func (n *News) Wait() {
	n.Wg.Wait()
}

// Close 关闭Chan
func (n *News) Close() {
	close(n.LinkChan)
	close(n.ContentChan)
}

// process 处理chan data函数
func (n *News) process(processFunc func(...any)) {
	for {
		select {
		case data, ok := <-n.LinkChan:
			if !ok {
				return
			}
			processFunc(data, n.Ctx)
		case data, ok := <-n.ContentChan:
			if !ok {
				return
			}
			processFunc(data, n.Ctx)
		}
	}
}

// GetLinkRes 回调获取LinkRes数据
func (n *News) GetLinkRes() {
	n.GetNews(n.CrawlLinkRes)

	go n.process(n.ProcessFunc)

	n.Wait()
	defer n.Close()
}

// GetContentNews 回调获取内容页数据
func (n *News) GetContentNews() {
	n.GetNews(n.CrawlContentNews)

	go n.process(n.ProcessFunc)

	n.Wait()
	defer n.Close()
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

// sleep depth只有一层时，需要等待几秒，避免wg done后直接退出，导致select来不及取出数据
func (n *News) sleep() {
	if n.depth == 1 {
		time.Sleep(2 * time.Second)
	}
}
