package github

type Type string

const (
	TypeFile Type = "file"
	TypeDir  Type = "dir"
)

type ContentItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Sha  string `json:"sha"`
	Type Type   `json:"type"`
}
