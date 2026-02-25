package provider

import (
	"context"
	"fmt"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &walletResource{}
	_ resource.ResourceWithConfigure   = &walletResource{}
	_ resource.ResourceWithImportState = &walletResource{}
)

func NewWalletResource() resource.Resource {
	return &walletResource{}
}

type walletResource struct {
	client *lago.Client
}

type walletResourceModel struct {
	ID                               types.String `tfsdk:"id"`
	LagoID                           types.String `tfsdk:"lago_id"`
	ExternalCustomerID               types.String `tfsdk:"external_customer_id"`
	Name                             types.String `tfsdk:"name"`
	Currency                         types.String `tfsdk:"currency"`
	RateAmount                       types.String `tfsdk:"rate_amount"`
	PaidCredits                      types.String `tfsdk:"paid_credits"`
	GrantedCredits                   types.String `tfsdk:"granted_credits"`
	ExpirationAt                     types.String `tfsdk:"expiration_at"`
	InvoiceRequiresSuccessfulPayment types.Bool   `tfsdk:"invoice_requires_successful_payment"`
	Status                           types.String `tfsdk:"status"`
	CreditsBalance                   types.String `tfsdk:"credits_balance"`
	RecurringTransactionRules        types.List   `tfsdk:"recurring_transaction_rules"`
	CreatedAt                        types.String `tfsdk:"created_at"`
}

type walletRecurringTransactionRuleModel struct {
	LagoID                           types.String `tfsdk:"lago_id"`
	Interval                         types.String `tfsdk:"interval"`
	Method                           types.String `tfsdk:"method"`
	Trigger                          types.String `tfsdk:"trigger"`
	PaidCredits                      types.String `tfsdk:"paid_credits"`
	GrantedCredits                   types.String `tfsdk:"granted_credits"`
	ThresholdCredits                 types.String `tfsdk:"threshold_credits"`
	StartedAt                        types.String `tfsdk:"started_at"`
	ExpirationAt                     types.String `tfsdk:"expiration_at"`
	InvoiceRequiresSuccessfulPayment types.Bool   `tfsdk:"invoice_requires_successful_payment"`
}

func walletRecurringTransactionRuleObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"lago_id":                             types.StringType,
		"interval":                            types.StringType,
		"method":                              types.StringType,
		"trigger":                             types.StringType,
		"paid_credits":                        types.StringType,
		"granted_credits":                     types.StringType,
		"threshold_credits":                   types.StringType,
		"started_at":                          types.StringType,
		"expiration_at":                       types.StringType,
		"invoice_requires_successful_payment": types.BoolType,
	}
}

func (r *walletResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wallet"
}

