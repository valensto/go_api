package filter

import (
	"net/url"
	"strconv"
	"time"

	"github.com/valensto/api_apbp/pkg/pagination"
)

type Range struct {
	Start time.Time
	End   time.Time
}

type Query struct {
	Populate   bool
	Filters    *map[string]string
	Sort       *map[string]string
	Range      *Range
	Term       string
	Pagination pagination.Query
}

func (q *Query) parseRange(query url.Values) {
	r := Range{}
	r.Start = time.Now()
	r.End = time.Time{}

	if startStr := query.Get("start"); startStr != "" {
		layout := "2006-01-02T15:04:05.000Z"
		if time, err := time.Parse(layout, startStr); err == nil {
			r.Start = time
		}
	}

	if endStr := query.Get("end"); endStr != "" {
		layout := "2006-01-02T15:04:05.000Z"
		if time, err := time.Parse(layout, endStr); err == nil {
			r.End = time
		}
	}

	q.Range = &r
}

func ParseQuery(queryStr string) Query {
	f := &Query{}

	u, err := url.Parse(queryStr)
	if err != nil {
		return *f
	}

	query := u.Query()

	f.Pagination = pagination.ParseQuery(queryStr)

	if populateStr := query.Get("populate"); populateStr != "" {
		if populate, err := strconv.ParseBool(populateStr); err == nil {
			f.Populate = populate
		}
	}

	if termStr := query.Get("term"); termStr != "" {
		f.Term = termStr
	}

	f.parseRange(query)

	// if sorts, present := query["sort"]; present || len(sorts) > 0 {
	// 	m := make(map[string]string)

	// 	jsonSort, err := json.Marshal(sorts)
	// 	if err != nil {
	// 		fmt.Printf("error occured during marshalling, got=%w", err)
	// 	}
	// 	fmt.Println(jsonSort)
	// 	err = json.Unmarshal(jsonSort, &m)
	// 	if err != nil {
	// 		fmt.Printf("error occured during unmarshalling, got=%w", err)
	// 	}
	// 	fmt.Println(m)
	// 	f.Sort = &m
	// }

	return *f
}
