package vo

type ApisTree struct {
	Id          string     `json:"id"`
	Label       string     `json:"label"`
	Leaf        bool       `json:"leaf"`
	HasChildren bool       `json:"hasChildren"`
	Children    []ApisTree `json:"children,omitempty"`
}
