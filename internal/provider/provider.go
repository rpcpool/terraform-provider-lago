package provider

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &lagoProvider{}

// New instantiates the Lago Terraform provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &lagoProvider{version: version}
	}
}

type lagoProvider struct {
	version string
}

type lagoProviderModel struct {
	APIEndpoint types.String `tfsdk:"api_endpoint"`
	APIKey      types.String `tfsdk:"api_key"`
}

type lagoProviderData struct {
	client *lago.Client
}

func (p *lagoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lago"
	resp.Version = p.version
}

func (p *lagoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lago API provider.",
		Attributes: map[string]schema.Attribute{
			"api_endpoint": schema.StringAttribute{
				MarkdownDescription: "Lago API endpoint (for example `https://api.getlago.com/api/v1`). Can also be set with `LAGO_API_ENDPOINT`.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Lago API key. Can also be set with `LAGO_API_KEY`.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *lagoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config lagoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiEndpoint := os.Getenv("LAGO_API_ENDPOINT")
	if !config.APIEndpoint.IsNull() && !config.APIEndpoint.IsUnknown() {
		apiEndpoint = config.APIEndpoint.ValueString()
	}
	apiEndpoint = strings.TrimSpace(apiEndpoint)

	apiKey := os.Getenv("LAGO_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}
	apiKey = strings.TrimSpace(apiKey)

	if apiEndpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_endpoint"),
			"Missing Lago API Endpoint",
			"Set `api_endpoint` in the provider configuration or set the `LAGO_API_ENDPOINT` environment variable.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Lago API Key",
			"Set `api_key` in the provider configuration or set the `LAGO_API_KEY` environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	lagoClient := lago.New().
		SetApiKey(apiKey).
		SetBaseURL(apiEndpoint)

	lagoClient.HttpClient.
		SetRetryCount(3).
		SetRetryWaitTime(200 * time.Millisecond).
		SetRetryMaxWaitTime(600 * time.Millisecond).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}
			sc := r.StatusCode()
			return sc == http.StatusTooManyRequests || sc >= 500
		})

	providerData := &lagoProviderData{client: lagoClient}
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *lagoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBillableMetricResource,
		NewPlanResource,
		NewAddOnResource,
		NewTaxResource,
		NewCustomerResource,
		NewWebhookEndpointResource,
		NewCouponResource,
		NewSubscriptionResource,
		NewWalletResource,
	}
}

func (p *lagoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
