package sourceanalysis

import (
	"context"
	"encoding/xml"
)

type Analysis struct {
	XMLName xml.Name `xml:"wave-highlight"`
	Engine  string   `xml:"engine,attr"`
	ABI     int      `xml:"abi,attr"`
	Tokens  []Token  `xml:"tokens>token"`
}

type Token struct {
	Kind  string `xml:"kind,attr"`
	Start int    `xml:"start,attr"`
	End   int    `xml:"end,attr"`
}

type Analyzer interface {
	Analyze(context.Context, []byte) (Analysis, error)
	Close() error
}
