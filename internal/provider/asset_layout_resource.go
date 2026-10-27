package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

var (
	_ resource.Resource                = (*assetLayoutResource)(nil)
	_ resource.ResourceWithConfigure   = (*assetLayoutResource)(nil)
	_ resource.ResourceWithImportState = (*assetLayoutResource)(nil)
)

// NewAssetLayoutResource returns a new hudu_asset_layout resource implementation.
func NewAssetLayoutResource() resource.Resource {
	return &assetLayoutResource{}
}

type assetLayoutResource struct {
	client *client.Client
}

type assetLayoutModel struct {
	ID               types.String            `tfsdk:"id"`
	Name             types.String            `tfsdk:"name"`
	Icon             types.String            `tfsdk:"icon"`
	Color            types.String            `tfsdk:"color"`
	IconColor        types.String            `tfsdk:"icon_color"`
	SidebarFolderID  types.Int64             `tfsdk:"sidebar_folder_id"`
	Active           types.Bool              `tfsdk:"active"`
	IncludePasswords types.Bool              `tfsdk:"include_passwords"`
	IncludePhotos    types.Bool              `tfsdk:"include_photos"`
	IncludeComments  types.Bool              `tfsdk:"include_comments"`
	IncludeFiles     types.Bool              `tfsdk:"include_files"`
	PasswordTypes    types.String            `tfsdk:"password_types"`
	Slug             types.String            `tfsdk:"slug"`
	CreatedAt        types.String            `tfsdk:"created_at"`
	UpdatedAt        types.String            `tfsdk:"updated_at"`
	Fields           []assetLayoutFieldModel `tfsdk:"fields"`
}

type assetLayoutFieldModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Label           types.String `tfsdk:"label"`
	FieldType       types.String `tfsdk:"field_type"`
	Position        types.Int64  `tfsdk:"position"`
	Required        types.Bool   `tfsdk:"required"`
	ShowInList      types.Bool   `tfsdk:"show_in_list"`
	Hint            types.String `tfsdk:"hint"`
	Min             types.Int64  `tfsdk:"min"`
	Max             types.Int64  `tfsdk:"max"`
	Options         types.String `tfsdk:"options"`
	Expiration      types.Bool   `tfsdk:"expiration"`
	LinkableID      types.Int64  `tfsdk:"linkable_id"`
	ListID          types.Int64  `tfsdk:"list_id"`
	MultipleOptions types.Bool   `tfsdk:"multiple_options"`
}

func (r *assetLayoutResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset_layout"
}

func (r *assetLayoutResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hudu Asset Layout. Note: the Hudu API has no DELETE endpoint for layouts; `terraform destroy` deactivates the layout (sets `active = false`) and removes it from state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Numeric ID assigned by Hudu, as a string.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":       schema.StringAttribute{Required: true, Description: "Asset layout name."},
			"icon":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Description: "Font Awesome icon class, e.g. `fas fa-laptop`."},
			"color":      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Description: "Background color as a hex string."},
			"icon_color": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Description: "Icon color as a hex string."},
			"sidebar_folder_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the sidebar folder this layout is grouped under in the Hudu sidebar. The folder must already exist in Hudu — the API has no endpoint to create sidebar folders, so create the folder in the Hudu UI and reference its numeric ID here.",
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the layout is active. Setting to false hides it from the UI but does not delete it.",
			},
			"include_passwords": schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether assets using this layout include a passwords section."},
			"include_photos":    schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether assets using this layout include a photos section."},
			"include_comments":  schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether assets using this layout include a comments section."},
			"include_files":     schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether assets using this layout include a files section."},
			"password_types":    schema.StringAttribute{Optional: true, Description: "Newline-separated list of password types."},
			"slug": schema.StringAttribute{
				Description:   "URL slug generated by Hudu.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Description:   "Timestamp the layout was created, in ISO 8601 UTC (e.g. `2026-01-15T12:34:56.789Z`).",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{Computed: true, Description: "Timestamp the layout was last updated, in ISO 8601 UTC."},
			"fields": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom fields. Order in HCL is preserved; matching against the API is by list position so reordering will be treated as updates rather than re-creates.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description:   "Numeric field ID assigned by Hudu.",
							Computed:      true,
							PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						},
						"label":            schema.StringAttribute{Required: true, Description: "Field label shown in the Hudu UI. Also determines the snake_case key used for this field in `hudu_asset.custom_fields`."},
						"field_type":       schema.StringAttribute{Required: true, Description: "One of: Text, RichText, Heading, CheckBox, Website, Password, Email, Phone, Number, Date, Dropdown, Embed, Address, AssetTag, ListSelect, etc."},
						"position":         schema.Int64Attribute{Required: true, Description: "1-based position of the field within the layout."},
						"required":         schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the field must be filled in."},
						"show_in_list":     schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the field is shown in asset list views."},
						"hint":             schema.StringAttribute{Optional: true, Description: "Help text shown beneath the field in the Hudu UI."},
						"min":              schema.Int64Attribute{Optional: true, Description: "Minimum accepted value (Number fields)."},
						"max":              schema.Int64Attribute{Optional: true, Description: "Maximum accepted value (Number fields)."},
						"options":          schema.StringAttribute{Optional: true, Description: "Newline-separated list of dropdown options."},
						"expiration":       schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether values in this field feed Hudu expiration tracking (Date fields)."},
						"linkable_id":      schema.Int64Attribute{Optional: true, Description: "Asset layout ID to pull values from (AssetTag fields)."},
						"list_id":          schema.Int64Attribute{Optional: true, Description: "List ID to pull values from (ListSelect fields)."},
						"multiple_options": schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether more than one option may be selected (Dropdown and ListSelect fields)."},
					},
				},
			},
		},
	}
}

