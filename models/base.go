package models

type PageReq[T any] struct {
	PageNum  int64 `form:"pageNum" binding:"required,gte=1"`
	PageSize int64 `form:"pageSize" binding:"required,gte=1"`
	Params   T     `form:"params"`
}

func (pageReq PageReq[T]) Skip() int64 {
	if pageReq.PageNum < 1 {
		return 0
	}
	return (pageReq.PageNum - 1) * pageReq.PageSize
}

func (pageReq PageReq[T]) Take() int64 {
	return pageReq.PageSize
}

type PageResp[T any] struct {
	Total   int64 `json:"total"`
	Pages   int64 `json:"pages"`
	PageNum int64 `json:"pageNum"`
	List    []T   `json:"list"`
}

type BaseTime struct {
	/**
	 * 创建时间
	 */
	CreateTime int64 `json:"createTime,string"`

	/**
	 * 最后修改时间
	 */
	ModifiedTime int64 `json:"modifiedTime,string"`
}
