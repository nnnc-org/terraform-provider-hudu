package provider

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

func TestReconcileKeepsAssetLink(t *testing.T) {
	// prior state: connected_firewall configured as a list with one id.
	cfTuple, _ := types.TupleValue(
		[]attr.Type{types.NumberType},
		[]attr.Value{types.NumberValue(big.NewFloat(6))},
	)
	priorObj, d := types.ObjectValue(
		map[string]attr.Type{
			"connected_firewall": types.TupleType{ElemTypes: []attr.Type{types.NumberType}},
			"serial":             types.StringType,
		},
		map[string]attr.Value{
			"connected_firewall": cfTuple,
			"serial":             types.StringValue("S123"),
		},
	)
	if d.HasError() {
		t.Fatalf("build prior: %v", d)
	}
	prior := types.DynamicValue(priorObj)

	// API echo: Hudu returns the link as a JSON *string* (double-encoded), which
	// is the real wire shape observed from GET /assets.
	api := &client.Asset{Fields: []client.AssetField{
		{Label: "Connected Firewall", Value: json.RawMessage(`"[{\"id\":6,\"url\":\"/a/x\",\"name\":\"FW\"}]"`)},
		{Label: "Serial", Value: json.RawMessage(`"S123"`)},
	}}

	// Direct equality check of the two per-field values.
	apiCF := customFieldsObject(api)
	apiObj := apiCF.UnderlyingValue().(types.Object)
	av := apiObj.Attributes()["connected_firewall"]
	t.Logf("prior canon = %#v", canonGo(attrToGo(cfTuple)))
	t.Logf("api   canon = %#v", canonGo(attrToGo(av)))
	if !semanticEqualAttr(cfTuple, av) {
		t.Errorf("semanticEqualAttr(prior [6], api array-object) = false, want true")
	}

	// End-to-end: reconcile must keep the prior [6] value.
	got := reconcileCustomFields(context.Background(), prior, api)
	gotObj := got.UnderlyingValue().(types.Object)
	gotCF := gotObj.Attributes()["connected_firewall"]
	if _, isTuple := gotCF.(types.Tuple); !isTuple {
		t.Errorf("reconciled connected_firewall = %T (%v), want the prior Tuple [6]", gotCF, gotCF)
	}
}
