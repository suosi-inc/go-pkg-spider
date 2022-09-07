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

func GetNews(link string, title string, timeout int) (*extract.News, *HttpResp, error) {
	return GetNewsDo(link, title, timeout, true)
}

func GetNewsDo(link string, title string, timeout int, top bool) (*extract.News, *HttpResp, error) {
	if timeout == 0 {
		timeout = HttpDefaultTimeOut
	}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 10 * 1024 * 1024,
			MaxRedirect:      1,
		},
		ForceTextContentType: true,
	}

	resp, err := HttpGetResp(link, req, timeout)

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
								return GetNewsDo(refreshUrl, title, timeout, false)
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
			return nil, resp, errors.New("ErrorParse")
		}
	}

	return nil, nil, errors.New("ErrorRequest")
}
