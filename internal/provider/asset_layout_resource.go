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
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":       schema.StringAttribute{Required: true},
			"icon":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Description: "Font Awesome icon class, e.g. `fas fa-laptop`."},
			"color":      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Description: "Background color as a hex string."},
			"icon_color": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Description: "Icon color as a hex string."},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the layout is active. Setting to false hides it from the UI but does not delete it.",
			},
			"include_passwords": schema.BoolAttribute{Optional: true, Computed: true},
			"include_photos":    schema.BoolAttribute{Optional: true, Computed: true},
			"include_comments":  schema.BoolAttribute{Optional: true, Computed: true},
			"include_files":     schema.BoolAttribute{Optional: true, Computed: true},
			"password_types":    schema.StringAttribute{Optional: true, Description: "Newline-separated list of password types."},
			"slug": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{Computed: true},
			"fields": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom fields. Order in HCL is preserved; matching against the API is by list position so reordering will be treated as updates rather than re-creates.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:      true,
							PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						},
						"label":            schema.StringAttribute{Required: true},
						"field_type":       schema.StringAttribute{Required: true, Description: "One of: Text, RichText, Heading, CheckBox, Website, Password, Email, Phone, Number, Date, Dropdown, Embed, Address, AssetTag, ListSelect, etc."},
						"position":         schema.Int64Attribute{Required: true},
						"required":         schema.BoolAttribute{Optional: true, Computed: true},
						"show_in_list":     schema.BoolAttribute{Optional: true, Computed: true},
						"hint":             schema.StringAttribute{Optional: true},
						"min":              schema.Int64Attribute{Optional: true},
						"max":              schema.Int64Attribute{Optional: true},
						"options":          schema.StringAttribute{Optional: true, Description: "Newline-separated list of dropdown options."},
						"expiration":       schema.BoolAttribute{Optional: true, Computed: true},
						"linkable_id":      schema.Int64Attribute{Optional: true, Description: "Asset layout ID to pull values from (AssetTag fields)."},
						"list_id":          schema.Int64Attribute{Optional: true, Description: "List ID to pull values from (ListSelect fields)."},
						"multiple_options": schema.BoolAttribute{Optional: true, Computed: true},
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
	assetLayoutToModel(out, &plan)
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
	assetLayoutToModel(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assetLayoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan assetLayoutModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in plan", err.Error())
		return
	}
	in := assetLayoutFromModel(&plan)
	out, err := r.client.UpdateAssetLayout(ctx, id, in)
	if err != nil {
		resp.Diagnostics.AddError("Update asset layout failed", err.Error())
		return
	}
	assetLayoutToModel(out, &plan)
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

func assetLayoutToModel(a *client.AssetLayout, m *assetLayoutModel) {
	m.ID = types.StringValue(fmt.Sprintf("%d", a.ID))
	m.Name = types.StringValue(a.Name)
	m.Icon = types.StringValue(a.Icon)
	m.Color = types.StringValue(a.Color)
	m.IconColor = types.StringValue(a.IconColor)
	m.Active = boolFromPtrDefault(a.Active, true)
	m.IncludePasswords = boolFromPtrDefault(a.IncludePasswords, false)
	m.IncludePhotos = boolFromPtrDefault(a.IncludePhotos, false)
	m.IncludeComments = boolFromPtrDefault(a.IncludeComments, false)
	m.IncludeFiles = boolFromPtrDefault(a.IncludeFiles, false)
	m.PasswordTypes = stringFromPtr(a.PasswordTypes)
	m.Slug = types.StringValue(a.Slug)
	m.CreatedAt = types.StringValue(a.CreatedAt)
	m.UpdatedAt = types.StringValue(a.UpdatedAt)

	m.Fields = make([]assetLayoutFieldModel, len(a.Fields))
	for i, f := range a.Fields {
		m.Fields[i] = assetLayoutFieldModel{
			ID:              types.Int64Value(f.ID),
			Label:           types.StringValue(f.Label),
			FieldType:       types.StringValue(f.FieldType),
			Position:        types.Int64Value(f.Position),
			Required:        boolFromPtrDefault(f.Required, false),
			ShowInList:      boolFromPtrDefault(f.ShowInList, false),
			Hint:            stringOrNull(f.Hint),
			Min:             int64FromPtr(f.Min),
			Max:             int64FromPtr(f.Max),
			Options:         stringOrNull(f.Options),
			Expiration:      boolFromPtrDefault(f.Expiration, false),
			LinkableID:      int64FromPtr(f.LinkableID),
			ListID:          int64FromPtr(f.ListID),
			MultipleOptions: boolFromPtrDefault(f.MultipleOptions, false),
		}
	}
}

func boolFromPtrDefault(p *bool, def bool) types.Bool {
	if p == nil {
		return types.BoolValue(def)
	}
	return types.BoolValue(*p)
}
