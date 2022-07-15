package domain

import (
	"fmt"
	"net/url"
	"testing"
)

func TestUrlParse(t *testing.T) {
	u, _ := url.Parse("https://www.test.com:8080/a")
	fmt.Println(u.Host)
	fmt.Println(u.Hostname())
}

func TestDomainParse(t *testing.T) {

	fmt.Println()

	fmt.Println(DomainParse("https://www.google.com"))
}
