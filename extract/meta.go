package extract

import (
	"strings"
)

var HostGovCountryMap = map[string]string{
	"hk": "中国",
	"tw": "中国",
	"mo": "中国",
	"jp": "日本",
	"kr": "韩国",
	"in": "印度",
	"uk": "英国",
	"us": "美国",
	"it": "意大利",
	"es": "西班牙",
	"ru": "俄罗斯",
	"de": "德国",
	"fr": "法国",
	"th": "泰国",
	"vn": "越南",
	"sg": "新加坡",
	"au": "澳大利亚",
	"ca": "加拿大",
	"il": "以色列",
	"mm": "缅甸",
	"dz": "阿尔及利亚",
	"pl": "波兰",
	"az": "南非",
	"ng": "尼日利亚",
	"kp": "朝鲜",
	"lb": "黎巴嫩",
	"ua": "乌克兰",
	"tr": "土耳其",
	"se": "瑞典",
	"lk": "斯里兰卡",
	"si": "斯洛文尼亚",
	"sk": "斯洛伐克",
	"ro": "罗马尼亚",
	"pt": "葡萄牙",
	"ph": "菲律宾",
	"pk": "巴基斯坦",
	"py": "巴拉圭",
	"np": "尼泊尔",
	"ma": "摩洛哥",
	"my": "马来西亚",
	"lt": "立陶宛",
	"ie": "爱尔兰",
	"iq": "伊拉克",
	"ir": "伊朗",
	"id": "印度尼西亚",
	"hu": "匈牙利",
	"gr": "希腊",
	"eg": "埃及",
	"cz": "捷克",
	"hr": "克罗地亚",
	"co": "哥伦比亚",
	"cl": "智利",
	"br": "巴西",
	"bg": "保加利亚",
	"be": "比利时",
	"bd": "孟加拉国",
	"aw": "阿鲁巴",
	"am": "亚美尼亚",
	"ai": "安圭拉",
	"ao": "安哥拉",
	"al": "阿尔巴尼亚",
	"af": "阿富汗",
	"sa": "沙特阿拉伯",
	"nl": "荷兰",
}

// MetaFromHost 根据域名尽可能返回一些固定信息
func MetaFromHost(host string, lang string) (string, string, string) {
	var tld string
	var country string
	var province string
	var category string

	host = strings.ToLower(host)

	if domain, err := DomainParse(host); err == nil {
		tld = domain.TLD
	} else {
		return country, province, category
	}

	// 美国政府顶级域名
	if tld == "gov" {
		country = "美国"
		category = "政务"
		return country, province, category
	}

	// 判断是否是政府域名
	for c, zh := range HostGovCountryMap {
		gov := "gov." + c
		if tld == gov {
			country = zh
			category = "政务"

			if strings.HasSuffix(host, ".hk") && lang == "zh" {
				province = "中国香港"
			}
			if strings.HasSuffix(host, ".tw") && lang == "zh" {
				province = "中国台湾"
			}
			if strings.HasSuffix(host, ".mo") && lang == "zh" {
				province = "中国澳门"
			}
			return country, province, category
		}
	}

	if strings.HasSuffix(host, ".hk") && lang == "zh" {
		country = "中国"
		province = "中国香港"
		return country, province, category
	}

	if strings.HasSuffix(host, ".tw") && lang == "zh" {
		country = "中国"
		province = "中国台湾"
		return country, province, category
	}

	if strings.HasSuffix(host, ".mo") && lang == "zh" {
		country = "中国"
		province = "中国澳门"
		return country, province, category
	}

	if strings.HasSuffix(host, ".cn") && lang == "zh" {
		country = "中国"
		return country, province, category
	}

	if strings.HasSuffix(host, ".jp") && lang == "ja" {
		country = "日本"
		return country, province, category
	}

	if strings.HasSuffix(host, ".kr") && lang == "ko" {
		country = "韩国"
		return country, province, category
	}

	if strings.HasSuffix(host, ".uk") && lang == "en" {
		country = "英国"
		return country, province, category
	}

	if strings.HasSuffix(host, ".us") && lang == "en" {
		country = "美国"
		return country, province, category
	}

	if strings.HasSuffix(host, ".in") && lang == "hi" {
		country = "印度"
		return country, province, category
	}

	if strings.HasSuffix(host, ".es") && lang == "es" {
		country = "西班牙"
		return country, province, category
	}

	if strings.HasSuffix(host, ".ru") && lang == "ru" {
		country = "俄罗斯"
		return country, province, category
	}

	if strings.HasSuffix(host, ".de") && lang == "de" {
		country = "德国"
		return country, province, category
	}

	if strings.HasSuffix(host, ".fr") && lang == "fr" {
		country = "法国"
		return country, province, category
	}

	return country, province, category
}
