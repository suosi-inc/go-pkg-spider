package extract

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/x-funs/go-fun"
)

var (
	ProvinceShortMap = map[string]string{
		"京": "北京",
		"津": "天津",
		"沪": "上海",
		"渝": "重庆",
		"黑": "黑龙江",
		"吉": "吉林",
		"辽": "辽宁",
		"冀": "河北",
		"豫": "河南",
		"鲁": "山东",
		"晋": "山西",
		"陕": "陕西",
		"秦": "陕西",
		"蒙": "内蒙古",
		"宁": "宁夏",
		"陇": "甘肃",
		"甘": "甘肃",
		"新": "新疆",
		"青": "青海",
		"藏": "西藏",
		"鄂": "湖北",
		"皖": "安徽",
		"苏": "江苏",
		"浙": "浙江",
		"闽": "福建",
		"湘": "湖南",
		"赣": "江西",
		"川": "四川",
		"蜀": "四川",
		"黔": "贵州",
		"贵": "贵州",
		"滇": "云南",
		"云": "云南",
		"粤": "广东",
		"桂": "广西",
		"琼": "海南",
		"港": "中国香港",
		"澳": "中国澳门",
		"台": "中国台湾",
	}
)

const (
	RegexIcp   = `(?i)(京|津|冀|晋|蒙|辽|吉|黑|沪|苏|浙|皖|闽|赣|鲁|豫|鄂|湘|粤|桂|琼|川|蜀|贵|黔|云|滇|渝|藏|陇|甘|陕|秦|青|宁|新)ICP(备|证|备案)?[0-9]+`
	RegexIcpGa = `(?i)(京|津|冀|晋|蒙|辽|吉|黑|沪|苏|浙|皖|闽|赣|鲁|豫|鄂|湘|粤|桂|琼|川|蜀|贵|黔|云|滇|渝|藏|陇|甘|陕|秦|青|宁|新)公网安备[0-9]+`
	RegexIcpDx = `(?i)(京|津|冀|晋|蒙|辽|吉|黑|沪|苏|浙|皖|闽|赣|鲁|豫|鄂|湘|粤|桂|琼|川|蜀|贵|黔|云|滇|渝|藏|陇|甘|陕|秦|青|宁|新)B2-[0-9]+`
)

var (
	RegexIcpPattern   = regexp.MustCompile(RegexIcp)
	RegexIcpGaPattern = regexp.MustCompile(RegexIcpGa)
	RegexIcpDxPattern = regexp.MustCompile(RegexIcpDx)
)

// Icp 返回网站备案相关的信息
func Icp(doc *goquery.Document) (string, string) {
	text := doc.Find("body").Text()
	text = fun.RemoveLines(text)
	text = strings.NewReplacer(fun.TAB, "", fun.SPACE, "").Replace(text)

	return IcpFromText(text)

}

// IcpFromText 提取文本中备案相关的信息
func IcpFromText(text string) (string, string) {
	var icp, loc string

	// 优先匹配ICP
	matches := RegexIcpPattern.FindStringSubmatch(text)
	if len(matches) > 1 {
		icp = matches[0]
		loc = matches[1]
	}

	// 匹配公网安备
	if icp == "" {
		matches = RegexIcpGaPattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			icp = matches[0]
			loc = matches[1]
		}
	}

	// 匹配电信增值业务
	if icp == "" {
		matches = RegexIcpDxPattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			icp = matches[0]
			loc = matches[1]
		}
	}

	return icp, loc
}
