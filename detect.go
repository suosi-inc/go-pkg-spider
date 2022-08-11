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
	Domain     string
	HomeDomain string
	Scheme     string
	Charset    CharsetRes
	Lang       LangRes
	Country    string
	Province   string
	Category   string
	Title      string
	TitleClean string
	Icp        string
	State      int
	HttpCode   int
	Articles   int
	SubDomains map[string]bool
	ErrorPos   string
}

func DetectDomain(domain string, timeout int, retry int) (*DomainRes, error) {
	if retry == 0 {
		retry = 1
	}
	for i := 0; i < retry; i++ {
		domainRes, err := DetectDomainDo(domain, timeout)
		if err == nil {
			return domainRes, nil
		}
	}
	return nil, errors.New("ErrorDomainDetect")
}

func DetectDomainDo(domain string, timeout int) (*DomainRes, error) {
	if timeout == 0 {
		timeout = 10000
	}

	domainRes := &DomainRes{}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 2 * 1024 * 1024,
			MaxRedirect:      1,
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
		if resp.Success && err == nil {
			domainRes.Domain = domain
			domainRes.HomeDomain = homeDomain
			domainRes.State = 1
			domainRes.HttpCode = resp.StatusCode
			domainRes.Charset = resp.Charset
			domainRes.Scheme = scheme

			// 如果发生了协议跳转，则重新设置 scheme
			if domainRes.Scheme != resp.RequestURL.Scheme {
				domainRes.Scheme = resp.RequestURL.Scheme
			}

			// 如果发生了跳转，则重新设置 homeDomain
			if domainRes.HomeDomain != resp.RequestURL.Hostname() {
				domainRes.HomeDomain = resp.RequestURL.Hostname()
			}

			// 解析 HTML
			u, _ := url.Parse(urlStr)
			doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
			if docErr == nil {
				doc.Find(DefaultRemoveTags).Remove()

				// 语言
				langRes := Lang(doc, resp.Charset.Charset, u.Hostname())
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

				// 标题
				domainRes.Title = extract.WebTitle(doc, 0)
				domainRes.TitleClean = extract.WebTitleClean(domainRes.Title, langRes.Lang)

				linkTitles, _ := extract.WebLinkTitles(doc, urlStr, true)

				links, subDomains := extract.LinkTypes(linkTitles, langRes.Lang, nil)

				domainRes.Articles = len(links.Content)
				domainRes.SubDomains = subDomains

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
