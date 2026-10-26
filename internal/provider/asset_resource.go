package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

var (
	_ resource.Resource                = (*assetResource)(nil)
	_ resource.ResourceWithConfigure   = (*assetResource)(nil)
	_ resource.ResourceWithImportState = (*assetResource)(nil)
)

// NewAssetResource returns a new hudu_asset resource implementation.
func NewAssetResource() resource.Resource {
	return &assetResource{}
}

type assetResource struct {
	client *client.Client
}

type assetModel struct {
	ID                  types.String  `tfsdk:"id"`
	CompanyID           types.Int64   `tfsdk:"company_id"`
	AssetLayoutID       types.Int64   `tfsdk:"asset_layout_id"`
	Name                types.String  `tfsdk:"name"`
	PrimarySerial       types.String  `tfsdk:"primary_serial"`
	PrimaryMail         types.String  `tfsdk:"primary_mail"`
	PrimaryModel        types.String  `tfsdk:"primary_model"`
	PrimaryManufacturer types.String  `tfsdk:"primary_manufacturer"`
	CustomFields        types.Dynamic `tfsdk:"custom_fields"`
	Slug                types.String  `tfsdk:"slug"`
	URL                 types.String  `tfsdk:"url"`
	CreatedAt           types.String  `tfsdk:"created_at"`
	UpdatedAt           types.String  `tfsdk:"updated_at"`
	AssetType           types.String  `tfsdk:"asset_type"`
}

