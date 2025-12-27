package dto

type PaginationMeta struct {
	TotalRecords int64 `json:"totalRecords"`
	CurrentPage  int32 `json:"currentPage"`
	PerPage      int32 `json:"perPage"`
}
