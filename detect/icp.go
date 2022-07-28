package detect

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	RegexIcpPattern = `(?i)(京|津|冀|晋|蒙|辽|吉|黑|沪|苏|浙|皖|闽|赣|鲁|豫|鄂|湘|粤|桂|琼|川|蜀|贵|黔|云|滇|渝|藏|陇|甘|陕|秦|青|宁|新)ICP(备|证|备案)?[0-9\-]+`
	RegexGaPattern  = `(?i)(京|津|冀|晋|蒙|辽|吉|黑|沪|苏|浙|皖|闽|赣|鲁|豫|鄂|湘|粤|桂|琼|川|蜀|贵|黔|云|滇|渝|藏|陇|甘|陕|秦|青|宁|新)公网安备[0-9\-]+`
	RegexDxPattern  = `(?i)(京|津|冀|晋|蒙|辽|吉|黑|沪|苏|浙|皖|闽|赣|鲁|豫|鄂|湘|粤|桂|琼|川|蜀|贵|黔|云|滇|渝|藏|陇|甘|陕|秦|青|宁|新)B2-[0-9]+`
)

// Icp 返回网站备案相关的信息
func Icp(doc *goquery.Document) (string, string) {
	text := doc.Find("body").Text()
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "\t", "")
	text = strings.ReplaceAll(text, " ", "")

	return IcpFromText(text)

}

func IcpFromText(text string) (icp string, loc string) {
	// 优先匹配ICP
	matches := regexp.MustCompile(RegexIcpPattern).FindStringSubmatch(text)
	if len(matches) > 1 {
		icp = matches[0]
		loc = matches[1]
		return
	}

	// 匹配公网安备
	if icp == "" {
		matches = regexp.MustCompile(RegexGaPattern).FindStringSubmatch(text)
		if len(matches) > 1 {
			icp = matches[0]
			loc = matches[1]
		}
	}

	// 匹配电信增值业务
	if icp == "" {
		matches = regexp.MustCompile(RegexDxPattern).FindStringSubmatch(text)
		if len(matches) > 1 {
			icp = matches[0]
			loc = matches[1]
			return
		}
	}

	return icp, loc
}
