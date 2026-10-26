package provider

import (
	"reflect"
	"testing"
)

func TestNormDate(t *testing.T) {
	cases := map[string]struct {
		want string
		ok   bool
	}{
		"2028/01/09":               {"2028-01-09", true},
		"2028-01-09":               {"2028-01-09", true},
		"2028-01-09T00:00:00.000Z": {"2028-01-09", true},
		"2028-01-09T12:34:56Z":     {"2028-01-09", true},
		"not a date":               {"", false},
		"":                         {"", false},
	}
	for in, exp := range cases {
		got, ok := normDate(in)
		if got != exp.want || ok != exp.ok {
			t.Errorf("normDate(%q) = (%q,%v), want (%q,%v)", in, got, ok, exp.want, exp.ok)
		}
	}
}

func TestExtractIDs(t *testing.T) {
	ids, ok := extractIDs([]any{
		map[string]any{"id": int64(6), "url": "/a/x", "name": "FW"},
		map[string]any{"id": float64(2)},
	})
	if !ok || !reflect.DeepEqual(ids, []int64{2, 6}) {
		t.Errorf("extractIDs link array = (%v,%v), want ([2 6],true)", ids, ok)
	}
	if _, ok := extractIDs([]any{int64(1), int64(2)}); ok {
		t.Error("extractIDs on plain number list should be false")
	}
	if _, ok := extractIDs(nil); ok {
		t.Error("extractIDs(nil) should be false")
	}
}

func TestCanonGoEquivalences(t *testing.T) {
	eq := func(name string, a, b any) {
		if !reflect.DeepEqual(canonGo(a), canonGo(b)) {
			t.Errorf("%s: canonGo(%#v)=%#v != canonGo(%#v)=%#v", name, a, canonGo(a), b, canonGo(b))
		}
	}
	// dates: config "YYYY/MM/DD" vs Hudu's ISO echo
	eq("date", "2028/01/09", "2028-01-09T00:00:00.000Z")
	// asset link: config [6] vs Hudu's [{id,url,name}] echo
	eq("assetlink", []any{int64(6)}, []any{map[string]any{"id": int64(6), "url": "/a/x", "name": "FW"}})
	// empty string vs null
	eq("empty-null", "", nil)
	// empty list vs null (unset asset link)
	eq("emptylist-null", []any{}, nil)
	// whole float vs int
	eq("float-int", float64(6), int64(6))

	// genuine differences must NOT be equal
	if reflect.DeepEqual(canonGo("2028/01/09"), canonGo("2029/01/09")) {
		t.Error("different dates should not be equal")
	}
	if reflect.DeepEqual(canonGo([]any{int64(6)}), canonGo([]any{int64(7)})) {
		t.Error("different link ids should not be equal")
	}
	if reflect.DeepEqual(canonGo("hello"), canonGo("world")) {
		t.Error("different strings should not be equal")
	}
}