func (r *walletResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Lago prepaid wallet for a customer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Terraform resource ID, set to the wallet Lago UUID.",
			},
			"lago_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Lago internal UUID for the wallet.",
			},
			"external_customer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "External identifier of the customer who owns this wallet. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Display name for the wallet.",
			},
			"currency": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Currency of the wallet (e.g. `USD`). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rate_amount": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Credit-to-currency conversion rate expressed as a decimal string (e.g. `\"1.0\"`).",
			},
			"paid_credits": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Number of credits purchased on create/top-up. Write-only — not returned by the API after creation.",
			},
			"granted_credits": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Number of free credits granted on create/top-up. Write-only — not returned by the API after creation.",
			},
			"expiration_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Wallet expiration date and time (RFC3339). When set, the wallet will expire at this date.",
			},
			"invoice_requires_successful_payment": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When `true`, invoices using wallet credits are only finalized after successful payment.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of the wallet (`active`, `terminated`).",
			},
			"credits_balance": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current credit balance of the wallet.",
			},
			"recurring_transaction_rules": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "List of rules that automatically top up the wallet on a schedule or threshold.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"lago_id": schema.StringAttribute{
							Computed:            true,
							Optional:            true,
							MarkdownDescription: "Lago UUID for this recurring transaction rule. Required when updating an existing rule.",
						},
						"interval": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Top-up interval. Allowed values: `weekly`, `monthly`, `quarterly`, `yearly`.",
							Validators: []validator.String{
								stringvalidator.OneOf("weekly", "monthly", "quarterly", "yearly"),
							},
						},
						"method": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Top-up method. Allowed values: `fixed`, `target`.",
							Validators: []validator.String{
								stringvalidator.OneOf("fixed", "target"),
							},
						},
						"trigger": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "What triggers the top-up. Allowed values: `interval`, `threshold`.",
							Validators: []validator.String{
								stringvalidator.OneOf("interval", "threshold"),
							},
						},
						"paid_credits": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Number of paid credits to add on each top-up.",
						},
						"granted_credits": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Number of free credits to grant on each top-up.",
						},
						"threshold_credits": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Credit balance threshold that triggers a top-up when the `trigger` is `threshold`.",
						},
						"started_at": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Start date and time for the recurring rule (RFC3339).",
						},
						"expiration_at": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Expiration date and time for the recurring rule (RFC3339).",
						},
						"invoice_requires_successful_payment": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "When `true`, invoices for this rule's top-up are only finalized after successful payment.",
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Wallet creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *walletResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *walletResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan walletResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandWalletInput(ctx, plan, nil)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, lagoErr := r.client.Wallet().Create(ctx, input)
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Creating Lago Wallet", lagoErr.Error())
		return
	}

	state, diags := mapWalletToModel(ctx, created, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *walletResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state walletResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wallet, lagoErr := r.client.Wallet().Get(ctx, state.LagoID.ValueString())
	if lagoErr != nil {
		if isNotFound(lagoErr) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Lago Wallet", lagoErr.Error())
		return
	}

	// Treat out-of-band termination as resource destruction.
	if wallet.Status == lago.Terminated {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := mapWalletToModel(ctx, wallet, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *walletResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan walletResourceModel
	var state walletResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := expandWalletInput(ctx, plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, lagoErr := r.client.Wallet().Update(ctx, input, state.LagoID.ValueString())
	if lagoErr != nil {
		resp.Diagnostics.AddError("Error Updating Lago Wallet", lagoErr.Error())
		return
	}

	newState, diags := mapWalletToModel(ctx, updated, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *walletResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state walletResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, lagoErr := r.client.Wallet().Delete(ctx, state.LagoID.ValueString())
	if lagoErr != nil && !isNotFound(lagoErr) {
		resp.Diagnostics.AddError("Error Deleting Lago Wallet", lagoErr.Error())
	}
}

func (r *walletResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lago_id"), req.ID)...)
}

// expandWalletInput converts the Terraform model into a WalletInput for the Lago API.
// When state is provided (i.e. on Update), existing recurring rule LagoIDs are
// threaded into the input so the API can match and update rules in-place.
func expandWalletInput(ctx context.Context, model walletResourceModel, state *walletResourceModel) (*lago.WalletInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &lago.WalletInput{
		ExternalCustomerID:               model.ExternalCustomerID.ValueString(),
		Currency:                         lago.Currency(model.Currency.ValueString()),
		RateAmount:                       model.RateAmount.ValueString(),
		InvoiceRequiresSuccessfulPayment: model.InvoiceRequiresSuccessfulPayment.ValueBool(),
	}

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		input.Name = model.Name.ValueString()
	}

	if !model.PaidCredits.IsNull() && !model.PaidCredits.IsUnknown() {
		input.PaidCredits = model.PaidCredits.ValueString()
	}

	if !model.GrantedCredits.IsNull() && !model.GrantedCredits.IsUnknown() {
		input.GrantedCredits = model.GrantedCredits.ValueString()
	}

	if !model.ExpirationAt.IsNull() && !model.ExpirationAt.IsUnknown() {
		t, err := time.Parse(time.RFC3339, model.ExpirationAt.ValueString())
		if err != nil {
			diags.AddError(
				"Invalid expiration_at",
				fmt.Sprintf("Must be a valid RFC3339 timestamp: %s", err),
			)
			return nil, diags
		}
		input.ExpirationAt = &t
	}

	rules, ruleDiags := expandRecurringTransactionRules(ctx, model.RecurringTransactionRules, state)
	diags.Append(ruleDiags...)
	if diags.HasError() {
		return nil, diags
	}
	input.RecurringTransactionRules = rules

	return input, diags
}

func expandRecurringTransactionRules(ctx context.Context, list types.List, state *walletResourceModel) ([]lago.RecurringTransactionRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return []lago.RecurringTransactionRuleInput{}, diags
	}

	var ruleModels []walletRecurringTransactionRuleModel
	diags.Append(list.ElementsAs(ctx, &ruleModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	// Build a map of existing state rules by lago_id so we can thread IDs into updates.
	existingLagoIDs := map[int]string{}
	if state != nil && !state.RecurringTransactionRules.IsNull() && !state.RecurringTransactionRules.IsUnknown() {
		var stateRules []walletRecurringTransactionRuleModel
		diags.Append(state.RecurringTransactionRules.ElementsAs(ctx, &stateRules, false)...)
		if diags.HasError() {
			return nil, diags
		}
		for i, sr := range stateRules {
			if !sr.LagoID.IsNull() && !sr.LagoID.IsUnknown() {
				existingLagoIDs[i] = sr.LagoID.ValueString()
			}
		}
	}

	rules := make([]lago.RecurringTransactionRuleInput, 0, len(ruleModels))
	for i, rm := range ruleModels {
		rule := lago.RecurringTransactionRuleInput{
			InvoiceRequiresSuccessfulPayment: rm.InvoiceRequiresSuccessfulPayment.ValueBool(),
		}

		// Populate LagoID from plan (if already computed) or from prior state.
		lagoIDStr := ""
		if !rm.LagoID.IsNull() && !rm.LagoID.IsUnknown() && rm.LagoID.ValueString() != "" {
			lagoIDStr = rm.LagoID.ValueString()
		} else if id, ok := existingLagoIDs[i]; ok {
			lagoIDStr = id
		}
		if lagoIDStr != "" {
			if parsed, err := uuid.Parse(lagoIDStr); err == nil {
				rule.LagoID = parsed
			}
		}

		if !rm.Interval.IsNull() && !rm.Interval.IsUnknown() {
			rule.Interval = rm.Interval.ValueString()
		}
		if !rm.Method.IsNull() && !rm.Method.IsUnknown() {
			rule.Method = rm.Method.ValueString()
		}
		if !rm.Trigger.IsNull() && !rm.Trigger.IsUnknown() {
			rule.Trigger = rm.Trigger.ValueString()
		}
		if !rm.PaidCredits.IsNull() && !rm.PaidCredits.IsUnknown() {
			rule.PaidCredits = rm.PaidCredits.ValueString()
		}
		if !rm.GrantedCredits.IsNull() && !rm.GrantedCredits.IsUnknown() {
			rule.GrantedCredits = rm.GrantedCredits.ValueString()
		}
		if !rm.ThresholdCredits.IsNull() && !rm.ThresholdCredits.IsUnknown() {
			rule.ThresholdCredits = rm.ThresholdCredits.ValueString()
		}

		if !rm.StartedAt.IsNull() && !rm.StartedAt.IsUnknown() {
			t, err := time.Parse(time.RFC3339, rm.StartedAt.ValueString())
			if err != nil {
				diags.AddError(
					"Invalid recurring_transaction_rules.started_at",
					fmt.Sprintf("Must be a valid RFC3339 timestamp: %s", err),
				)
				return nil, diags
			}
			rule.StartedAt = &t
		}

		if !rm.ExpirationAt.IsNull() && !rm.ExpirationAt.IsUnknown() {
			t, err := time.Parse(time.RFC3339, rm.ExpirationAt.ValueString())
			if err != nil {
				diags.AddError(
					"Invalid recurring_transaction_rules.expiration_at",
					fmt.Sprintf("Must be a valid RFC3339 timestamp: %s", err),
				)
				return nil, diags
			}
			rule.ExpirationAt = &t
		}

		rules = append(rules, rule)
	}

	return rules, diags
}

// mapWalletToModel converts a Lago API Wallet response into the Terraform state model.
// The base model is used to preserve write-only fields (paid_credits, granted_credits)
// that the API does not echo back.
func mapWalletToModel(ctx context.Context, wallet *lago.Wallet, base walletResourceModel) (walletResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	state := base

	state.ID = types.StringValue(wallet.LagoID.String())
	state.LagoID = types.StringValue(wallet.LagoID.String())
	state.ExternalCustomerID = types.StringValue(wallet.ExternalCustomerID)
	state.Currency = types.StringValue(string(wallet.Currency))
	state.RateAmount = types.StringValue(wallet.RateAmount)
	state.Status = types.StringValue(string(wallet.Status))
	state.CreditsBalance = types.StringValue(wallet.CreditsBalance)
	state.InvoiceRequiresSuccessfulPayment = types.BoolValue(wallet.InvoiceRequiresSuccessfulPayment)

	state.Name = stringOrNull(wallet.Name)

	// expiration_at: the Wallet struct uses time.Time (not *time.Time), zero means unset.
	if wallet.ExpirationAt.IsZero() {
		state.ExpirationAt = types.StringNull()
	} else {
		state.ExpirationAt = types.StringValue(wallet.ExpirationAt.Format(time.RFC3339))
	}

	if wallet.CreatedAt.IsZero() {
		state.CreatedAt = types.StringNull()
	} else {
		state.CreatedAt = types.StringValue(wallet.CreatedAt.Format(time.RFC3339))
	}

	// paid_credits and granted_credits are write-only — not returned by the API.
	// Preserve whatever was in the base (plan on create/update, prior state on read).
	// state.PaidCredits and state.GrantedCredits are already carried from base.

	rulesList, ruleDiags := flattenRecurringTransactionRules(ctx, wallet.RecurringTransactionRules, base.RecurringTransactionRules)
	diags.Append(ruleDiags...)
	if diags.HasError() {
		return state, diags
	}
	state.RecurringTransactionRules = rulesList

	return state, diags
}

func flattenRecurringTransactionRules(ctx context.Context, rules []lago.RecurringTransactionRuleResponse, base types.List) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}

	if len(rules) == 0 {
		// If the config had no rules (null), keep null so Terraform doesn't see null→[] drift.
		if base.IsNull() {
			return types.ListNull(objType), diags
		}
		return types.ListValueMust(objType, []attr.Value{}), diags
	}

	elems := make([]attr.Value, 0, len(rules))
	for _, rule := range rules {
		rm := walletRecurringTransactionRuleModel{
			LagoID:                           types.StringValue(rule.LagoID.String()),
			Interval:                         stringOrNull(rule.Interval),
			Method:                           stringOrNull(rule.Method),
			Trigger:                          stringOrNull(rule.Trigger),
			PaidCredits:                      stringOrNull(rule.PaidCredits),
			GrantedCredits:                   stringOrNull(rule.GrantedCredits),
			ThresholdCredits:                 stringOrNull(rule.ThresholdCredits),
			StartedAt:                        timeOrNull(rule.StartedAt),
			ExpirationAt:                     timeOrNull(rule.ExpirationAt),
			InvoiceRequiresSuccessfulPayment: types.BoolValue(rule.InvoiceRequiresSuccessfulPayment),
		}

		obj, objDiags := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), rm)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ListNull(objType), diags
		}

		elems = append(elems, obj)
	}

	list, listDiags := types.ListValue(objType, elems)
	diags.Append(listDiags...)

	return list, diags
}
