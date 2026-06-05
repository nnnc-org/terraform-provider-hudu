package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringPtr converts a Terraform string to a *string suitable for an API
// request body, returning nil if the value is null or unknown so the field
// is omitted from JSON.
func stringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// int64Ptr is the equivalent of stringPtr for int64s.
func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

// boolPtr is the equivalent of stringPtr for booleans.
func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// stringFromPtr converts a *string (possibly nil) to a Terraform string,
// using a null value when the pointer is nil.
func stringFromPtr(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// int64FromPtr converts a *int64 (possibly nil) to a Terraform int64.
func int64FromPtr(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}

// boolFromPtr converts a *bool (possibly nil) to a Terraform bool.
func boolFromPtr(p *bool) types.Bool {
	if p == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*p)
}
