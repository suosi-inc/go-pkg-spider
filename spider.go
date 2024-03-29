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

// GetLinkData 获取页面链接数据
func GetLinkData(urlStr string, strictDomain bool, timeout int, retry int) (*LinkData, error) {
	if retry <= 0 {
		retry = 1
	}

	errs := make([]string, 0)

	for i := 0; i < retry; i++ {
		linkData, err := GetLinkDataDo(urlStr, strictDomain, nil, nil, timeout)
		if err == nil {
			return linkData, err
		} else {
			errs = append(errs, err.Error())
		}
	}

	return nil, errors.New("ErrorLinkRes" + fun.ToString(errs))
}

// GetLinkDataWithReq 获取页面链接数据
func GetLinkDataWithReq(urlStr string, strictDomain bool, req *HttpReq, timeout int, retry int) (*LinkData, error) {
	if retry <= 0 {
		retry = 1
	}

	errs := make([]string, 0)

	for i := 0; i < retry; i++ {
		linkData, err := GetLinkDataDo(urlStr, strictDomain, nil, req, timeout)
		if err == nil {
			return linkData, err
		} else {
			errs = append(errs, err.Error())
		}
	}

	return nil, errors.New("ErrorLinkRes" + fun.ToString(errs))
}

// GetLinkDataWithReqAndRule 获取页面链接数据
func GetLinkDataWithReqAndRule(urlStr string, strictDomain bool, rules extract.LinkTypeRule, req *HttpReq, timeout int, retry int) (*LinkData, error) {
	if retry <= 0 {
		retry = 1
	}

	errs := make([]string, 0)

	for i := 0; i < retry; i++ {
		linkData, err := GetLinkDataDo(urlStr, strictDomain, rules, req, timeout)
		if err == nil {
			return linkData, err
		} else {
			errs = append(errs, err.Error())
		}
	}

	return nil, errors.New("ErrorLinkRes" + fun.ToString(errs))
}

// GetLinkDataWithRule 获取页面链接数据
func GetLinkDataWithRule(urlStr string, strictDomain bool, rules extract.LinkTypeRule, timeout int, retry int) (*LinkData, error) {
	if retry <= 0 {
		retry = 1
	}

	errs := make([]string, 0)

	for i := 0; i < retry; i++ {
		linkData, err := GetLinkDataDo(urlStr, strictDomain, rules, nil, timeout)
		if err == nil {
			return linkData, err
		} else {
			errs = append(errs, err.Error())
		}
	}

	return nil, errors.New("ErrorLinkRes" + fun.ToString(errs))
}

// GetLinkDataDo 获取页面链接数据
func GetLinkDataDo(urlStr string, strictDomain bool, rules extract.LinkTypeRule, req *HttpReq, timeout int) (*LinkData, error) {
	if timeout == 0 {
		timeout = 10000
	}

	if req == nil {
		req = &HttpReq{
			HttpReq: &fun.HttpReq{
				MaxContentLength: HttpDefaultMaxContentLength,
				MaxRedirect:      3,
			},
			ForceTextContentType: true,
		}
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

// GetNews 获取链接新闻数据
func GetNews(urlStr string, title string, timeout int, retry int) (*extract.News, *HttpResp, error) {
	if retry <= 0 {
		retry = 1
	}

	errs := make([]string, 0)

	for i := 0; i < retry; i++ {
		news, resp, err := GetNewsDo(urlStr, title, nil, timeout)
		if err == nil {
			return news, resp, nil
		} else {
			errs = append(errs, err.Error())
		}
	}

	return nil, nil, errors.New("ErrorRequest" + fun.ToString(errs))
}

// GetNewsWithReq 获取链接新闻数据
func GetNewsWithReq(urlStr string, title string, req *HttpReq, timeout int, retry int) (*extract.News, *HttpResp, error) {
	if retry <= 0 {
		retry = 1
	}

	errs := make([]string, 0)

	for i := 0; i < retry; i++ {
		news, resp, err := GetNewsDo(urlStr, title, req, timeout)
		if err == nil {
			return news, resp, nil
		} else {
			errs = append(errs, err.Error())
		}
	}

	return nil, nil, errors.New("ErrorRequest" + fun.ToString(errs))
}

// GetNewsDo 获取链接新闻数据
func GetNewsDo(urlStr string, title string, req *HttpReq, timeout int) (*extract.News, *HttpResp, error) {
	return getNewsDoTop(urlStr, title, req, timeout, true)
}

// getNewsDoTop 获取链接新闻数据
func getNewsDoTop(urlStr string, title string, req *HttpReq, timeout int, top bool) (*extract.News, *HttpResp, error) {
	if timeout == 0 {
		timeout = HttpDefaultTimeOut
	}

	if req == nil {
		req = &HttpReq{
			HttpReq: &fun.HttpReq{
				MaxContentLength: HttpDefaultMaxContentLength,
				MaxRedirect:      2,
			},
			ForceTextContentType: true,
		}
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
								return getNewsDoTop(refreshUrl, title, req, timeout, false)
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
