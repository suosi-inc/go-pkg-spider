package spider

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

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
		domainRes, err := DetectDomainDo(domain, true, timeout)
		if domainRes.StatusCode != 0 || err == nil {
			return domainRes, err
		}
	}

	domainRes := &DomainRes{}
	return domainRes, errors.New("ErrorDomainDetect")
}

// DetectSubDomain 子域名探测
// DomainRes.State true 和 err nil 表示探测成功
// DomainRes.State true 可能会返回 err, 如 doc 解析失败
// DomainRes.State false 时根据 StatusCode 判断是请求是否成功或请求成功但响应失败(如404)
func DetectSubDomain(domain string, timeout int, retry int) (*DomainRes, error) {
	if retry == 0 {
		retry = 1
	}

	for i := 0; i < retry; i++ {
		domainRes, err := DetectDomainDo(domain, false, timeout)
		if domainRes.StatusCode != 0 || err == nil {
			return domainRes, err
		}
	}

	domainRes := &DomainRes{}
	return domainRes, errors.New("ErrorDomainDetect")
}

func DetectDomainDo(domain string, isTop bool, timeout int) (*DomainRes, error) {
	if timeout == 0 {
		timeout = 10000
	}

	domainRes := &DomainRes{}

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 10 * 1024 * 1024,
			MaxRedirect:      3,
		},
		ForceTextContentType: true,
	}

	scheme := "http"

	// 是否进行首页探测
	var homes []string
	if isTop {
		homes = []string{"www", ""}
	} else {
		homes = []string{""}
	}

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

			// 如果发生 HTTP 跳转, 则重新设置 homeDomain, 判断跳转后是否是同一个主域名, 如果域名改变则记录并返回错误
			domainRes.HomeDomain = homeDomain
			requestHostname := resp.RequestURL.Hostname()
			if domainRes.HomeDomain != requestHostname {
				requestTopDomain := extract.DomainTop(requestHostname)
				if requestTopDomain != "" && requestTopDomain != domain {
					// 验证主机名
					if RegexHostnameIpPattern.MatchString(requestHostname) {
						return domainRes, errors.New("ErrorRedirectHost")
					}
					// 验证非常规端口
					if resp.RequestURL.Port() != "" {
						return domainRes, errors.New("ErrorRedirectHost")
					}

					return domainRes, errors.New("ErrorRedirect:" + requestTopDomain)
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
				doc.Find(DefaultDocRemoveTags).Remove()

				// 具有 HTML 跳转属性, HTTP 无法自动处理永远返回错误, 判断跳转后是否是同一个主域名, 记录并返回
				if refresh, exists := doc.Find("meta[http-equiv='refresh' i]").Attr("content"); exists {
					refreshMatch := regexMetaRefreshPattern.FindStringSubmatch(refresh)
					if len(refreshMatch) > 1 {
						refreshUrl := refreshMatch[1]
						if r, err := fun.UrlParse(refreshUrl); err == nil {
							refreshHostname := r.Hostname()
							refreshTopDomain := extract.DomainTop(refreshHostname)
							if refreshTopDomain != "" && refreshTopDomain != domain {
								// 验证主机名
								if RegexHostnameIpPattern.MatchString(refreshHostname) {
									return domainRes, errors.New("ErrorMetaJumpHost")
								}
								// 验证非常规端口
								if r.Port() != "" {
									return domainRes, errors.New("ErrorMetaJumpHost")
								}

								return domainRes, errors.New("ErrorMetaJump:" + refreshTopDomain)
							}
						}
						return domainRes, errors.New("ErrorMetaJump")
					}
				}

				// 中国 ICP 解析
				icp, province := extract.Icp(doc)
				if icp != "" && province != "" {
					domainRes.Country = "中国"
					domainRes.Icp = icp
					domainRes.Province = extract.ProvinceShortMap[province]
				}

				// 语言
				langRes := Lang(doc, resp.Charset.Charset, true)
				domainRes.Lang = langRes

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

func DetectFriendDomain(domain string, timeout int, retry int) (map[string]string, error) {
	if retry == 0 {
		retry = 1
	}

	friendDomains := make(map[string]string, 0)

	for i := 0; i < retry; i++ {
		friendDomains, err := DetectFriendDomainDo(domain, timeout)
		if err == nil {
			return friendDomains, err
		}
	}

	return friendDomains, errors.New("ErrorDomainDetect")
}

func DetectFriendDomainDo(domain string, timeout int) (map[string]string, error) {
	if timeout == 0 {
		timeout = 10000
	}

	friendDomains := make(map[string]string, 0)

	req := &HttpReq{
		HttpReq: &fun.HttpReq{
			MaxContentLength: 10 * 1024 * 1024,
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

		if resp.Success && err == nil {

			doc, docErr := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
			if docErr == nil {
				doc.Find(DefaultDocRemoveTags).Remove()

				// 非限制域名所有链接
				linkTitles, _ := extract.WebLinkTitles(doc, resp.RequestURL, false)

				if len(linkTitles) > 0 {
					for link, title := range linkTitles {
						if link == "" || title == "" {
							continue
						}

						u, e := fun.UrlParse(link)
						if e != nil {
							continue
						}

						// 验证非常规端口
						if u.Port() != "" {
							continue
						}

						// 验证主机名
						if fun.Matches(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, u.Hostname()) {
							continue
						}

						pathDir := strings.TrimSpace(u.Path)
						if pathDir == "" || pathDir == fun.SLASH || pathDir == "/index.html" || pathDir == "/index.htm" || pathDir == "/index.shtml" {
							hostname := u.Hostname()
							domainTop := extract.DomainTop(hostname)
							baseDomainTop := domain
							if domainTop != baseDomainTop {
								friendDomains[domainTop] = title
							}
						}
					}
				}

				return friendDomains, nil
			} else {
				return friendDomains, errors.New("ErrorDocParse")
			}
		} else {
			return friendDomains, err
		}
	}

	return friendDomains, errors.New("ErrorDomainDetect")
}
