package web

// calculatePagination calculates pagination information.
func calculatePagination(totalBookmarks, currentPage, itemsPerPage int) PaginationInfo {
	totalPages := (totalBookmarks + itemsPerPage - 1) / itemsPerPage

	// Adjust currentPage if it's out of bounds
	if currentPage > totalPages && totalPages > 0 {
		currentPage = totalPages
	} else if totalPages == 0 {
		currentPage = 1
	}

	startIndex := min(max((currentPage-1)*itemsPerPage, 0), totalBookmarks)
	endIndex := min(startIndex+itemsPerPage, totalBookmarks)

	return PaginationInfo{
		CurrentPage:    currentPage,
		TotalPages:     totalPages,
		ItemsPerPage:   itemsPerPage,
		TotalBookmarks: totalBookmarks,
		StartIndex:     startIndex,
		EndIndex:       endIndex,
	}
}

func filterToggleURL(p *RequestParams, current, path string) string {
	if p.FilterBy == current {
		return p.with().Filter("").Build(path)
	}
	return p.with().Filter(current).Build(path)
}
