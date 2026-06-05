package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"Office Location":      "office_location",
		"Operating System":     "operating_system",
		"already_snake":        "already_snake",
		"PascalCaseExample":    "pascalcaseexample",
		"  leading whitespace": "leading_whitespace",
		"hyphen-separated":     "hyphen_separated",
	}
	for in, want := range cases {
		if got := snakeCase(in); got != want {
			t.Errorf("snakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToGoValue_PrimitivesAndContainers(t *testing.T) {
	ctx := context.Background()

	// string
	if got, err := toGoValue(ctx, types.StringValue("x")); err != nil || got != "x" {
		t.Errorf("string: got %v err %v", got, err)
	}
	// bool
	if got, err := toGoValue(ctx, types.BoolValue(true)); err != nil || got != true {
		t.Errorf("bool: got %v err %v", got, err)
	}
	// int64
	if got, err := toGoValue(ctx, types.Int64Value(42)); err != nil || got != int64(42) {
		t.Errorf("int64: got %v err %v", got, err)
	}
	// null
	if got, err := toGoValue(ctx, types.StringNull()); err != nil || got != nil {
		t.Errorf("null: got %v err %v", got, err)
	}

	// object with mixed types
	obj, _ := types.ObjectValue(
		map[string]attr.Type{
			"a": types.StringType,
			"b": types.Int64Type,
		},
		map[string]attr.Value{
			"a": types.StringValue("hello"),
			"b": types.Int64Value(7),
		},
	)
	got, err := toGoValue(ctx, obj)
	if err != nil {
		t.Fatalf("toGoValue(object): %v", err)
	}
	want := map[string]any{"a": "hello", "b": int64(7)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("object: got %v, want %v", got, want)
	}

	// list of strings
	lst, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("a"), types.StringValue("b"),
	})
	got, err = toGoValue(ctx, lst)
	if err != nil {
		t.Fatalf("toGoValue(list): %v", err)
	}
	if !reflect.DeepEqual(got, []any{"a", "b"}) {
		t.Errorf("list: got %v", got)
	}
}

func TestToAttrValue_Roundtrip(t *testing.T) {
	v, typ := toAttrValue("hello")
	if _, ok := v.(types.String); !ok {
		t.Errorf("string -> %T", v)
	}
	if typ != types.StringType {
		t.Errorf("string type -> %v", typ)
	}

	v, _ = toAttrValue(true)
	if _, ok := v.(types.Bool); !ok {
		t.Errorf("bool -> %T", v)
	}

	v, _ = toAttrValue(float64(5))
	if _, ok := v.(types.Number); !ok {
		t.Errorf("float -> %T", v)
	}

	v, _ = toAttrValue(map[string]any{"k": "v"})
	if _, ok := v.(types.Object); !ok {
		t.Errorf("map -> %T", v)
	}
}
