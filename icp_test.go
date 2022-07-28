package spider

import (
	"bytes"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDetectIcp(t *testing.T) {
	var urlStrs = []string{
		"http://suosi.com.cn",
		"https://www.163.com",
		"https://www.sohu.com",
		"https://www.qq.com",
		"https://www.hexun.com",
		"https://www.wfmc.edu.cn/",
		"https://www.cankaoxiaoxi.com/",
	}

	for _, urlStr := range urlStrs {

		resp, err := HttpGetResp(urlStr, nil, 30000)

		t.Log(err)
		t.Log(urlStr)

		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		doc.Find("noscript,style,iframe,br,link,svg").Remove()
		icp, loc := Icp(doc)
		t.Log(icp, loc)
	}
}

func TestIcpFromText(t *testing.T) {
	texts := []string{
		"粤ICP备17055554号",
		"粤ICP备17055554-34号",
		"沪ICP备05018492",
		"粤B2-20090059",
		"京公网安备31010402001073号",
		"京公网安备-31010-4020010-73号",
		"鲁ICP备05002386鲁公网安备37070502000027号",
	}

	for _, text := range texts {
		icp, loc := IcpFromText(text)
		t.Log(icp, loc)
	}
}
