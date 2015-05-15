package webmention

import (
	"fmt"
	"testing"
)

func TestHeader(t *testing.T) {
	test := []string{`<http://alice.host/webmention-endpoint>; rel="webmention"`}

	links := GetHeaderLinks(test)
	for _, l := range links {
		fmt.Println(l.URL)
		fmt.Println(l.Params)
	}

}
