// Code generated by xgen. DO NOT EDIT.

package schema

import (
	"encoding/xml"
)

// MyType2 ...
type MyType2 struct {
	XMLName xml.Name `xml:"myType2"`
	Length  int      `xml:"length,attr,omitempty"`
	Value   []byte   `xml:",chardata"`
}

// MyType3 ...
type MyType3 struct {
	XMLName xml.Name `xml:"myType3"`
	Length  int      `xml:"length,attr,omitempty"`
	Value   string   `xml:",chardata"`
}

// MyType4 ...
type MyType4 struct {
	XMLName   xml.Name `xml:"myType4"`
	Title     string   `xml:"title"`
	Blob      []byte   `xml:"blob"`
	Timestamp string   `xml:"timestamp"`
}

// MyType6 ...
type MyType6 struct {
	Code       string `xml:"code,attr,omitempty"`
	Identifier int    `xml:"identifier,attr,omitempty"`
}

// MyType7 ...
type MyType7 struct {
	Origin string `xml:"origin,attr"`
	Value  string `xml:",chardata"`
}

// TopLevel ...
type TopLevel struct {
	Cost        float64   `xml:"cost,attr,omitempty"`
	LastUpdated string    `xml:"LastUpdated,attr,omitempty"`
	Nested      *MyType7  `xml:"nested,omitempty"`
	Nested2     MyType7   `xml:"nested2"`
	MyType1     [][]byte  `xml:"myType1,omitempty"`
	MyType2     []MyType2 `xml:"myType2,omitempty"`
	MyString    *string   `xml:"myString,omitempty"`
	MyInt       *int      `xml:"myInt,omitempty"`
	*MyType6
}
