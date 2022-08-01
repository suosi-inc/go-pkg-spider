package extract

import "testing"

func TestTitleClean(t *testing.T) {
	strs := map[string]string{
		"“暴徒试图杀死他！”阿拉木图市长在1月5日的暗杀企图中幸存_网易订阅":                                                   "zh",
		"“暴徒试图杀死他！”阿拉木图市长在1月5日的暗杀企图中幸存 - 网易订阅":                                                 "zh",
		"北极圈内最高温达到38℃ 北极熊还好吗？南极情况怎么样？_科技频道_中国青年网":                                              "zh",
		"About the Project on Nuclear Issues | Center for Strategic and International Studies": "en",
	}

	for str, l := range strs {
		t.Log(WebTitleClean(str, l))
	}
}
