<h1 align="center">
  go-pkg-spider
</h1>

<p align="center">go-pkg-spider 期望实现一个相对智能、无需规则维护的通用新闻网站采集工具库。</p>

## 快速预览

前往 [go-pkg-spider-gui](https://github.com/suosi-inc/go-pkg-spider-gui) 下载支持 Windows、MacOS Gui 客户端，进行快速功能预览。

## 功能介绍

### Http 客户端

Http 客户端在 go-fun 中的 `fun.HttpGet` 相关函数进行了一些扩展，增加了以下功能：

* 自动探测字符集和转换字符集
* 响应文本类型限制

最终转换为 UTF-8

### 网页语种自动探测

集成了 n-gram model [lingua-go](https://github.com/pemistahl/lingua-go)，但是移除了很多语种和语料（因为完整包很大）

当前支持以下主流语种：中文、英语、日语、韩语、俄语、阿拉伯语、印地语、德语、法语、西班牙语、葡萄牙语、意大利语、泰语、越南语、缅甸语

```

```



### 域名自动探测

### 网页链接分类提取


### 网页新闻三要素提取

新闻最重要的三要素包含：标题、发布时间、正文。其中发布时间对精准度要求高，标题和正文更追求完整性。

体验下来，业内最强大的是： [diffbot](https://www.diffbot.com/) 公司，猜测它可能是基于网页视觉+深度学习来实现。

有不少新闻正文抽取的开源的方案，大都是基于规则或统计方法实现。如：

* Python: [GeneralNewsExtractor](https://github.com/GeneralNewsExtractor/GeneralNewsExtractor)
* Java: [WebCollector/ContentExtractor](https://github.com/CrawlScript/WebCollector)

更古老的还有：[python-goose](https://github.com/grangier/python-goose), [newspaper](https://github.com/codelucas/newspaper)，甚至 Readability、Html2Article 等等。

其中：`WebCollector/ContentExtractor` 是 [基于标签路径特征融合新闻内容抽取的 CEPF 算法](http://www.jos.org.cn/jos/article/abstract/4868) 的 Java 实现版本。

本项目新闻要素提取部分实现了 CEPF 算法的 Golang 版本，并在此基础上做了大量优化，内置了一些通用规则，更精细的控制了标题和发布时间的提取，支持多语种新闻网站的要素提取。

新闻要素提取结果结构体：

```
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

在 go-pkg-spider GUI 功能演示：


