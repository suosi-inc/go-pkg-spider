package spider

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"testing"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/suosi-inc/go-pkg-spider/extract"
	"github.com/x-funs/go-fun"
)

func BenchmarkHtmlParse(b *testing.B) {

	resp, _ := fun.HttpGetResp("https://www.163.com", nil, 30000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find(DefaultDocRemoveTags).Remove()
	}
}

func TestGoquery(t *testing.T) {
	body, _ := HttpGet("https://jp.news.cn/index.htm")
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	// lang, exist := doc.Find("html").Attr("id")

	doc.Find("script,noscript,style,iframe,br,link,svg,textarea").Remove()
	text := doc.Find("body").Text()
	text = fun.RemoveSign(text)

	fmt.Println(text)
}

func TestRegex(t *testing.T) {
	str := ",.!，，D_NAME。！；‘’”“《》**dfs#%^&()-+我1431221     中国123漢字かどうかのjavaを<決定>$¥"
	r := regexp.MustCompile(`[\p{Hiragana}|\p{Katakana}]`)
	s := r.FindAllString(str, -1)
	t.Log(str)
	t.Log(s)
}

func TestUrlParse(t *testing.T) {
	var urlStrs = []string{
		"https://www.163.com",
		"https://www.163.com/",
		"https://www.163.com/a",
		"https://www.163.com/aa.html",
		"https://www.163.com/a/b",
		"https://www.163.com/a/bb.html",
		"https://www.163.com/a/b/",
		"https://www.163.com/a/b/c",
		"https://www.163.com/a/b/cc.html",
	}

	for _, urlStr := range urlStrs {
		u, _ := url.Parse(urlStr)
		link := "javascript:;"
		absolute, err := u.Parse(link)
		t.Log(err)

		_, err = url.Parse(absolute.String())
		if err != nil {
			t.Log(err)
		}

		t.Log(urlStr + "	+ " + link + " => " + absolute.String())
	}

}

func TestCount(t *testing.T) {
	fmt.Println(regexLangHtmlPattern.MatchString("zh"))
	fmt.Println(regexLangHtmlPattern.MatchString("en"))
	fmt.Println(regexLangHtmlPattern.MatchString("zh-cn"))
	fmt.Println(regexLangHtmlPattern.MatchString("utf-8"))

	fmt.Println(utf8.RuneCountInString("https://khmers.cn/2022/05/23/%e6%b4%aa%e6%a3%ae%e6%80%bb%e7%90%86%ef%bc%9a%e6%9f%ac%e5%9f%94%e5%af%a8%e7%b4%af%e8%ae%a1%e8%8e%b7%e5%be%97%e8%b6%85%e8%bf%875200%e4%b8%87%e5%89%82%e6%96%b0%e5%86%a0%e7%96%ab%e8%8b%97%ef%bc%8c/"))
}

func TestGetLinkRes(t *testing.T) {
	var urlStrs = []string{
		// "https://www.1905.com",
		"https://www.people.com.cn",
		// "https://www.36kr.com",
		// "https://www.163.com",
		// "https://news.163.com/",
		// "http://jyj.suqian.gov.cn",
		// "https://www.huxiu.com/",
		// "http://www.news.cn/politicspro/",
		// "http://www.cankaoxiaoxi.com",
		// "http://www.bbc.com",
		// "https://www.ft.com",
		// "https://www.reuters.com/",
		// "https://nypost.com/",
		// "http://www.mengcheng.gov.cn/",
		// "https://www.chunichi.co.jp",
		// "https://www.donga.com/",
		// "https://people.com/",
		// "https://czql.gov.cn/",
		// "https://qiye.163.com/",
		// "https://www.washingtontimes.com/",
		// "https://www.gamersky.com/",
		// "https://www.cdns.com.tw/",
		// "http://www.163.com/",
	}

	for _, urlStr := range urlStrs {

		if linkRes, filters, err := GetLinkRes(urlStr, 10000, 1); err == nil {
			fmt.Println("content:", len(linkRes.Content))
			fmt.Println("list:", len(linkRes.List))
			fmt.Println("unknown:", len(linkRes.Unknown))
			fmt.Println("none:", len(linkRes.None))

			i := 0
			for a, title := range filters {
				i = i + 1
				fmt.Println(i, "filter:"+a+"\t=>\t"+title)
			}
			i = 0
			for a, title := range linkRes.Content {
				i = i + 1
				fmt.Println(i, "content:"+a+"\t=>\t"+title)
			}
			i = 0
			for a, title := range linkRes.Unknown {
				i = i + 1
				fmt.Println(i, "unknown:"+a+"\t=>\t"+title)
			}
			i = 0
			for a, title := range linkRes.List {
				i = i + 1
				fmt.Println(i, "list:"+a+"\t=>\t"+title)
			}
			i = 0
			for a, title := range linkRes.None {
				i = i + 1
				fmt.Println(i, "none:"+a+"\t=>\t"+title)
			}
		}
	}
}

