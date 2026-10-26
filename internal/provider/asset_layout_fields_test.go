package provider

import (
	"reflect"
	"testing"
)

func TestMatchLayoutFields(t *testing.T) {
	type tc struct {
		name        string
		plan, state []fieldKey
		wantIDs     []int64
		wantDestroy []int64
	}
	cases := []tc{
		{
			name:    "rename in place reuses id at same position",
			plan:    []fieldKey{{"Connectivity Profile", 1, 0}, {"Serial", 2, 0}},
			state:   []fieldKey{{"Model", 1, 10}, {"Serial", 2, 11}},
			wantIDs: []int64{10, 11},
		},
		{
			name:        "removed field is destroyed",
			plan:        []fieldKey{{"A", 1, 0}},
			state:       []fieldKey{{"A", 1, 10}, {"B", 2, 11}},
			wantIDs:     []int64{10},
			wantDestroy: []int64{11},
		},
		{
			name:    "new appended field gets no id",
			plan:    []fieldKey{{"A", 1, 0}, {"B", 2, 0}},
			state:   []fieldKey{{"A", 1, 10}},
			wantIDs: []int64{10, 0},
		},
		{
			name:    "reorder matches by label, not position",
			plan:    []fieldKey{{"B", 1, 0}, {"A", 2, 0}},
			state:   []fieldKey{{"A", 1, 10}, {"B", 2, 11}},
			wantIDs: []int64{11, 10},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ids, destroy := matchLayoutFields(c.plan, c.state)
			if !reflect.DeepEqual(ids, c.wantIDs) {
				t.Errorf("planIDs = %v, want %v", ids, c.wantIDs)
			}
			if !reflect.DeepEqual(destroy, c.wantDestroy) {
				t.Errorf("destroy = %v, want %v", destroy, c.wantDestroy)
			}
		})
	}
}
