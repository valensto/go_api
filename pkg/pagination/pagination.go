package pagination

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
)

type Meta struct {
	PerPages      int `json:"perPages"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
}

func NewMeta(m map[string]interface{}) (Meta, error) {
	meta := Meta{}
	data, err := json.Marshal(m)
	if err != nil {
		return meta, err
	}

	err = json.Unmarshal(data, &meta)
	if err != nil {
		return meta, err
	}

	return meta, nil
}

// Query has pagination query parameters.
type Query struct {
	Limit int
	Skip  int
}

func (q *Query) init() {
	q.Limit = 10
	q.Skip = 0
}

func (q *Query) setMin() {
	if q.Limit <= 0 {
		q.Limit = 1
	}
	if q.Skip <= 0 {
		q.Skip = 0
	}

}

func (q Query) pageToSkip(p int) int {
	if p <= 0 {
		p = 1
	}
	return (p - 1) * q.Limit
}

func (q Query) skipToPage() int {
	return (q.Skip / q.Limit) + 1
}

func (q Query) GetLinks(base string, total int) map[string]string {
	links := make(map[string]string)
	if total <= 0 {
		return links
	}

	last := math.Ceil(float64(total) / float64(q.Limit))

	links["last"] = fmt.Sprintf("%v?limit=%v&page=%v", base, q.Limit, last)

	links["first"] = fmt.Sprintf("%v?limit=%v&page=%v", base, q.Limit, 1)
	links["self"] = fmt.Sprintf("%v?limit=%v&page=%v", base, q.Limit, q.skipToPage())

	if q.skipToPage() > 1 {
		links["prev"] = fmt.Sprintf("%v?limit=%v&page=%v", base, q.Limit, q.skipToPage()-1)
	}

	if q.skipToPage()+q.Limit < total {
		links["next"] = fmt.Sprintf("%v?limit=%v&page=%v", base, q.Limit, q.skipToPage()+1)
	}

	return links
}

// ParseQuery parses URL query string to get limit, page and sort
func ParseQuery(queryStr string) Query {
	// Set default values.
	p := &Query{}
	p.init()

	u, err := url.Parse(queryStr)
	if err != nil {
		return *p
	}
	query := u.Query()

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			p.Limit = limit
		}
	}

	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			p.Skip = p.pageToSkip(page)
		}
	}

	p.setMin()
	return *p
}