func (r *assetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (r *assetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hudu asset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"company_id": schema.Int64Attribute{
				Required:      true,
				Description:   "Owning company ID. Changing this forces a new asset.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"asset_layout_id": schema.Int64Attribute{
				Required:      true,
				Description:   "Layout ID. Changing this forces a new asset.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name":                 schema.StringAttribute{Required: true},
			"primary_serial":       schema.StringAttribute{Optional: true},
			"primary_mail":         schema.StringAttribute{Optional: true},
			"primary_model":        schema.StringAttribute{Optional: true},
			"primary_manufacturer": schema.StringAttribute{Optional: true},
			"custom_fields": schema.DynamicAttribute{
				Optional: true,
				Computed: true,
				Description: "Custom fields for the asset, as an object literal keyed by the layout field's label converted " +
					"to snake_case (lowercase, spaces become underscores). Values may be strings, numbers, booleans, lists, " +
					"or nested objects (for Address fields). Example:\n\n" +
					"```hcl\ncustom_fields = {\n  operating_system = \"Linux\"\n  score            = 5\n  is_active        = true\n  categories       = [\"Item 1\", \"Item 2\"]\n  office_location  = {\n    address_line_1 = \"123 Main St\"\n    city           = \"Denver\"\n    state          = \"CO\"\n    zip            = \"80202\"\n    country_name   = \"USA\"\n  }\n}\n```",
			},
			"slug": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{Computed: true},
			"asset_type": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *assetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", "Expected *client.Client.")
		return
	}
	r.client = c
}

func (r *assetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cf, diag := customFieldsToWire(ctx, plan.CustomFields)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &client.AssetCreateRequest{
		Name:                plan.Name.ValueString(),
		AssetLayoutID:       plan.AssetLayoutID.ValueInt64(),
		PrimarySerial:       plan.PrimarySerial.ValueString(),
		PrimaryMail:         plan.PrimaryMail.ValueString(),
		PrimaryModel:        plan.PrimaryModel.ValueString(),
		PrimaryManufacturer: plan.PrimaryManufacturer.ValueString(),
		CustomFields:        cf,
	}
	out, err := r.client.CreateAsset(ctx, plan.CompanyID.ValueInt64(), in)
	if err != nil {
		resp.Diagnostics.AddError("Create asset failed", err.Error())
		return
	}
	// Keep the configured values authoritative; only adopt server-owned
	// computed attributes. This avoids "inconsistent result after apply" when
	// Hudu normalizes values (dates, asset links, empty strings) on write.
	assetComputedToModel(out, &plan)
	if plan.CustomFields.IsUnknown() {
		plan.CustomFields = customFieldsObject(out)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	out, err := r.client.GetAsset(ctx, state.CompanyID.ValueInt64(), id)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read asset failed", err.Error())
		return
	}
	assetReadToModel(ctx, out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan assetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in plan", err.Error())
		return
	}

	cf, diag := customFieldsToWire(ctx, plan.CustomFields)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	ps := plan.PrimarySerial.ValueString()
	pm := plan.PrimaryMail.ValueString()
	pmod := plan.PrimaryModel.ValueString()
	pmf := plan.PrimaryManufacturer.ValueString()

	in := &client.AssetUpdateRequest{
		Asset: client.AssetUpdateBody{
			Name:                &name,
			PrimarySerial:       &ps,
			PrimaryMail:         &pm,
			PrimaryModel:        &pmod,
			PrimaryManufacturer: &pmf,
			CustomFields:        cf,
		},
	}
	out, err := r.client.UpdateAsset(ctx, plan.CompanyID.ValueInt64(), id, in)
	if err != nil {
		resp.Diagnostics.AddError("Update asset failed", err.Error())
		return
	}
	assetComputedToModel(out, &plan)
	if plan.CustomFields.IsUnknown() {
		plan.CustomFields = customFieldsObject(out)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	if err := r.client.DeleteAsset(ctx, state.CompanyID.ValueInt64(), id); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Delete asset failed", err.Error())
	}
}

func (r *assetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <company_id>:<asset_id>
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<company_id>:<asset_id>`.")
		return
	}
	companyID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid company ID", err.Error())
		return
	}
	if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		resp.Diagnostics.AddError("Invalid asset ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("company_id"), companyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// customFieldsToWire walks the dynamic plan value and produces the Hudu API's
// expected `[ {key: value}, ... ]` shape.
func customFieldsToWire(ctx context.Context, d types.Dynamic) ([]map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if d.IsNull() || d.IsUnknown() {
		return nil, diags
	}
	under := d.UnderlyingValue()
	if under == nil {
		return nil, diags
	}

	var pairs []map[string]any
	switch v := under.(type) {
	case types.Object:
		for k, av := range v.Attributes() {
			g, err := toGoValue(ctx, av)
			if err != nil {
				diags.AddError("Invalid custom_fields value", err.Error())
				return nil, diags
			}
			pairs = append(pairs, map[string]any{snakeCase(k): g})
		}
	case types.Map:
		for k, av := range v.Elements() {
			g, err := toGoValue(ctx, av)
			if err != nil {
				diags.AddError("Invalid custom_fields value", err.Error())
				return nil, diags
			}
			pairs = append(pairs, map[string]any{snakeCase(k): g})
		}
	default:
		diags.AddError("Invalid custom_fields type",
			fmt.Sprintf("custom_fields must be an object or map, got %T", under))
		return nil, diags
	}
	return pairs, diags
}

// assetComputedToModel copies only the server-owned, computed attributes from
// an API response onto the model, leaving configuration-controlled attributes
// (name, primary_*, custom_fields) untouched so they stay authoritative.
func assetComputedToModel(a *client.Asset, m *assetModel) {
	m.ID = types.StringValue(fmt.Sprintf("%d", a.ID))
	m.Slug = types.StringValue(a.Slug)
	m.URL = types.StringValue(a.URL)
	m.CreatedAt = types.StringValue(a.CreatedAt)
	m.UpdatedAt = types.StringValue(a.UpdatedAt)
	m.AssetType = types.StringValue(a.AssetType)
}

// customFieldsObject builds a dynamic object value from the API response's
// `fields` array (keys are the snake_case label).
func customFieldsObject(a *client.Asset) types.Dynamic {
	if len(a.Fields) == 0 {
		return types.DynamicNull()
	}
	attrs := make(map[string]attr.Value, len(a.Fields))
	atypes := make(map[string]attr.Type, len(a.Fields))
	for _, f := range a.Fields {
		if len(f.Value) == 0 {
			continue
		}
		var raw any
		dec := json.NewDecoder(bytes.NewReader(f.Value))
		dec.UseNumber()
		if err := dec.Decode(&raw); err != nil {
			raw = string(f.Value)
		}
		val, typ := toAttrValue(raw)
		k := snakeCase(f.Label)
		attrs[k] = val
		atypes[k] = typ
	}
	if len(attrs) == 0 {
		return types.DynamicNull()
	}
	obj, _ := types.ObjectValue(atypes, attrs)
	return types.DynamicValue(obj)
}

// reconcileCustomFields merges the API's custom fields onto the prior value,
// keeping the prior value for any key that differs only by Hudu's write-side
// normalization (dates, asset links, empty-vs-null) so plans stay stable.
func reconcileCustomFields(ctx context.Context, prior types.Dynamic, a *client.Asset) types.Dynamic {
	api := customFieldsObject(a)
	if prior.IsNull() || prior.IsUnknown() {
		return api
	}
	priorObj, ok := prior.UnderlyingValue().(types.Object)
	if !ok {
		return api
	}
	priorAttrs := priorObj.Attributes()
	priorTypes := priorObj.AttributeTypes(ctx)

	apiAttrs := map[string]attr.Value{}
	apiTypes := map[string]attr.Type{}
	if !api.IsNull() {
		if apiObj, ok := api.UnderlyingValue().(types.Object); ok {
			apiAttrs = apiObj.Attributes()
			apiTypes = apiObj.AttributeTypes(ctx)
		}
	}

	attrs := map[string]attr.Value{}
	atypes := map[string]attr.Type{}
	for k, pv := range priorAttrs {
		av, present := apiAttrs[k]
		if !present {
			// Hudu dropped the key (e.g. an empty value) — keep the prior entry
			// so the object shape stays stable across refreshes.
			attrs[k] = pv
			atypes[k] = priorTypes[k]
			continue
		}
		if semanticEqualAttr(pv, av) {
			attrs[k] = pv
			atypes[k] = priorTypes[k]
		} else {
			attrs[k] = av
			atypes[k] = apiTypes[k]
		}
	}
	for k, av := range apiAttrs {
		if _, seen := attrs[k]; !seen {
			attrs[k] = av
			atypes[k] = apiTypes[k]
		}
	}
	if len(attrs) == 0 {
		return types.DynamicNull()
	}
	obj, _ := types.ObjectValue(atypes, attrs)
	return types.DynamicValue(obj)
}

// assetReadToModel refreshes state from a GET, preserving config-shaped values
// that are only cosmetically different from what Hudu returns.
func assetReadToModel(ctx context.Context, a *client.Asset, m *assetModel) {
	assetComputedToModel(a, m)
	m.CompanyID = types.Int64Value(a.CompanyID)
	m.AssetLayoutID = types.Int64Value(a.AssetLayoutID)
	m.Name = types.StringValue(a.Name)
	m.PrimarySerial = preserveString(m.PrimarySerial, a.PrimarySerial)
	m.PrimaryMail = preserveString(m.PrimaryMail, a.PrimaryMail)
	m.PrimaryModel = preserveString(m.PrimaryModel, a.PrimaryModel)
	m.PrimaryManufacturer = preserveString(m.PrimaryManufacturer, a.PrimaryManufacturer)
	m.CustomFields = reconcileCustomFields(ctx, m.CustomFields, a)
}

// preserveString keeps the prior value when it is empty-equivalent to the API
// value (Hudu stores unset top-level strings as null), else adopts the API value.
func preserveString(prior types.String, api string) types.String {
	priorS := ""
	if !prior.IsNull() && !prior.IsUnknown() {
		priorS = prior.ValueString()
	}
	if emptyEqual(priorS, api) {
		return prior
	}
	return stringOrNull(api)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
