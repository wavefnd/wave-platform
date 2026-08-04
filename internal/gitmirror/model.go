package gitmirror

import (
	"encoding/xml"

	"github.com/wavefnd/wave-platform/internal/sourceanalysis"
)

type Repository struct {
	XMLName       xml.Name `xml:"repository"`
	ID            string   `xml:"id"`
	Owner         string   `xml:"owner"`
	Name          string   `xml:"name"`
	Description   string   `xml:"description"`
	RemoteURL     string   `xml:"remote-url,omitempty"`
	DefaultBranch string   `xml:"default-branch"`
	HeadOID       string   `xml:"head-oid,omitempty"`
	HeadCommit    *Commit  `xml:"head-commit,omitempty"`
	LastFetchedAt string   `xml:"last-fetched-at,omitempty"`
	Status        string   `xml:"status"`
}

type Commit struct {
	OID        string `xml:"oid"`
	ShortOID   string `xml:"short-oid"`
	Author     string `xml:"author"`
	AuthoredAt string `xml:"authored-at"`
	Subject    string `xml:"subject"`
}

type ChangedFile struct {
	Status  string `xml:"status"`
	Path    string `xml:"path"`
	OldPath string `xml:"old-path,omitempty"`
}

type CommitDetail struct {
	XMLName        xml.Name      `xml:"commit-detail"`
	Commit         Commit        `xml:"commit"`
	Body           string        `xml:"body,omitempty"`
	Parents        []string      `xml:"parents>parent"`
	Files          []ChangedFile `xml:"files>file"`
	Patch          string        `xml:"patch,omitempty"`
	PatchTruncated bool          `xml:"patch-truncated"`
}

type TreeEntry struct {
	Name       string  `xml:"name"`
	Path       string  `xml:"path"`
	Type       string  `xml:"type"`
	OID        string  `xml:"oid"`
	Size       int64   `xml:"size,omitempty"`
	Binary     bool    `xml:"binary,omitempty"`
	Generated  bool    `xml:"generated,omitempty"`
	LastCommit *Commit `xml:"last-commit,omitempty"`
}

type LanguageStat struct {
	Name       string  `xml:"name"`
	Bytes      int64   `xml:"bytes"`
	Files      int     `xml:"files"`
	Percentage float64 `xml:"percentage"`
}

type Tree struct {
	Repository Repository     `xml:"repository"`
	Ref        string         `xml:"ref"`
	Path       string         `xml:"path"`
	Commit     Commit         `xml:"commit"`
	Entries    []TreeEntry    `xml:"entries>entry"`
	Readme     *Blob          `xml:"readme,omitempty"`
	Languages  []LanguageStat `xml:"languages>language"`
}

type Blob struct {
	Path          string                   `xml:"path"`
	OID           string                   `xml:"oid"`
	Size          int64                    `xml:"size"`
	Binary        bool                     `xml:"binary"`
	Truncated     bool                     `xml:"truncated"`
	Content       string                   `xml:"content,omitempty"`
	WaveHighlight *sourceanalysis.Analysis `xml:"wave-highlight,omitempty"`
}

type Ref struct {
	Name      string `xml:"name"`
	OID       string `xml:"oid"`
	UpdatedAt string `xml:"updated-at,omitempty"`
}

type Refs struct {
	Branches []Ref `xml:"branches>branch"`
	Tags     []Ref `xml:"tags>tag"`
}
