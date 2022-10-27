package spider

import (
	"strings"
	"sync"
	"time"

	"github.com/x-funs/go-fun"
)

// 新闻采集器结构体
type NewsSpider struct {
	Url         string            // 根链接
	Depth       uint8             // 采集页面深度
	seen        map[string]bool   // 是否已采集
	IsSub       bool              // 是否采集子域名
	linkChan    chan *NewsData    // NewsData 通道共享
	contentChan chan *NewsContent // NewsContent 通道共享
	ProcessFunc func(...any)      // 处理函数
	RetryTime   int               // 请求重试次数
	TimeOut     int               // 请求响应时间
	wg          *sync.WaitGroup   // 同步等待组
	Req         *HttpReq          // 请求体
	Ctx         any               // 任务详情上下文，传入ProcessFunc函数中
}

// 新闻内容结构体
type NewsContent struct {
	Url     string // 链接
	Title   string // 标题
	Time    string // 发布时间
	Content string // 正文纯文本
	Lang    string // 语种
}

// 新闻LinkData总数据
type NewsData struct {
	*LinkData
	Depth   uint8  // 采集深度溯源
	ListUrl string // 列表页溯源
	Error   error
}

// 自定义配置函数
type Option func(*NewsSpider)

// 原型链接口
type Prototype interface {
	Clone() Prototype
}

// NewNewsSpider 初始化
func NewNewsSpider(url string, depth uint8, pf func(...any), ctx any, options ...Option) *NewsSpider {
	n := &NewsSpider{
		Url:         url,
		Depth:       depth,
		seen:        map[string]bool{},
		IsSub:       false,
		linkChan:    make(chan *NewsData),
		contentChan: make(chan *NewsContent),
		ProcessFunc: pf,
		RetryTime:   2,
		TimeOut:     20000,
		wg:          &sync.WaitGroup{},
		Req:         nil,
		Ctx:         ctx,
	}

	// 函数式选项模式
	for _, option := range options {
		option(n)
	}

	return n
}

func WithRetryTime(retryTime int) Option {
	return func(n *NewsSpider) {
		n.RetryTime = retryTime
	}
}

func WithTimeOut(timeout int) Option {
	return func(n *NewsSpider) {
		n.TimeOut = timeout
	}
}

func WithReq(req *HttpReq) Option {
	return func(n *NewsSpider) {
		n.Req = req
	}
}

func WithIsSub(isSub bool) Option {
	return func(n *NewsSpider) {
		n.IsSub = isSub
	}
}

// 原型链结构体拷贝
func (n *NewsSpider) Clone() Prototype {
	nc := *n

	// 拷贝时需重置chan和wg等字段
	nc.seen = map[string]bool{}
	nc.linkChan = make(chan *NewsData)
	nc.contentChan = make(chan *NewsContent)
	nc.wg = &sync.WaitGroup{}

	return &nc
}

// GetNews 开始采集
func (n *NewsSpider) GetNews(linksHandleFunc func(*NewsData)) {
	// 初始化列表页和内容页切片
	var (
		listSlice      []string
		listSliceTemp  []string
		subDomainSlice []string
	)

	// 获取首页url和协议
	scheme, indexUrl := GetIndexUrl(n.Url)

	// 首次添加当前页
	listSliceTemp = append(listSliceTemp, n.Url)

	if n.IsSub {
		// 先探测出首页url的所有子域名
		subDomains, _ := GetSubdomains(indexUrl, n.Req, n.TimeOut, n.RetryTime*100)

		for subDomain := range subDomains {
			subDomainSlice = append(subDomainSlice, subDomain)
			listSliceTemp = append(listSliceTemp, subDomain)
		}
	}

	// 深度优先循环遍历获取页面列表页和内容页
	for i := 0; i < int(n.Depth); i++ {
		listS, _ := n.GetNewsLinkRes(linksHandleFunc, scheme, listSliceTemp, uint8(i+1), n.TimeOut, n.RetryTime)
		listSlice = append(listSlice, listS...)

		// 重置循环列表页
		if len(listS) == 0 {
			break
		}
		listSliceTemp = listS
	}
}

// GetNewsLinkRes 获取news页面链接分组, 仅返回列表页和内容页
func (n *NewsSpider) GetNewsLinkRes(linksHandleFunc func(*NewsData), scheme string, urls []string, depth uint8, timeout int, retry int) ([]string, error) {
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

			newsData := &NewsData{linkData, depth, url, nil}

			n.wg.Add(1)
			go linksHandleFunc(newsData)

		} else {
			// 报错空的LinkData也需要push
			newsData := &NewsData{nil, depth, url, err}

			n.wg.Add(1)
			go linksHandleFunc(newsData)

			// return nil, errors.New("GetNewsLinkRes Err")
		}
	}

	return listSlice, nil
}

// CrawlLinkRes 直接推送列表页内容页
func (n *NewsSpider) CrawlLinkRes(l *NewsData) {
	defer n.wg.Done()
	defer n.sleep()

	n.PushLinks(l)
}

// GetContentNews 解析内容页详情数据
func (n *NewsSpider) CrawlContentNews(l *NewsData) {
	defer n.wg.Done()
	defer n.sleep()

	if l.Error == nil {
		for c, v := range l.LinkRes.Content {
			if !n.seen[c] {
				n.seen[c] = true
				cc := map[string]string{}
				cc[c] = v

				n.wg.Add(1)
				go n.ReqContentNews(cc)
			}
		}
	}

}

// ReqContentNews 获取内容页详情数据
func (n *NewsSpider) ReqContentNews(content map[string]string) {
	defer n.wg.Done()

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
func (n *NewsSpider) PushLinks(data *NewsData) {
	n.linkChan <- data
}

// PushContentNews 推送详情页数据
func (n *NewsSpider) PushContentNews(data *NewsContent) {
	n.contentChan <- data
}

// Wait wg阻塞等待退出
func (n *NewsSpider) Wait() {
	n.wg.Wait()
}

// Close 关闭Chan
func (n *NewsSpider) Close() {
	close(n.linkChan)
	close(n.contentChan)
}

// process 处理chan data函数
func (n *NewsSpider) process(processFunc func(...any)) {
	for {
		select {
		case data, ok := <-n.linkChan:
			if !ok {
				return
			}
			processFunc(data, n.Ctx)
		case data, ok := <-n.contentChan:
			if !ok {
				return
			}
			processFunc(data, n.Ctx)
		}
	}
}

// GetLinkRes 回调获取LinkRes数据
func (n *NewsSpider) GetLinkRes() {
	n.GetNews(n.CrawlLinkRes)

	go n.process(n.ProcessFunc)

	n.Wait()
	defer n.Close()
}

// GetContentNews 回调获取内容页数据
func (n *NewsSpider) GetContentNews() {
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
func (n *NewsSpider) sleep() {
	if n.Depth == 1 {
		time.Sleep(2 * time.Second)
	}
}
