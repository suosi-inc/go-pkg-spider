```
                            __                          _     __         
   ____ _____        ____  / /______ _      _________  (_)___/ /__  _____
  / __ `/ __ \______/ __ \/ //_/ __ `/_____/ ___/ __ \/ / __  / _ \/ ___/
 / /_/ / /_/ /_____/ /_/ / ,< / /_/ /_____(__  ) /_/ / / /_/ /  __/ /    
 \__, /\____/     / .___/_/|_|\__, /     /____/ .___/_/\__,_/\___/_/     
/____/           /_/         /____/          /_/                         

```

一个 Golang 实现的相对智能、无需规则维护的通用新闻网站数据提取工具库。含域名探测、网页编码语种识别、网页链接分类提取、网页新闻要素抽取以及新闻正文抽取等组件。

# 预览

前往 [go-pkg-spider-gui Releases](https://github.com/suosi-inc/go-pkg-spider-gui/releases) 下载支持 Windows、MacOS GUI 客户端，进行体验。

<p align="center" markdown="1" style="max-width: 100%">
  <img src="https://raw.githubusercontent.com/suosi-inc/go-pkg-spider-gui/main/images/zh/win10.png" width="800" style="max-width: 100%" />
</p>

# 使用

```shell
go get -u github.com/suosi-inc/go-pkg-spider
```

# 介绍

## Http 客户端

Http 客户端在 go-fun 中的 `fun.HttpGet` 相关函数进行了一些扩展，增加了以下功能：

* 自动识别字符集和转换字符集，统一转换为 UTF-8
* 响应文本类型限制

- **<big>`HttpGet(urlStr string, args ...any) ([]byte, error)`</big>** Http Get 请求
- **<big>`HttpGetResp(urlStr string, r *HttpReq, timeout int) (*HttpResp, error)`</big>** Http Get 请求, 返回 HttpResp

## 网页语种自动识别

当前支持以下主流语种：**中文、英语、日语、韩语、俄语、阿拉伯语、印地语、德语、法语、西班牙语、葡萄牙语、意大利语、泰语、越南语、缅甸语**。

语种识别通过 HTML 、文本特征、字符集统计规则优先识别中文、英语、日语、韩语。

同时辅助集成了 [lingua-go](https://github.com/pemistahl/lingua-go) n-gram model 语言识别模型，fork 并移除了很多语种和语料（因为完整包很大）

- **<big>`LangText(text string) (string, string)`</big>** 识别纯文本语种
- **<big>`Lang(doc *goquery.Document, charset string, listMode bool) LangRes `</big>** 识别 HTML 语种

### 示例

识别纯文本语种：

```go
// 识别纯文本语种
lang, langPos := spider.LangText(text)
```

识别 HTML 语种：

```go
// Http 请求获取响应
resp, err := spider.HttpGetResp(urlStr, req, timeout)

// 转换 goquery.*Document
doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))

// 根据字符集、页面类型返回
langRes := spider.Lang(doc, resp.Charset.Charset, false)
```

## 域名自动探测

- **<big>`DetectDomain(domain string, timeout int, retry int) (*DomainRes, error)`</big>** 探测主域名基本信息
- **<big>`func DetectSubDomain(domain string, timeout int, retry int) (*DomainRes, error)`</big>** 探测子域名基本信息

根据网站域名，尽可能的探测一些基本信息，基本信息包括：

```go
type DomainRes struct {
	// 域名
	Domain       string
	// 主页域名
	HomeDomain   string
	// 协议
	Scheme       string
	// 字符集
	Charset      CharsetRes
	// 语种
	Lang         LangRes
	// 国家
	Country      string
	// 省份
	Province     string
	// 分类
	Category     string
	// 标题
	Title        string
	// 描述
	Description  string
	// ICP
	Icp          string
	// 状态
	State        bool
	// 状态码
	StatusCode   int
	// 内容页链接数量
	ContentCount int
	// 列表页链接数量
	ListCount    int
	// 子域名列表
	SubDomains   map[string]bool
}
```

## 网页链接分类提取

根据页面内容，自动分析识别并提取页面上的内容页、列表页以及其他链接，支持传入自定义规则干扰最终结果

分类依据通过链接标题、URL特征、以及统计归纳的方式

- **<big>`GetLinkData(urlStr string, strictDomain bool, timeout int, retry int) (*LinkData, error)`</big>** 获取页面链接分类数据

### 链接分类提取结果定义

```go
type LinkData struct {
	LinkRes    *extract.LinkRes
	// 过滤
	Filters    map[string]string
	// 子域名
	SubDomains map[string]bool
}

