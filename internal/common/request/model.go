package request

import "searchengine/internal/config"

// filters request
type OneSelectFilterReq struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BoolSelectFilterReq struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type FilterRequest struct {
	Category    string                     `json:"category"`
	Range       []config.RangeFilter       `json:"range"`
	MultiSelect []config.MultiSelectFilter `json:"multi-select"`
	OneSelect   []OneSelectFilterReq       `json:"one-select"`
	BoolSelect  []BoolSelectFilterReq      `json:"bool-select"`
}
