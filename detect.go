package spider

import (
	"bytes"
	"errors"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/suosi-inc/go-pkg-spider/extract"
	"github.com/x-funs/go-fun"
)

type DomainRes struct {
	Domain       string
	HomeDomain   string
	Scheme       string
	Charset      CharsetRes
	Lang         LangRes
	Country      string
	Province     string
	Category     string
	Title        string
	TitleClean   string
	Description  string
	Icp          string
	State        bool
	StatusCode   int
	ContentCount int
	ListCount    int
	SubDomains   map[string]bool
}

// DetectDomain 域名探测
// DomainRes.State true 和 err nil 表示探测成功
// DomainRes.State true 可能会返回 err, 如 doc 解析失败
// DomainRes.State false 时根据 StatusCode 判断是请求是否成功或请求成功但响应失败(如404)
func DetectDomain(domain string, timeout int, retry int) (*DomainRes, error) {
	if retry == 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		domainRes, err := DetectDomainDo(domain, timeout)
		if domainRes.StatusCode != 0 || err == nil {
			return domainRes, err
		}
	}

	domainRes := &DomainRes{}
	return domainRes, errors.New("ErrorDomainDetect")
}

func DetectDomainDo(domain string, timeout int) (*DomainRes, error) {
	if timeout == 0 {
		timeout = 10000
	}

	domainRes := &DomainRes{}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 4 * 1024 * 1024,
			MaxRedirect:      3,
		},
		ForceTextContentType: true,
	}

	scheme := "http"
	homes := []string{"www", ""}

	for _, home := range homes {

		var urlStr string
		var homeDomain string
		if home != "" {
			homeDomain = home + fun.DOT + domain
			urlStr = scheme + "://" + homeDomain
		} else {
			homeDomain = domain
			urlStr = scheme + "://" + homeDomain
		}

		resp, err := HttpGetResp(urlStr, req, timeout)

		domainRes.StatusCode = resp.StatusCode

		if resp.Success && err == nil {
			domainRes.Domain = domain

			// 如果发生了跳转, 则重新设置 homeDomain, 前提是还是同一个主域名
			domainRes.HomeDomain = homeDomain
			requestHostname := resp.RequestURL.Hostname()
			if domainRes.HomeDomain != requestHostname {
				if !fun.HasSuffixCase(requestHostname, "."+domain) {
					return domainRes, errors.New("ErrorRedirect:" + requestHostname)
				}

				domainRes.HomeDomain = requestHostname
			}

			// 如果发生了协议跳转, 则重新设置 scheme
			domainRes.Scheme = scheme
			if domainRes.Scheme != resp.RequestURL.Scheme {
				domainRes.Scheme = resp.RequestURL.Scheme
			}

			// 字符集
			domainRes.Charset = resp.Charset

			// 解析 HTML
			u, _ := url.Parse(urlStr)
			doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
			if docErr == nil {
				doc.Find(DefaultRemoveTags).Remove()

				// 语言
				langRes := Lang(doc, resp.Charset.Charset)
				domainRes.Lang = langRes

				// 中国 ICP 解析
				icp, province := extract.Icp(doc)
				if icp != "" && province != "" {
					domainRes.Country = "中国"
					domainRes.Icp = icp
					domainRes.Province = extract.ProvinceShortMap[province]
				}

				// 尽可能的探测一些信息国家/省份/类别
				if domainRes.Country == "" {
					country, province, category := extract.MetaFromHost(u.Hostname(), langRes.Lang)
					domainRes.Country = country
					domainRes.Province = province
					domainRes.Category = category
				}

				// 标题摘要
				domainRes.Title = extract.WebTitle(doc, 0)
				domainRes.TitleClean = extract.WebTitleClean(domainRes.Title, langRes.Lang)
				domainRes.Description = extract.WebDescription(doc, 0)

				// 站内链接
				linkTitles, _ := extract.WebLinkTitles(doc, resp.RequestURL, true)

				// 链接分类
				links, subDomains := extract.LinkTypes(linkTitles, langRes.Lang, nil)

				domainRes.ContentCount = len(links.Content)
				domainRes.ListCount = len(links.List)
				domainRes.SubDomains = subDomains

				domainRes.State = true

				return domainRes, nil
			} else {
				return domainRes, errors.New("ErrorDocParse")
			}
		} else {
			return domainRes, err
		}
	}

	return domainRes, errors.New("ErrorDomainDetect")
}