func TestGetNews(t *testing.T) {

	var urlStrs = []string{
		// "http://www.cankaoxiaoxi.com/finance/20220831/2489264.shtml",
		// "https://www.163.com/news/article/HG3DE7AQ000189FH.html",
		// "http://suosi.com.cn/",
		// "http://www.cankaoxiaoxi.com/world/20220831/2489267.shtml",
		// "http://www.cankaoxiaoxi.com/photo/20220901/2489404.shtml",
		// "http://column.cankaoxiaoxi.com/2022/0831/2489330.shtml",
		// "http://www.gov.cn/xinwen/2022-08/31/content_5707661.htm",
		// "http://suosi.com.cn/2019/14.shtml",
		// "https://www.wangan.com/p/7fy78317feb66b37",
		// "https://www.wangan.com/news/7fy78y38c7207bf0",
		// "http://env.people.com.cn/n1/2022/0901/c1010-32516651.html",
		// "http://www.changzhou.gov.cn/ns_news/827166202029392",
		// "https://www.163.com/money/article/HG4TRBL1002580S6.html?clickfrom=w_yw_money",
		// "https://mp.weixin.qq.com/s?__biz=MzUxODkxNTYxMA==&mid=2247484842&idx=1&sn=d9822ee4662523609aee7441066c2a96&chksm=f980d6dfcef75fc93cb1e7942cb16ec82a7fb7ec3c2d857c307766daff667bd63ab1b4941abd&exportkey=AXWfguuAyJjlOJgCHf10io8%3D&acctmode=0&pass_ticket=8eXqj",
		// "https://www.bbc.com/news/world-asia-62744522",
		// "https://www.sohu.com/a/581634395_121284943",
		// "https://edition.cnn.com/2022/01/30/europe/lithuania-took-on-china-intl-cmd/index.html",
		// "https://www.36kr.com/p/1897541916043649",
		// "https://www.huxiu.com/article/651531.html",
		// "http://www.news.cn/politics/2022-09/02/c_1128969463.htm",
		// "https://www.ccdi.gov.cn/yaowenn/202209/t20220901_215343.html",
		// "https://new.qq.com/omn/20200701/20200701A04H7500",
		// "http://v.china.com.cn/2022-09/06/content_78407150.html",
		// "http://www.chinagwy.org.cn/content-cat-10/143162.html",
		// "https://news.52pk.com/xwlm/201912/7366710.shtml",
		// "https://www.business-standard.com/article/finance/govt-rbi-propose-action-plan-for-facilitating-special-rupee-accounts-122090701260_1.html",
		// "https://www.squirepattonboggs.com/en/news/2022/09/squire-patton-boggs-advises-new-wave-group-ab-on-uk-acquisition",
		// "https://www.thebulletin.be/number-road-deaths-belgium-rises-sharply",
		"https://mp.weixin.qq.com/s/iXRZdy9EqK6dDHz9M2FG9A",
	}

	for _, urlStr := range urlStrs {
		if news, resp, err := GetNews(urlStr, "", 10000, 1); err == nil {
			t.Log(resp.Charset)
			t.Log(news.Spend)
			t.Log(news.Title)
			t.Log(news.TitlePos)
			t.Log(news.TimeLocal)
			t.Log(news.Time)
			t.Log(news.TimePos)
			t.Log(news.Content)

			if news.ContentNode != nil {
				// 内容 html 节点
				node := goquery.NewDocumentFromNode(news.ContentNode)
				contentHtml, _ := node.Html()
				t.Log(fun.NormaliseLine(contentHtml))

				// 内容 html 节点清理, 仅保留 p img 标签
				p := bluemonday.NewPolicy()
				p.AllowElements("p")
				p.AllowImages()
				html := p.Sanitize(contentHtml)
				t.Log(fun.NormaliseLine(html))
			}
		}
	}
}

func TestDemo(t *testing.T) {
	a := "2022-05-26 17:00:57 UTC"
	findString := regexp.MustCompile(extract.RegexPublishDate).FindStringSubmatch(a)
	t.Log(findString)
	t.Log(fun.Date(fun.StrToTime("2022-04-10T18:24:00")))
}
