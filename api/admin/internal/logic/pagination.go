package logic

// normalizePage 统一页码默认值，避免列表接口返回 0 页。
func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// normalizePageSize 统一分页大小。
// 最大限制为 100 条，避免后台列表接口一次查询过多数据。
func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
