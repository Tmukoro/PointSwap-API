package models


type PaginationMeta struct {
	Limit       int   `json:"limit"`
	Offset      int   `json:"offset"`
	Total       int64 `json:"total"`
	Has_more    bool  `json:"has_more"`
	Next_offset *int  `json:"next_offset"`
}

type InfiniteScrollData struct {
	Items any            `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}
