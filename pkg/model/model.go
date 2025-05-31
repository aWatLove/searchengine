package model

type DocMsg struct {
	DocId    string                 `json:"doc_id"`
	Document map[string]interface{} `json:"doc"`
	Deleted  bool                   `json:"delete"`
}
