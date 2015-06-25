package xmltest_test

import (
	"fmt"
	"strings"

	"github.com/rsto/xmltest"
)

func ExampleNormalizer_EqualXML() {
	var n xmltest.Normalizer
	s1 := `<root xmlns="space"/>`
	s2 := `<s:root xmlns:s="space"></s:root>`
	fmt.Println(n.EqualXML(strings.NewReader(s1), strings.NewReader(s2)))
	// Output: true
}
