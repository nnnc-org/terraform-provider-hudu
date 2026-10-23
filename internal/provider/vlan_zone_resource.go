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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

var (
	_ resource.Resource                = (*vlanZoneResource)(nil)
	_ resource.ResourceWithConfigure   = (*vlanZoneResource)(nil)
	_ resource.ResourceWithImportState = (*vlanZoneResource)(nil)
)

// NewVlanZoneResource returns a new hudu_vlan_zone resource implementation.
func NewVlanZoneResource() resource.Resource {
	return &vlanZoneResource{}
}

type vlanZoneResource struct {
	client *client.Client
}

type vlanZoneModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	CompanyID    types.Int64  `tfsdk:"company_id"`
	VlanIDRanges types.String `tfsdk:"vlan_id_ranges"`
	Description  types.String `tfsdk:"description"`
	Archived     types.Bool   `tfsdk:"archived"`
	Slug         types.String `tfsdk:"slug"`
	URL          types.String `tfsdk:"url"`
	VlansCount   types.Int64  `tfsdk:"vlans_count"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *vlanZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_zone"
}

func (r *vlanZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hudu VLAN Zone — a logical grouping of related VLANs (e.g. a datacenter, building, or rack range).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":       schema.StringAttribute{Required: true},
			"company_id": schema.Int64Attribute{Required: true},
			"vlan_id_ranges": schema.StringAttribute{
				Required:    true,
				Description: "Comma-separated list of numeric ranges (e.g. `100-500,1000-1500`). Each range must be inside 1-4094 and start <= end.",
			},
			"description": schema.StringAttribute{Optional: true},
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
			"vlans_count": schema.Int64Attribute{Computed: true},
			"created_at": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *vlanZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vlanZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.client.CreateVlanZone(ctx, vlanZoneFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create VLAN Zone failed", err.Error())
		return
	}
	vlanZoneToModel(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlanZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	out, err := r.client.GetVlanZone(ctx, id)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read VLAN Zone failed", err.Error())
		return
	}
	vlanZoneToModel(out, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlanZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlanZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in plan", err.Error())
		return
	}
	out, err := r.client.UpdateVlanZone(ctx, id, vlanZoneFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update VLAN Zone failed", err.Error())
		return
	}
	vlanZoneToModel(out, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlanZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID in state", err.Error())
		return
	}
	if err := r.client.DeleteVlanZone(ctx, id); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Delete VLAN Zone failed", err.Error())
	}
}

func (r *vlanZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func vlanZoneFromModel(m *vlanZoneModel) *client.VlanZone {
	return &client.VlanZone{
		Name:         m.Name.ValueString(),
		CompanyID:    m.CompanyID.ValueInt64(),
		VlanIDRanges: m.VlanIDRanges.ValueString(),
		Description:  stringPtr(m.Description),
		Archived:     boolPtr(m.Archived),
	}
}

func vlanZoneToModel(v *client.VlanZone, m *vlanZoneModel) {
	m.ID = types.StringValue(fmt.Sprintf("%d", v.ID))
	m.Name = types.StringValue(v.Name)
	m.CompanyID = types.Int64Value(v.CompanyID)
	m.VlanIDRanges = types.StringValue(v.VlanIDRanges)
	m.Description = stringFromPtr(v.Description)
	m.Archived = types.BoolValue(v.ArchivedAt != nil && *v.ArchivedAt != "")
	m.Slug = types.StringValue(v.Slug)
	m.URL = types.StringValue(v.URL)
	m.VlansCount = types.Int64Value(v.VlansCount)
	m.CreatedAt = types.StringValue(v.CreatedAt)
	m.UpdatedAt = types.StringValue(v.UpdatedAt)
}
