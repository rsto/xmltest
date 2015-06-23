// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package xmltest provides utilities for XML testing.
package xmltest

import (
	"bytes"
	"encoding/xml"
	"io"
	"sort"
)

// Normalizer normalizes XML.
type Normalizer struct {
	// OmitWhitespace instructs to ignore whitespace between element tags.
	OmitWhitespace bool
	// OmitComments instructs to ignore XML comments.
	OmitComments bool
}

// Normalize writes the normalized XML content of r to w. It applies the
// following rules
//
//     * Rename namespace prefixes according to an internal heuristic.
//     * Remove unnecessary namespace declarations.
//     * Sort attributes in XML start elements in lexical order of their
//       fully qualified name.
//     * Remove XML directives and processing instructions.
//     * Remove CDATA between XML tags that only contains whitespace, if
//       instructed to do so.
//     * Remove comments, if instructed to do so.
//
// Note that the normalized XML content might differ from canonicalized XML
// as defined by W3C.
func (n *Normalizer) Normalize(w io.Writer, r io.Reader) error {
	d := xml.NewDecoder(r)
	e := xml.NewEncoder(w)
	for {
		t, err := d.Token()
		if err != nil {
			if t == nil && err == io.EOF {
				break
			}
			return err
		}
		switch val := t.(type) {
		case xml.Directive, xml.ProcInst:
			continue
		case xml.Comment:
			if n.OmitComments {
				continue
			}
		case xml.CharData:
			if n.OmitWhitespace && len(bytes.TrimSpace(val)) == 0 {
				continue
			}
		case xml.StartElement:
			start, _ := xml.CopyToken(val).(xml.StartElement)
			attr := start.Attr[:0]
			for _, a := range start.Attr {
				if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
					continue
				}
				attr = append(attr, a)
			}
			sort.Sort(byName(attr))
			start.Attr = attr
			t = start
		}
		err = e.EncodeToken(t)
		if err != nil {
			return err
		}
	}
	return e.Flush()
}

// EqualXML tests for equality of the normalized XML contents of a and b.
func (n *Normalizer) EqualXML(a, b io.Reader) (bool, error) {
	var buf bytes.Buffer
	if err := n.Normalize(&buf, a); err != nil {
		return false, err
	}
	normA := buf.String()
	buf.Reset()
	if err := n.Normalize(&buf, b); err != nil {
		return false, err
	}
	normB := buf.String()
	return normA == normB, nil
}

type byName []xml.Attr

func (a byName) Len() int      { return len(a) }
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool {
	if a[i].Name.Space != a[j].Name.Space {
		return a[i].Name.Space < a[j].Name.Space
	}
	return a[i].Name.Local < a[j].Name.Local
}
