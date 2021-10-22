package helper

import (
	"html/template"
	"strconv"
	"time"
)

func DisplayDate(dbDatestr string) string {
	dt, _ := time.Parse("2006-01-02 15:04:05", dbDatestr)
	return dt.Format("2006/1/02 15:04")
	// return dt.Format("01月02日2006年 15:04")
}

func PaginationHTML(link, urlParams string, pageNo, totalPages int) template.HTML {
	PAGETOSHOW := 5
	NOCLASS := ""
	ACTIVECLASS := "active"

	active := NOCLASS
	firstDisabled := NOCLASS
	lastDisabled := NOCLASS
	first := "?pageNo=1" + urlParams
	previous := first
	if pageNo-1 < 1 {
		previous = "?pageNo=" + strconv.Itoa(pageNo-1) + urlParams
		firstDisabled = " disabled"
	}
	last := "?pageNo=" + strconv.Itoa(totalPages) + urlParams
	next := "?pageNo=" + strconv.Itoa(pageNo+1) + urlParams
	if pageNo == totalPages {
		next = last
		lastDisabled = " disabled"
	}
	firstPage := 1
	lastPage := totalPages
	if totalPages > (PAGETOSHOW * 2) {
		if pageNo == firstPage || pageNo < PAGETOSHOW*2 {
			lastPage = firstPage + (PAGETOSHOW*2 - 1)
		} else if pageNo == lastPage || pageNo+PAGETOSHOW > totalPages {
			firstPage = lastPage - (PAGETOSHOW*2 - 1)
		} else if pageNo-PAGETOSHOW > 0 {
			firstPage = pageNo - (PAGETOSHOW - 1)
			lastPage = firstPage + (PAGETOSHOW*2 - 1)
		} else if pageNo+PAGETOSHOW < totalPages {
			lastPage = pageNo + (PAGETOSHOW - 1)
			firstPage = lastPage - (PAGETOSHOW*2 - 1)
		}
	}

	html := "<nav aria-label=\"浏览\"><ul class=\"pagination\">"
	// Previous
	html += "<li class=\"page-item" + firstDisabled + "\"><a class=\"page-link\" href=\"" + link + previous + "\" aria-label=\"上一页\"><span aria-hidden=\"true\">«</span></a></li>"
	//first
	html += "<li class=\"page-item" + firstDisabled + "\"><a class=\"page-link\" href=\"" + link + first + "\" aria-label=\"第一页\"><span aria-hidden=\"第一页\">第一页</span></a></li>"

	for i := firstPage; i <= lastPage; i++ {
		active = NOCLASS
		if i == pageNo {
			active = ACTIVECLASS
		}
		// pages
		html += "<li class=\"page-item " + active + "\"><a class=\"page-link\" href=\"" + link + "?pageNo=" + strconv.Itoa(i) + urlParams + "\">" + strconv.Itoa(i) + "</a></li>"
	}
	//last
	html += "<li class=\"page-item" + lastDisabled + "\"><a class=\"page-link\" href=\"" + link + last + "\" aria-label=\"最后一页\"><span aria-hidden=\"最后一页\">最后一页</span></a></li>"
	//next
	html += "<li class=\"page-item" + lastDisabled + "\"><a class=\"page-link\" href=\"" + link + next + "\" aria-label=\"下一页\"><span aria-hidden=\"下一页\">»</span></a></li>"
	html += "</ul></nav>"
	return template.HTML(html)
}
