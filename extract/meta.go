package extract

import (
	"strings"
)

var HostGovCountryMap = map[string]string{
	"cn": "中国",
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
}

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

	// .cn 在中国必须要备案且实名
	if tld == "cn" {
		country = "中国"
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

	if strings.HasSuffix(host, ".it") && lang == "it" {
		country = "意大利"
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

	if strings.HasSuffix(host, ".th") && lang == "th" {
		country = "泰国"
		return country, province, category
	}

	if strings.HasSuffix(host, ".vn") && lang == "vi" {
		country = "越南"
		return country, province, category
	}

	return country, province, category
}
