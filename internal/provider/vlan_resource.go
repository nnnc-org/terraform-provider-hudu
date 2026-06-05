package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

var (
	_ resource.Resource                = (*vlanResource)(nil)
	_ resource.ResourceWithConfigure   = (*vlanResource)(nil)
	_ resource.ResourceWithImportState = (*vlanResource)(nil)
)

// NewVlanResource returns a new hudu_vlan resource implementation.
func NewVlanResource() resource.Resource {
	return &vlanResource{}
}

type vlanResource struct {
	client *client.Client
}

type vlanModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	VlanID           types.Int64  `tfsdk:"vlan_id"`
	CompanyID        types.Int64  `tfsdk:"company_id"`
	Description      types.String `tfsdk:"description"`
	Notes            types.String `tfsdk:"notes"`
	VlanZoneID       types.Int64  `tfsdk:"vlan_zone_id"`
	StatusListItemID types.Int64  `tfsdk:"status_list_item_id"`
	RoleListItemID   types.Int64  `tfsdk:"role_list_item_id"`
	Archived         types.Bool   `tfsdk:"archived"`
	Slug             types.String `tfsdk:"slug"`
	URL              types.String `tfsdk:"url"`
	NetworksCount    types.Int64  `tfsdk:"networks_count"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *vlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (r *vlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hudu VLAN.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{Required: true},
			"vlan_id": schema.Int64Attribute{
				Required:    true,
				Description: "Numeric VLAN ID (1-4094).",
				Validators:  []validator.Int64{int64validator.Between(1, 4094)},
			},
			"company_id":          schema.Int64Attribute{Required: true},
			"description":         schema.StringAttribute{Optional: true},
			"notes":               schema.StringAttribute{Optional: true},
			"vlan_zone_id":        schema.Int64Attribute{Optional: true, Computed: true},
			"status_list_item_id": schema.Int64Attribute{Optional: true, Computed: true},
			"role_list_item_id":   schema.Int64Attribute{Optional: true, Computed: true},
			"archived": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"slug": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"url": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"networks_count": schema.Int64Attribute{Computed: true},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *vlanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.CreateVlan(ctx, vlanFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create VLAN failed", err.Error())
		return
	}
	vlanToModel(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	out, err := r.client.GetVlan(ctx, id)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VLAN failed", err.Error())
		return
	}
	vlanToModel(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in plan", err.Error())
		return
	}
	out, err := r.client.UpdateVlan(ctx, id, vlanFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update VLAN failed", err.Error())
		return
	}
	vlanToModel(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	if err := r.client.DeleteVlan(ctx, id); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Delete VLAN failed", err.Error())
	}
}

func (r *vlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func vlanFromModel(m *vlanModel) *client.Vlan {
	return &client.Vlan{
		Name:             m.Name.ValueString(),
		VlanID:           m.VlanID.ValueInt64(),
		CompanyID:        m.CompanyID.ValueInt64(),
		Description:      stringPtr(m.Description),
		Notes:            stringPtr(m.Notes),
		VlanZoneID:       int64Ptr(m.VlanZoneID),
		StatusListItemID: int64Ptr(m.StatusListItemID),
		RoleListItemID:   int64Ptr(m.RoleListItemID),
		Archived:         boolPtr(m.Archived),
	}
}

func vlanToModel(v *client.Vlan, m *vlanModel) {
	m.ID = types.StringValue(fmt.Sprintf("%d", v.ID))
	m.Name = types.StringValue(v.Name)
	m.VlanID = types.Int64Value(v.VlanID)
	m.CompanyID = types.Int64Value(v.CompanyID)
	m.Description = stringFromPtr(v.Description)
	m.Notes = stringFromPtr(v.Notes)
	m.VlanZoneID = int64FromPtr(v.VlanZoneID)
	m.StatusListItemID = int64FromPtr(v.StatusListItemID)
	m.RoleListItemID = int64FromPtr(v.RoleListItemID)
	m.Archived = types.BoolValue(v.ArchivedAt != nil && *v.ArchivedAt != "")
	m.Slug = types.StringValue(v.Slug)
	m.URL = types.StringValue(v.URL)
	m.NetworksCount = types.Int64Value(v.NetworksCount)
	m.CreatedAt = types.StringValue(v.CreatedAt)
	m.UpdatedAt = types.StringValue(v.UpdatedAt)
}
