package provider

import (
	"context"
	"math"
	"reflect"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// Hudu normalizes several custom-field value shapes on write (dates to ISO8601,
// asset-link fields to a JSON array of {id,url,name} objects, empty strings to
// null). These helpers let Read decide whether a value the API echoed back is
// semantically the same as what the configuration/state holds, so cosmetic
// normalization does not surface as a perpetual diff.

var dateLayouts = []string{
	"2006/01/02",
	"2006-01-02",
	time.RFC3339,
	"2006-01-02T15:04:05.000Z07:00",
	"2006-01-02T15:04:05Z07:00",
}

// normDate parses a date-ish string and returns it as YYYY-MM-DD. ok is false
// when the string is not a recognized date (so callers leave it untouched).
func normDate(s string) (string, bool) {
	for _, l := range dateLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.Format("2006-01-02"), true
		}
	}
	return "", false
}

// linkID reports whether m is a single asset-link object ({"id":..,"url":..,
// "name":..}) and returns its id. Address and other object-valued fields have
// no "id"/"url", so they are not mistaken for links.
func linkID(m map[string]any) (int64, bool) {
	idv, ok := m["id"]
	if !ok {
		return 0, false
	}
	if _, hasURL := m["url"]; !hasURL {
		if _, hasName := m["name"]; !hasName {
			return 0, false
		}
	}
	switch n := idv.(type) {
	case int64:
		return n, true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}

// extractIDs reports whether every element of xs is an object carrying an "id"
// and, if so, returns those ids sorted. Hudu returns an asset-link (AssetTag)
// field as an array of such objects when more than one asset is linked:
// [{"id":6,"url":...,"name":...}].
func extractIDs(xs []any) ([]int64, bool) {
	if len(xs) == 0 {
		return nil, false
	}
	ids := make([]int64, 0, len(xs))
	for _, e := range xs {
		m, ok := e.(map[string]any)
		if !ok {
			return nil, false
		}
		idv, ok := m["id"]
		if !ok {
			return nil, false
		}
		switch n := idv.(type) {
		case int64:
			ids = append(ids, n)
		case float64:
			ids = append(ids, int64(n))
		default:
			return nil, false
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, true
}

// canonGo reduces a JSON-friendly Go value to a canonical form for comparison:
//   - "" and empty collections become nil (Hudu drops/º nulls them),
//   - date strings collapse to YYYY-MM-DD,
//   - whole-number floats become int64,
//   - an array of {id,...} objects (an asset link) becomes its sorted id list.
func canonGo(g any) any {
	switch x := g.(type) {
	case nil:
		return nil
	case string:
		if x == "" {
			return nil
		}
		if d, ok := normDate(x); ok {
			return d
		}
		return x
	case bool:
		return x
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		if x == math.Trunc(x) {
			return int64(x)
		}
		return x
	case []any:
		if len(x) == 0 {
			return nil
		}
		if ids, ok := extractIDs(x); ok {
			out := make([]any, len(ids))
			for i, id := range ids {
				out[i] = id
			}
			return out
		}
		out := make([]any, 0, len(x))
		for _, e := range x {
			out = append(out, canonGo(e))
		}
		return out
	case map[string]any:
		// A single asset link is returned as a bare {id,url,name} object; reduce
		// it to the same [id] shape as the multi-link array and the configured
		// list of ids so they compare equal.
		if id, ok := linkID(x); ok {
			return []any{id}
		}
		out := map[string]any{}
		for k, e := range x {
			if cv := canonGo(e); cv != nil {
				out[k] = cv
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return g
	}
}

// attrToGo converts a framework attr.Value to a plain Go value (best effort).
func attrToGo(v attr.Value) any {
	g, err := toGoValue(context.Background(), v)
	if err != nil {
		return nil
	}
	return g
}

// semanticEqualAttr reports whether two custom-field values are equivalent once
// Hudu's write-side normalization is accounted for.
func semanticEqualAttr(a, b attr.Value) bool {
	return reflect.DeepEqual(canonGo(attrToGo(a)), canonGo(attrToGo(b)))
}

// emptyEqual treats "" and null (the empty string) as equivalent, which is how
// Hudu stores unset top-level string attributes like primary_model.
func emptyEqual(a, b string) bool { return a == b }
