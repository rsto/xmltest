package xmltest

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	testCases := []struct {
		desc    string
		n       Normalizer
		in      string
		wantXML string
		wantErr error
	}{{
		desc:    "single root element in default namespace",
		in:      "<root/>",
		wantXML: "<root></root>",
	}, {
		desc:    "single root element in namespace",
		in:      `<root xmlns="space"/>`,
		wantXML: `<space:root xmlns:space="space"></space:root>`,
	}, {
		desc:    "single root element in prefixed namespace",
		in:      `<s:root xmlns:s="space"/>`,
		wantXML: `<space:root xmlns:space="space"></space:root>`,
	}, {
		desc: "element with inherited prefixed namespace",
		in:   `<s:root xmlns:s="space" xmlns:f="foons"><f:foo/></s:root>`,
		wantXML: `` +
			`<space:root xmlns:space="space">` +
			`<foons:foo xmlns:foons="foons"></foons:foo>` +
			`</space:root>`,
	}, {
		desc:    "preserve attributes except xmlns",
		in:      `<root xmlns:i="ignored" a="foo"/>`,
		wantXML: `<root a="foo"></root>`,
	}, {
		desc:    "preserve attribute namespaces",
		in:      `<root xmlns:i="ignored" xmlns:b="bar" b:a="foo"/>`,
		wantXML: `<root xmlns:bar="bar" bar:a="foo"></root>`,
	}, {
		desc: "sort attributes in lexical order",
		in: `<root` +
			` xmlns:a="a" xmlns:b="b" xmlns:c="c"` +
			` c:bar="bar" a:baz="baz" b:bam="bam" a:bam="bam">` +
			`</root>`,
		wantXML: `` +
			`<root` +
			` xmlns:c="c" xmlns:b="b" xmlns:a="a"` +
			` a:bam="bam" a:baz="baz" b:bam="bam" c:bar="bar">` +
			`</root>`,
	}, {
		desc:    "omit directives",
		in:      `<!DOCTYPE foo><root/>`,
		wantXML: `<root></root>`,
	}, {
		desc:    "omit preamble",
		in:      `<?xml version="1.0"?><root/>`,
		wantXML: `<root></root>`,
	}, {
		desc:    "omit processing instruction",
		in:      `<root><?foo?></root>`,
		wantXML: `<root></root>`,
	}, {
		desc:    "keep comments by default",
		in:      `<root><!-- a comment --></root>`,
		wantXML: `<root><!-- a comment --></root>`,
	}, {
		desc:    "omit comments if requested",
		n:       Normalizer{OmitComments: true},
		in:      `<root><!-- a comment --></root>`,
		wantXML: `<root></root>`,
	}, {
		desc:    "keep whitespace by default",
		in:      `<root>  <foo>  </foo>  </root>`,
		wantXML: `<root>  <foo>  </foo>  </root>`,
	}, {
		desc:    "omit whitespace if requested",
		n:       Normalizer{OmitWhitespace: true},
		in:      `<root>  <foo>  </foo>  </root>`,
		wantXML: `<root><foo></foo></root>`,
	}, {
		desc:    "omit whitespace cdata but don't trim",
		n:       Normalizer{OmitWhitespace: true},
		in:      `<root>  <foo>  </foo> a  </root>`,
		wantXML: `<root><foo></foo> a  </root>`,
	}, {
		desc:    "bad: make decoder fail with a syntax error",
		in:      "<root></foo>",
		wantErr: errors.New("some error"),
	}}

	for _, tc := range testCases {
		var b bytes.Buffer
		err := tc.n.Normalize(&b, strings.NewReader(tc.in))
		if tc.wantErr != nil {
			if err == nil {
				t.Errorf("%s: got nil error, want %v", tc.desc, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: got err %v, want nil", tc.desc, err)
			continue
		}
		if got, want := b.String(), tc.wantXML; got != want {
			t.Errorf("%s:\ngot  %s\nwant %s", tc.desc, got, tc.wantXML)
		}
	}
}

type errorWriter struct{}

func (w *errorWriter) Write(buf []byte) (n int, err error) {
	return 0, errors.New("Why have I always been a failure?")
}

func TestNormalizeFailWriter(t *testing.T) {
	// This test makes the Normalizer fail to write to the Writer.
	// TBH it mainly exists to pimp test coverage.
	var n Normalizer
	err := n.Normalize(&errorWriter{}, strings.NewReader("<root/>"))
	if err == nil {
		t.Errorf("failwriter: got nil error, want non-nil")
	}
}

func TestEqualXML(t *testing.T) {
	testCases := []struct {
		desc      string
		a, b      string
		n         Normalizer
		wantEqual bool
		wantErr   error
	}{{
		desc:      "identity",
		a:         "<root/>",
		b:         "<root/>",
		wantEqual: true,
		// TODO define more test cases for EqualXML
	}}

	for _, tc := range testCases {
		got, err := tc.n.EqualXML(strings.NewReader(tc.a), strings.NewReader(tc.b))
		if tc.wantErr != nil {
			if err == nil {
				t.Errorf("%s: got nil error, want %v", tc.desc, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: got err %v, want nil", tc.desc, err)
			continue
		}
		if got != tc.wantEqual {
			t.Errorf("%s:\ngot  %v\nwant %v", tc.desc, got, tc.wantEqual)
		}
	}
}
