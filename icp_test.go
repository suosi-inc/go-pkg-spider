package spider

import (
	"testing"
)

func TestIcp(t *testing.T) {
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
