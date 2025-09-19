package models

type CRDAttribute struct {
	Name     string `xml:"name,attr"`
	Value    string `xml:",chardata"`
	Disabled bool   `xml:"-"`
}

type SubOperation struct {
	CRDAttribute
	Time string `xml:"time,attr"`
}
