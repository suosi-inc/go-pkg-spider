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

// GetLinkRes 获取页面链接分组
func GetLinkRes(urlStr string, timeout int, retry int) (*extract.LinkRes, map[string]string, error) {
	if retry <= 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		res, filters, err := GetLinkResDo(urlStr, timeout)
		if err == nil {
			return res, filters, err
		}
	}

	return nil, nil, errors.New("ErrorLinkRes")
}

// GetLinkResDo 获取页面链接分组
func GetLinkResDo(urlStr string, timeout int) (*extract.LinkRes, map[string]string, error) {
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
	if resp.Success && err == nil {
		// 解析 HTML
		doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		if docErr == nil {
			doc.Find(DefaultDocRemoveTags).Remove()

			// 语言
			langRes := Lang(doc, resp.Charset.Charset, true)

			// 站内链接
			linkTitles, filters := extract.WebLinkTitles(doc, resp.RequestURL, true)

			// 链接分类
			links, _ := extract.LinkTypes(linkTitles, langRes.Lang, nil)

			return links, filters, nil
		} else {
			return nil, nil, errors.New("ErrorDocParse")
		}
	}

	return nil, nil, errors.New("ErrorRequest")
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

	if resp.Success && err == nil {
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
			content := extract.NewContent(contentDoc, langRes.Lang, title)
			news := content.News()

			return news, resp, nil
		} else {
			return nil, resp, errors.New("ErrorDocParse")
		}
	}

	return nil, nil, errors.New("ErrorRequest")
}
