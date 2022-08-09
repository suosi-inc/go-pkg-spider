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
	Scheme     string
	HomeDomain string
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
}

func DetectDomain(domain string) (*DomainRes, error) {
	domainRes := &DomainRes{}

	schemes := []string{"http", "https"}
	homeDomains := []string{"www", ""}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 4 * 1024 * 1024,
			MaxRedirect:      2,
		},
		ForceTextContentType: true,
	}

	for _, scheme := range schemes {

		for _, homeDomain := range homeDomains {

			var urlStr string
			if homeDomain != "" {
				urlStr = scheme + "://" + homeDomain + fun.DOT + domain
			} else {
				urlStr = scheme + "://" + domain
			}

			resp, err := HttpGetResp(urlStr, req, 10000)
			if resp.Success && err == nil {
				domainRes.Domain = domain
				domainRes.State = 1
				domainRes.Scheme = scheme
				domainRes.HomeDomain = homeDomain
				domainRes.HttpCode = resp.StatusCode
				domainRes.Charset = resp.Charset

				// 解析 HTML
				u, _ := url.Parse(urlStr)
				doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
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

				return domainRes, nil
			}
		}
	}

	return nil, errors.New("domain detect error")
}