type LinkRes struct {
	// 内容页
	Content map[string]string
	// 列表页
	List map[string]string
	// 未知链接
	Unknown map[string]string
	// 过滤链接
	None map[string]string
}
```

## 网页新闻提取

新闻最重要的三要素：标题、发布时间、正文。其中发布时间对精准度要求高，标题和正文更追求完整性。

体验下来，业内最强大的是： [diffbot](https://www.diffbot.com/) 公司，猜测它可能是基于网页视觉+深度学习来实现。

有不少新闻正文提取或新闻正文抽取的开源的方案，大都是基于规则或统计方法实现。如：

* Python: [GeneralNewsExtractor](https://github.com/GeneralNewsExtractor/GeneralNewsExtractor)
* Java: [WebCollector/ContentExtractor](https://github.com/CrawlScript/WebCollector)

更古老的还有：[python-goose](https://github.com/grangier/python-goose), [newspaper](https://github.com/codelucas/newspaper)，甚至 Readability、Html2Article 等等。

其中：`WebCollector/ContentExtractor` 是 [基于标签路径特征融合新闻内容抽取的 CEPF 算法](http://www.jos.org.cn/jos/article/abstract/4868) 的 Java 实现版本。

go-pkg-spider 实现了 CEPF 算法的 Golang 版本，在此基础上做了大量优化，内置了一些通用规则，更精细的控制了标题和发布时间的提取与转换，并支持多语种新闻网站的要素提取。


### 新闻要素提取结果定义

```go
type News struct {
	// 标题
	Title string
	// 标题提取依据
	TitlePos string
	// 发布时间
	TimeLocal string
	// 原始时间
	Time string
	// 发布时间时间提取依据
	TimePos string
	// 正文纯文本
	Content string
	// 正文 Node 节点
	ContentNode *html.Node
	// 提取用时（毫秒）
	Spend int64
	// 语种
	Lang string
}
```

可根据 `ContentNode *html.Node` 来重新定义需要清洗保留的标签。

### 效果

<p align="center" markdown="1" style="max-width: 100%">
  <img src="https://raw.githubusercontent.com/suosi-inc/go-pkg-spider-gui/main/images/zh/content.png" width="800" style="max-width: 100%" />
</p>

### 示例

```go
// Http 请求获取响应
resp, err := spider.HttpGetResp(urlStr, req, timeout)

// 转换 goquery.*Document
doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))

// 基本清理
doc.Find(spider.DefaultDocRemoveTags).Remove()

// 语种
langRes := Lang(doc, resp.Charset.Charset, false)

// 新闻提取
content := extract.NewContent(contentDoc, langRes.Lang, listTitle, urlStr)

// 新闻提取结果
news := content.ExtractNews()
```

可以通过下面的已经封装好的方法完成以上步骤：

- **<big>`GetNews(urlStr string, title string, timeout int, retry int) (*extract.News, *HttpResp, error)`</big>** 获取链接新闻数据

# 免责声明

本项目是一个数据提取工具库，不是爬虫框架或采集软件，只限于技术交流，源码中请求目标网站的相关代码仅为功能测试需要。

请在符合法律法规和相关规定的情况下使用本项目，禁止使用本项目进行任何非法、侵权或者违反公序良俗的行为。

使用本项目造成的直接或间接的风险由用户自行承担。
