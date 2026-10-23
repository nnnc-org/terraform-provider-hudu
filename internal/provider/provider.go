package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

// Ensure HuduProvider satisfies various provider interfaces.
var _ provider.Provider = (*HuduProvider)(nil)

// HuduProvider is the Terraform provider for Hudu.
type HuduProvider struct {
	version string
}

// providerModel mirrors the provider block schema.
type providerModel struct {
	Host     types.String `tfsdk:"host"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// New returns a provider.Provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HuduProvider{version: version}
	}
}

func (p *HuduProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hudu"
	resp.Version = p.version
}

func (p *HuduProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Hudu IT documentation resources (companies, asset layouts, assets, networks, VLANs, VLAN zones) via Terraform/OpenTofu.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the Hudu instance, e.g. `https://hudu.example.com`. Falls back to the `HUDU_API_HOST` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API key (sent as the `x-api-key` header). Falls back to the `HUDU_API_KEY` environment variable.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification. Defaults to `false`.",
			},
		},
	}
}

func (p *HuduProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("HUDU_API_HOST")
	if !data.Host.IsNull() && !data.Host.IsUnknown() {
		host = data.Host.ValueString()
	}

	apiKey := os.Getenv("HUDU_API_KEY")
	if !data.APIKey.IsNull() && !data.APIKey.IsUnknown() {
		apiKey = data.APIKey.ValueString()
	}

	insecure := false
	if !data.Insecure.IsNull() && !data.Insecure.IsUnknown() {
		insecure = data.Insecure.ValueBool()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Hudu host",
			"Set the `host` provider attribute or the HUDU_API_HOST environment variable.",
		)
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Hudu API key",
			"Set the `api_key` provider attribute or the HUDU_API_KEY environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.New(client.Config{
		Host:      host,
		APIKey:    apiKey,
		Insecure:  insecure,
		UserAgent: "terraform-provider-hudu/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Hudu client", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *HuduProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCompanyResource,
		NewAssetLayoutResource,
		NewAssetResource,
		NewNetworkResource,
		NewVlanResource,
		NewVlanZoneResource,
	}
}

func (p *HuduProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
