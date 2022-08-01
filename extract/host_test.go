package extract

import "testing"

func TestHostMeta(t *testing.T) {
	hosts := []string{
		"matichon.co.th",
		"wanbao.com.sg",
		"wanbao.com.sg",
		"waou.com.mo",
		"archives.gov.mo",
		"mfa.gov.sg",
		"nasa.gov",
	}

	for _, host := range hosts {

		t.Log(HostMeta(host, ""))
	}
}
