package provider

import (
	"context"
	"fmt"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// webhookEndpointUpdateParams wraps the input with the JSON key the Lago API requires.
// The lago-go-client Update method sends the input unwrapped (upstream bug), so we
// call the HTTP client directly for updates.
type webhookEndpointUpdateParams struct {
	WebhookEndpoint *lago.WebhookEndpointInput `json:"webhook_endpoint"`
}

var (
	_ resource.Resource                = &webhookEndpointResource{}
	_ resource.ResourceWithConfigure   = &webhookEndpointResource{}
	_ resource.ResourceWithImportState = &webhookEndpointResource{}
)

func NewWebhookEndpointResource() resource.Resource {
	return &webhookEndpointResource{}
}

type webhookEndpointResource struct {
	client *lago.Client
}

type webhookEndpointResourceModel struct {
	ID            types.String `tfsdk:"id"`
	LagoID        types.String `tfsdk:"lago_id"`
	WebhookURL    types.String `tfsdk:"webhook_url"`
	SignatureAlgo types.String `tfsdk:"signature_algo"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (r *webhookEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_endpoint"
}

func (r *webhookEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago webhook endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the webhook endpoint Lago UUID.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the webhook endpoint.",
			},
			"webhook_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL to which Lago will send webhook events. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"signature_algo": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Algorithm used to sign webhook payloads. Allowed values: `jwt`, `hmac`.",
				Validators: []validator.String{
					stringvalidator.OneOf("jwt", "hmac"),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Webhook endpoint creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *webhookEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*lagoProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *lagoProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = providerData.client
}

func (r *webhookEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookEndpointResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandWebhookEndpointInput(plan)

	created, lagoErr := r.client.WebhookEndpoint().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Webhook Endpoint", lagoErr.Error())
		return
	}

	state := mapWebhookEndpointToModel(created, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webhookEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookEndpointResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, lagoErr := r.client.WebhookEndpoint().Get(ctx, state.LagoID.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Webhook Endpoint", lagoErr.Error())
		return
	}

	newState := mapWebhookEndpointToModel(endpoint, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *webhookEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookEndpointResourceModel
	var state webhookEndpointResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandWebhookEndpointInput(plan)

	// The lago-go-client Update method sends the input without the required
	// "webhook_endpoint" JSON wrapper (upstream bug). Use the HTTP client directly.
	result := &lago.WebhookEndpointResult{}
	subPath := fmt.Sprintf("webhook_endpoints/%s", state.LagoID.ValueString())
	httpResp, httpErr := r.client.HttpClient.R().
		SetContext(ctx).
		SetError(&lago.Error{}).
		SetResult(result).
		SetBody(webhookEndpointUpdateParams{WebhookEndpoint: input}).
		Put(subPath)
	if httpErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Webhook Endpoint", httpErr.Error())
		return
	}
	if httpResp.IsError() {
		if lagoErr, ok := httpResp.Error().(*lago.Error); ok {
			resp.Diagnostics.AddError("Error Updating Lago Webhook Endpoint", lagoErr.Error())
		} else {
			resp.Diagnostics.AddError("Error Updating Lago Webhook Endpoint", httpResp.Status())
		}
		return
	}

	newState := mapWebhookEndpointToModel(result.WebhookEndpoint, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *webhookEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookEndpointResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, lagoErr := r.client.WebhookEndpoint().Delete(ctx, state.LagoID.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Webhook Endpoint", lagoErr.Error())
	}
}

func (r *webhookEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lago_id"), req.ID)...)
}

func expandWebhookEndpointInput(model webhookEndpointResourceModel) *lago.WebhookEndpointInput {
	input := &lago.WebhookEndpointInput{
		WebhookURL: model.WebhookURL.ValueString(),
	}

	if !model.SignatureAlgo.IsNull() && !model.SignatureAlgo.IsUnknown() {
		input.SignatureAlgo = lago.SignatureAlgo(model.SignatureAlgo.ValueString())
	}

	return input
}

func mapWebhookEndpointToModel(endpoint *lago.WebhookEndpoint, base webhookEndpointResourceModel) webhookEndpointResourceModel {
	state := base

	state.ID = types.StringValue(endpoint.LagoID.String())
	state.LagoID = types.StringValue(endpoint.LagoID.String())
	state.WebhookURL = types.StringValue(endpoint.WebhookURL)

	if endpoint.SignatureAlgo == "" {
		state.SignatureAlgo = types.StringNull()
	} else {
		state.SignatureAlgo = types.StringValue(string(endpoint.SignatureAlgo))
	}

	if endpoint.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(endpoint.CreatedAt.Format(time.RFC3339))
	}

	return state
}
