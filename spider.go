package spider

import (
	"bytes"
	"errors"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/extract"
	"github.com/x-funs/go-fun"
)

const (
	RegexHostnameIp = `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`

	RegexMetaRefresh = `(?i)url=(.+)`
)

var (
	DefaultDocRemoveTags = "script,noscript,style,iframe,br,link,svg"

	RegexHostnameIpPattern = regexp.MustCompile(RegexHostnameIp)

	regexMetaRefreshPattern = regexp.MustCompile(RegexMetaRefresh)
)

type LinkData struct {
	LinkRes    *extract.LinkRes
	Filters    map[string]string
	SubDomains map[string]bool
}

// GetLinkData 获取页面链接分组
func GetLinkData(urlStr string, strictDomain bool, timeout int, retry int) (*LinkData, error) {
	if retry <= 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		linkData, err := GetLinkDataDo(urlStr, strictDomain, nil, timeout)
		if err == nil {
			return linkData, err
		}
	}

	return nil, errors.New("ErrorLinkRes")
}

// GetLinkDataWithRule 获取页面链接分组
func GetLinkDataWithRule(urlStr string, strictDomain bool, rules extract.LinkTypeRule, timeout int, retry int) (*LinkData, error) {
	if retry <= 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		linkData, err := GetLinkDataDo(urlStr, strictDomain, rules, timeout)
		if err == nil {
			return linkData, err
		}
	}

	return nil, errors.New("ErrorLinkRes")
}

// GetLinkDataDo 获取页面链接分组
func GetLinkDataDo(urlStr string, strictDomain bool, rules extract.LinkTypeRule, timeout int) (*LinkData, error) {
	if timeout == 0 {
		timeout = 10000
	}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: HttpDefaultMaxContentLength,
			MaxRedirect:      3,
		},
		ForceTextContentType: true,
	}

	resp, err := HttpGetResp(urlStr, req, timeout)
	if resp != nil && err == nil && resp.Success {
		// 解析 HTML
		doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		if docErr == nil {
			linkData := &LinkData{}

			doc.Find(DefaultDocRemoveTags).Remove()

			// 语言
			langRes := Lang(doc, resp.Charset.Charset, true)

			// 站内链接
			linkTitles, filters := extract.WebLinkTitles(doc, resp.RequestURL, strictDomain)

			// 链接分类
			linkRes, subDomains := extract.LinkTypes(linkTitles, langRes.Lang, rules)

			linkData.LinkRes = linkRes
			linkData.Filters = filters
			linkData.SubDomains = subDomains

			return linkData, nil
		} else {
			return nil, errors.New("ErrorDocParse")
		}
	}

	return nil, errors.New("ErrorRequest")
}

// GetNews 获取正文
func GetNews(urlStr string, title string, timeout int, retry int) (*extract.News, *HttpResp, error) {
	if retry <= 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		news, resp, err := GetNewsDo(urlStr, title, timeout)
		if err == nil {
			return news, resp, nil
		}
	}

	return nil, nil, errors.New("ErrorRequest")
}

// GetNewsDo 获取正文
func GetNewsDo(urlStr string, title string, timeout int) (*extract.News, *HttpResp, error) {
	return getNewsDoTop(urlStr, title, timeout, true)
}

// getNewsDoTop 获取正文
func getNewsDoTop(urlStr string, title string, timeout int, top bool) (*extract.News, *HttpResp, error) {
	if timeout == 0 {
		timeout = HttpDefaultTimeOut
	}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: HttpDefaultMaxContentLength,
			MaxRedirect:      2,
		},
		ForceTextContentType: true,
	}

	resp, err := HttpGetResp(urlStr, req, timeout)

	if resp != nil && err == nil && resp.Success {
		doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		if docErr == nil {
			contentDoc := goquery.CloneDocument(doc)
			doc.Find(DefaultDocRemoveTags).Remove()

			// 具有 HTML 跳转属性, 如果为本域名下, 则跳转一次
			if top {
				if refresh, exists := doc.Find("meta[http-equiv='refresh' i]").Attr("content"); exists {
					refreshMatch := regexMetaRefreshPattern.FindStringSubmatch(refresh)
					if len(refreshMatch) > 1 {
						requestHostname := resp.RequestURL.Hostname()
						requestTopDomain := extract.DomainTop(requestHostname)
						refreshUrl := strings.TrimSpace(refreshMatch[1])
						if r, err := fun.UrlParse(refreshUrl); err == nil {
							refreshHostname := r.Hostname()
							refreshTopDomain := extract.DomainTop(refreshHostname)
							if refreshTopDomain != "" && refreshTopDomain == requestTopDomain {
								return getNewsDoTop(refreshUrl, title, timeout, false)
							}
						}
					}
				}
			}

			// 语言
			langRes := Lang(doc, resp.Charset.Charset, false)

			// 正文抽取
			content := extract.NewContent(contentDoc, langRes.Lang, title, urlStr)
			news := content.ExtractNews()

			return news, resp, nil
		} else {
			return nil, resp, errors.New("ErrorDocParse")
		}
	}

	return nil, nil, errors.New("ErrorRequest")
}
