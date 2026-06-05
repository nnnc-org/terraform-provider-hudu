package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// toGoValue converts a framework attr.Value into a JSON-friendly Go value
// suitable for marshalling into a Hudu custom_fields request body.
func toGoValue(ctx context.Context, v attr.Value) (any, error) {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	switch x := v.(type) {
	case types.String:
		return x.ValueString(), nil
	case types.Bool:
		return x.ValueBool(), nil
	case types.Int64:
		return x.ValueInt64(), nil
	case types.Int32:
		return x.ValueInt32(), nil
	case types.Float64:
		return x.ValueFloat64(), nil
	case types.Float32:
		return x.ValueFloat32(), nil
	case types.Number:
		f := x.ValueBigFloat()
		if f == nil {
			return nil, nil
		}
		if f.IsInt() {
			i, _ := f.Int64()
			return i, nil
		}
		out, _ := f.Float64()
		return out, nil
	case types.List:
		elems := x.Elements()
		out := make([]any, len(elems))
		for i, e := range elems {
			val, err := toGoValue(ctx, e)
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil
	case types.Set:
		elems := x.Elements()
		out := make([]any, len(elems))
		for i, e := range elems {
			val, err := toGoValue(ctx, e)
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil
	case types.Tuple:
		elems := x.Elements()
		out := make([]any, len(elems))
		for i, e := range elems {
			val, err := toGoValue(ctx, e)
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil
	case types.Map:
		out := map[string]any{}
		for k, e := range x.Elements() {
			val, err := toGoValue(ctx, e)
			if err != nil {
				return nil, err
			}
			out[k] = val
		}
		return out, nil
	case types.Object:
		out := map[string]any{}
		for k, e := range x.Attributes() {
			val, err := toGoValue(ctx, e)
			if err != nil {
				return nil, err
			}
			out[k] = val
		}
		return out, nil
	case types.Dynamic:
		return toGoValue(ctx, x.UnderlyingValue())
	default:
		return nil, fmt.Errorf("unsupported attr.Value type %T", v)
	}
}

// toAttrValue converts a JSON-decoded Go value into an attr.Value plus its
// attr.Type, recursively. Used when reading custom field values from the API.
func toAttrValue(v any) (attr.Value, attr.Type) {
	switch x := v.(type) {
	case nil:
		return types.DynamicNull(), types.DynamicType
	case bool:
		return types.BoolValue(x), types.BoolType
	case string:
		return types.StringValue(x), types.StringType
	case json.Number:
		if i, err := x.Int64(); err == nil {
			f := new(big.Float).SetInt64(i)
			return types.NumberValue(f), types.NumberType
		}
		if f, err := x.Float64(); err == nil {
			bf := big.NewFloat(f)
			return types.NumberValue(bf), types.NumberType
		}
		return types.StringValue(x.String()), types.StringType
	case float64:
		bf := big.NewFloat(x)
		return types.NumberValue(bf), types.NumberType
	case int64:
		bf := new(big.Float).SetInt64(x)
		return types.NumberValue(bf), types.NumberType
	case []any:
		elems := make([]attr.Value, len(x))
		etypes := make([]attr.Type, len(x))
		for i, e := range x {
			ev, et := toAttrValue(e)
			elems[i] = ev
			etypes[i] = et
		}
		tup, _ := types.TupleValue(etypes, elems)
		return tup, types.TupleType{ElemTypes: etypes}
	case map[string]any:
		attrs := make(map[string]attr.Value, len(x))
		atypes := make(map[string]attr.Type, len(x))
		for k, e := range x {
			ev, et := toAttrValue(e)
			attrs[k] = ev
			atypes[k] = et
		}
		obj, _ := types.ObjectValue(atypes, attrs)
		return obj, types.ObjectType{AttrTypes: atypes}
	default:
		// Fallback: stringify.
		return types.StringValue(fmt.Sprintf("%v", x)), types.StringType
	}
}

// snakeCase replicates Hudu's label-to-key normalization: lowercase, with
// runs of whitespace collapsed to a single underscore.
func snakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevUnderscore := false
	for _, r := range s {
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '-':
			if !prevUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				prevUnderscore = true
			}
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r - 'A' + 'a')
			prevUnderscore = false
		default:
			b.WriteRune(r)
			prevUnderscore = false
		}
	}
	return strings.Trim(b.String(), "_")
}