func (r *assetLayoutResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *assetLayoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assetLayoutModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := assetLayoutFromModel(&plan)
	out, err := r.client.CreateAssetLayout(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Create asset layout failed", err.Error())
		return
	}
	assetLayoutScalarsToModel(out, &plan)
	plan.Fields = reconcileLayoutFields(plan.Fields, out.Fields)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assetLayoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assetLayoutModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	out, err := r.client.GetAssetLayout(ctx, id)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read asset layout failed", err.Error())
		return
	}
	assetLayoutScalarsToModel(out, &state)
	state.Fields = reconcileLayoutFields(state.Fields, out.Fields)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assetLayoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state assetLayoutModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in plan", err.Error())
		return
	}
	in := assetLayoutFromModel(&plan)

	// Reconcile planned fields against existing ones so each targets the right
	// server id: a field is matched by label, or (for a rename) by position when
	// its old label is going away. Existing fields that nothing reuses are
	// deleted via is_destroyed (Hudu has no field DELETE endpoint, and omitting a
	// field on PUT leaves it in place).
	planKeys := fieldKeys(plan.Fields)
	stateKeys := fieldKeys(state.Fields)
	planIDs, destroyIDs := matchLayoutFields(planKeys, stateKeys)
	for i := range in.Fields {
		if in.Fields[i].ID == 0 && planIDs[i] != 0 {
			in.Fields[i].ID = planIDs[i]
		}
	}
	stateByID := map[int64]assetLayoutFieldModel{}
	for _, f := range state.Fields {
		stateByID[f.ID.ValueInt64()] = f
	}
	for _, did := range destroyIDs {
		f := stateByID[did]
		destroyed := true
		in.Fields = append(in.Fields, client.AssetLayoutField{
			ID:          did,
			Label:       f.Label.ValueString(),
			FieldType:   f.FieldType.ValueString(),
			IsDestroyed: &destroyed,
		})
	}

	out, err := r.client.UpdateAssetLayout(ctx, id, in)
	if err != nil {
		resp.Diagnostics.AddError("Update asset layout failed", err.Error())
		return
	}
	assetLayoutScalarsToModel(out, &plan)
	plan.Fields = reconcileLayoutFields(plan.Fields, out.Fields)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assetLayoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assetLayoutModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	// The API has no DELETE endpoint for layouts, so destroy soft-deletes
	// by PUTing active=false.
	if err := r.client.DeactivateAssetLayout(ctx, id); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Deactivate asset layout failed", err.Error())
	}
}

