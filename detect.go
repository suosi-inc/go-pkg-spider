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
	Domain      string
	Scheme      string
	Charset     CharsetRes
	Lang        extract.LangRes
	Country     string
	Province    string
	Category    string
	City        string
	Site        string
	Title       string
	OriginTitle string
	HttpCode    int
	State       int
	Articles    int
	Icp         string
	Level       string
}

func DetectDomain(domain string) (*DomainRes, error) {

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 4 * 1024 * 1024,
			MaxRedirect:      3,
		},
		ForceTextContentType: true,
	}

	schemes := []string{"https", "http"}
	homepages := []string{"www.", ""}

	domainRes := &DomainRes{
		Domain: domain,
	}

	for _, scheme := range schemes {

		for _, homepage := range homepages {
			urlStr := scheme + "://" + homepage + domain

			resp, err := HttpGetResp(urlStr, req, 10000)
			if err == nil {

				domainRes.State = 1
				domainRes.Scheme = scheme
				domainRes.HttpCode = resp.StatusCode
				domainRes.Charset = resp.Charset

				// 解析 HTML
				u, _ := url.Parse(urlStr)
				doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
				doc.Find(DefaultRemoveTags).Remove()

				// 语言
				langRes := extract.Lang(doc, resp.Charset.Charset, u.Hostname())
				domainRes.Lang = langRes

				// 中国的 ICP
				icp, province := extract.Icp(doc)
				if icp != "" && province != "" {
					domainRes.Country = "中国"
					domainRes.Province = extract.ProvinceMap[province]
				}

				// 尽可能的探测国家
				if domainRes.Country == "" {
					country, province, category := extract.HostMeta(u.Hostname(), langRes.Lang)
					domainRes.Country = country
					domainRes.Province = province
					domainRes.Category = category
				}

				return domainRes, nil
			}
		}
	}

	return nil, errors.New("domain detect error")
}
