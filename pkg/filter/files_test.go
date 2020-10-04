package filter_test

import (
	"reflect"
	"testing"

	"github.com/valensto/api_apbp/pkg/filter"
)

func TestParseQuery(t *testing.T) {
	sort := make(map[string]string)
	sort["created"] = "desc"
	var tests = []struct {
		in       string
		expected filter.Query
	}{
		{"/v1/orders?populate=1", filter.Query{
			Populate: true,
		}},
	}

	for _, tt := range tests {
		f := filter.ParseQuery(tt.in)
		if f.Populate != tt.expected.Populate {
			t.Errorf("ParseQuery failed to populate, expected: %v, got: %v", tt.expected.Populate, f.Populate)
		}
		if !reflect.DeepEqual(&f.Sort, &tt.expected.Sort) {
			t.Errorf("ParseQuery failed to sort, expected: %v, got: %v", tt.expected.Sort, f.Sort)
		}
	}
}