func (r *assetLayoutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func assetLayoutFromModel(m *assetLayoutModel) *client.AssetLayout {
	out := &client.AssetLayout{
		Name:             m.Name.ValueString(),
		Icon:             m.Icon.ValueString(),
		Color:            m.Color.ValueString(),
		IconColor:        m.IconColor.ValueString(),
		SidebarFolderID:  int64Ptr(m.SidebarFolderID),
		Active:           boolPtr(m.Active),
		IncludePasswords: boolPtr(m.IncludePasswords),
		IncludePhotos:    boolPtr(m.IncludePhotos),
		IncludeComments:  boolPtr(m.IncludeComments),
		IncludeFiles:     boolPtr(m.IncludeFiles),
		PasswordTypes:    stringPtr(m.PasswordTypes),
	}
	if len(m.Fields) > 0 {
		out.Fields = make([]client.AssetLayoutField, len(m.Fields))
		for i, f := range m.Fields {
			out.Fields[i] = client.AssetLayoutField{
				ID:              f.ID.ValueInt64(),
				Label:           f.Label.ValueString(),
				FieldType:       f.FieldType.ValueString(),
				Position:        f.Position.ValueInt64(),
				Required:        boolPtr(f.Required),
				ShowInList:      boolPtr(f.ShowInList),
				Hint:            f.Hint.ValueString(),
				Min:             int64Ptr(f.Min),
				Max:             int64Ptr(f.Max),
				Options:         f.Options.ValueString(),
				Expiration:      boolPtr(f.Expiration),
				LinkableID:      int64Ptr(f.LinkableID),
				ListID:          int64Ptr(f.ListID),
				MultipleOptions: boolPtr(f.MultipleOptions),
			}
		}
	}
	return out
}

// assetLayoutScalarsToModel copies the layout's scalar attributes from an API
// response. It deliberately leaves m.Fields alone — field reconciliation is
// handled by reconcileLayoutFields so the configured order/positions stay
// authoritative and cosmetic API reordering/renumbering does not churn.
func assetLayoutScalarsToModel(a *client.AssetLayout, m *assetLayoutModel) {
	m.ID = types.StringValue(fmt.Sprintf("%d", a.ID))
	m.Name = types.StringValue(a.Name)
	m.Icon = types.StringValue(a.Icon)
	m.Color = types.StringValue(a.Color)
	m.IconColor = types.StringValue(a.IconColor)
	m.SidebarFolderID = int64FromPtr(a.SidebarFolderID)
	m.Active = boolFromPtrDefault(a.Active, true)
	m.IncludePasswords = boolFromPtrDefault(a.IncludePasswords, false)
	m.IncludePhotos = boolFromPtrDefault(a.IncludePhotos, false)
	m.IncludeComments = boolFromPtrDefault(a.IncludeComments, false)
	m.IncludeFiles = boolFromPtrDefault(a.IncludeFiles, false)
	m.PasswordTypes = stringFromPtr(a.PasswordTypes)
	m.Slug = types.StringValue(a.Slug)
	m.CreatedAt = types.StringValue(a.CreatedAt)
	m.UpdatedAt = types.StringValue(a.UpdatedAt)
}

// reconcileLayoutFields keeps the model's field order and configured values,
// matching each field to its API counterpart by label to stamp the server id
// and fill in computed booleans the configuration left unset. API-only fields
// are ignored (the configuration owns the field set).
func reconcileLayoutFields(model []assetLayoutFieldModel, api []client.AssetLayoutField) []assetLayoutFieldModel {
	byLabel := make(map[string]client.AssetLayoutField, len(api))
	for _, f := range api {
		byLabel[f.Label] = f
	}
	out := make([]assetLayoutFieldModel, len(model))
	for i, mf := range model {
		rf := mf
		if af, ok := byLabel[mf.Label.ValueString()]; ok {
			rf.ID = types.Int64Value(af.ID)
			rf.Required = boolFillUnknown(mf.Required, af.Required, false)
			rf.ShowInList = boolFillUnknown(mf.ShowInList, af.ShowInList, false)
			rf.Expiration = boolFillUnknown(mf.Expiration, af.Expiration, false)
			rf.MultipleOptions = boolFillUnknown(mf.MultipleOptions, af.MultipleOptions, false)
		} else {
			rf.ID = int64FillUnknown(mf.ID, 0)
			rf.Required = boolFillUnknown(mf.Required, nil, false)
			rf.ShowInList = boolFillUnknown(mf.ShowInList, nil, false)
			rf.Expiration = boolFillUnknown(mf.Expiration, nil, false)
			rf.MultipleOptions = boolFillUnknown(mf.MultipleOptions, nil, false)
		}
		out[i] = rf
	}
	return out
}

// fieldKey is the minimal identity of a layout field used for reconciliation.
type fieldKey struct {
	label string
	pos   int64
	id    int64
}

func fieldKeys(fs []assetLayoutFieldModel) []fieldKey {
	out := make([]fieldKey, len(fs))
	for i, f := range fs {
		out[i] = fieldKey{label: f.Label.ValueString(), pos: f.Position.ValueInt64(), id: f.ID.ValueInt64()}
	}
	return out
}

// matchLayoutFields maps each planned field to the existing field id it should
// reuse (0 = brand new) and returns the ids of existing fields that nothing
// reuses (to be destroyed). A field is matched by label first; a planned field
// whose label is new is treated as a rename of the existing field at the same
// position when that field's label is itself going away. This keeps renames
// in place (same id, data preserved) instead of delete+recreate.
func matchLayoutFields(plan, state []fieldKey) (planIDs []int64, destroy []int64) {
	stateByLabel := map[string]fieldKey{}
	stateByPos := map[int64]fieldKey{}
	for _, s := range state {
		stateByLabel[s.label] = s
		stateByPos[s.pos] = s
	}
	planLabels := map[string]bool{}
	for _, p := range plan {
		planLabels[p.label] = true
	}
	planIDs = make([]int64, len(plan))
	used := map[int64]bool{}
	for i, p := range plan {
		if s, ok := stateByLabel[p.label]; ok {
			planIDs[i] = s.id
			used[s.id] = true
			continue
		}
		if s, ok := stateByPos[p.pos]; ok && !planLabels[s.label] && !used[s.id] {
			planIDs[i] = s.id // rename in place
			used[s.id] = true
		}
	}
	for _, s := range state {
		if !used[s.id] {
			destroy = append(destroy, s.id)
		}
	}
	return planIDs, destroy
}

func boolFillUnknown(v types.Bool, apiPtr *bool, def bool) types.Bool {
	if !v.IsUnknown() && !v.IsNull() {
		return v
	}
	if apiPtr != nil {
		return types.BoolValue(*apiPtr)
	}
	return types.BoolValue(def)
}

func int64FillUnknown(v types.Int64, api int64) types.Int64 {
	if !v.IsUnknown() && !v.IsNull() {
		return v
	}
	return types.Int64Value(api)
}

func boolFromPtrDefault(p *bool, def bool) types.Bool {
	if p == nil {
		return types.BoolValue(def)
	}
	return types.BoolValue(*p)
}
