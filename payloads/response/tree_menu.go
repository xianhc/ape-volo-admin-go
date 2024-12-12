package response

type TreeMenuVo struct {
	TreeMenuMeta TreeMenuMeta `json:"meta"`
	Name         string       `json:"name"`
	Path         string       `json:"path"`
	Hidden       bool         `json:"hidden"`
	Redirect     string       `json:"redirect,omitempty"`
	Component    string       `json:"component"`
	AlwaysShow   bool         `json:"alwaysShow"`
	Children     []TreeMenuVo `json:"children,omitempty"`
}

type TreeMenuMeta struct {
	Title   string `json:"title"`
	Icon    string `json:"icon"`
	NoCache bool   `json:"noCache"`
}
