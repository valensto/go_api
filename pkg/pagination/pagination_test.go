package pagination_test

import (
	"testing"

	"github.com/valensto/api_apbp/pkg/pagination"
)

func TestParseQuery(t *testing.T) {
	var tests = []struct {
		in       string
		expected pagination.Query
	}{
		{"/v1/orders?limit=20&page=1", pagination.Query{
			Limit: 20,
			Skip:  0,
		}},
		{"/v1/orders", pagination.Query{
			Limit: 10,
			Skip:  0,
		}},
		{"/v1/orders?limit=40&page=5", pagination.Query{
			Limit: 40,
			Skip:  160,
		}},
		{"/v1/orders?limit=0&page=0", pagination.Query{
			Limit: 1,
			Skip:  0,
		}},
		{"/v1/orders?limit=-1&page=-1", pagination.Query{
			Limit: 1,
			Skip:  0,
		}},
	}

	for _, tt := range tests {
		p := pagination.ParseQuery(tt.in)

		if p.Limit != tt.expected.Limit {
			t.Errorf("ParseQuery failed to limit on %v, expected: %v, got: %v", tt.in, tt.expected.Limit, p.Limit)
		}

		if p.Skip != tt.expected.Skip {
			t.Errorf("ParseQuery failed to skip  on %v, expected: %v, got: %v", tt.in, tt.expected.Skip, p.Skip)
		}
	}

}

func TestNewMeta(t *testing.T) {
	m := make(map[string]interface{})
	m["perPages"] = 10
	m["totalElements"] = 2
	m["totalPages"] = 1

	var tests = []struct {
		in       map[string]interface{}
		expected pagination.Meta
	}{
		{m, pagination.Meta{
			PerPages:      10,
			TotalElements: 2,
			TotalPages:    1,
		}},
	}

	for _, tt := range tests {
		meta, err := pagination.NewMeta(tt.in)
		if err != nil {
			t.Errorf("NewMeta failed on init")
		}
		if tt.expected.PerPages != meta.PerPages {
			t.Errorf("NewMeta failed on PerPages, expected: %v, got: %v", tt.expected.PerPages, meta.PerPages)
		}
		if tt.expected.TotalElements != meta.TotalElements {
			t.Errorf("NewMeta failed on TotalElements, expected: %v, got: %v", tt.expected.TotalElements, meta.TotalElements)
		}
		if tt.expected.TotalPages != meta.TotalPages {
			t.Errorf("NewMeta failed TotalPages, expected: %v, got: %v", tt.expected.TotalPages, meta.TotalPages)
		}
	}
}

func TestGetLinks(t *testing.T) {
	var tests = []struct {
		query    pagination.Query
		base     string
		total    int
		expected map[string]string
	}{
		{pagination.Query{
			Limit: 10,
			Skip:  10,
		}, "http://localhost/users", 40, map[string]string{
			"self":  "http://localhost/users?limit=10&page=2",
			"first": "http://localhost/users?limit=10&page=1",
			"prev":  "http://localhost/users?limit=10&page=1",
			"next":  "http://localhost/users?limit=10&page=3",
			"last":  "http://localhost/users?limit=10&page=4",
		}},
		{pagination.Query{
			Limit: 10,
			Skip:  0,
		}, "http://localhost/users", 40, map[string]string{
			"self":  "http://localhost/users?limit=10&page=1",
			"first": "http://localhost/users?limit=10&page=1",
			"next":  "http://localhost/users?limit=10&page=2",
			"last":  "http://localhost/users?limit=10&page=4",
		}},
		{pagination.Query{
			Limit: 10,
			Skip:  30,
		}, "http://localhost/users", 40, map[string]string{
			"self":  "http://localhost/users?limit=10&page=4",
			"first": "http://localhost/users?limit=10&page=1",
			"prev":  "http://localhost/users?limit=10&page=3",
			"last":  "http://localhost/users?limit=10&page=4",
		}},
	}

	for _, tt := range tests {
		links := tt.query.GetLinks(tt.base, tt.total)

		if links["self"] != tt.expected["self"] {
			t.Errorf("GetLinks failed to self, expected: %v, got: %v", links["self"], tt.expected["self"])
		}

		if links["first"] != tt.expected["first"] {
			t.Errorf("GetLinks failed to first, expected: %v, got: %v", links["first"], tt.expected["first"])
		}

		if _, ok := tt.expected["prev"]; ok {
			if links["prev"] != tt.expected["prev"] {
				t.Errorf("GetLinks failed to prev, expected: %v, got: %v", links["prev"], tt.expected["prev"])
			}
		}

		if _, ok := tt.expected["next"]; ok {
			if links["next"] != tt.expected["next"] {
				t.Errorf("GetLinks failed to next, expected: %v, got: %v", links["next"], tt.expected["next"])
			}
		}

		if links["last"] != tt.expected["last"] {
			t.Errorf("GetLinks failed to last, expected: %v, got: %v", links["last"], tt.expected["last"])
		}
	}
}
