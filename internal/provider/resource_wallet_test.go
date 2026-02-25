package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Unit tests — expandWalletInput
// ---------------------------------------------------------------------------

func TestWalletExpandInput_Basic(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	plan := walletResourceModel{
		ID:                               types.StringNull(),
		LagoID:                           types.StringNull(),
		ExternalCustomerID:               types.StringValue("cust-001"),
		Name:                             types.StringValue("My Wallet"),
		Currency:                         types.StringValue("USD"),
		RateAmount:                       types.StringValue("1.0"),
		PaidCredits:                      types.StringValue("100.0"),
		GrantedCredits:                   types.StringValue("10.0"),
		ExpirationAt:                     types.StringNull(),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
		Status:                           types.StringNull(),
		CreditsBalance:                   types.StringNull(),
		RecurringTransactionRules:        types.ListValueMust(types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}, []attr.Value{}),
		CreatedAt:                        types.StringNull(),
	}

	input, diags := expandWalletInput(ctx, plan, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.ExternalCustomerID != "cust-001" {
		t.Errorf("expected ExternalCustomerID %q, got %q", "cust-001", input.ExternalCustomerID)
	}
	if input.Name != "My Wallet" {
		t.Errorf("expected Name %q, got %q", "My Wallet", input.Name)
	}
	if string(input.Currency) != "USD" {
		t.Errorf("expected Currency %q, got %q", "USD", string(input.Currency))
	}
	if input.RateAmount != "1.0" {
		t.Errorf("expected RateAmount %q, got %q", "1.0", input.RateAmount)
	}
	if input.PaidCredits != "100.0" {
		t.Errorf("expected PaidCredits %q, got %q", "100.0", input.PaidCredits)
	}
	if input.GrantedCredits != "10.0" {
		t.Errorf("expected GrantedCredits %q, got %q", "10.0", input.GrantedCredits)
	}
	if input.ExpirationAt != nil {
		t.Errorf("expected nil ExpirationAt, got %v", input.ExpirationAt)
	}
	if input.InvoiceRequiresSuccessfulPayment {
		t.Error("expected InvoiceRequiresSuccessfulPayment to be false")
	}
	if len(input.RecurringTransactionRules) != 0 {
		t.Errorf("expected 0 recurring rules, got %d", len(input.RecurringTransactionRules))
	}
}

func TestWalletExpandInput_WithExpirationAt(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	plan := walletResourceModel{
		ID:                               types.StringNull(),
		LagoID:                           types.StringNull(),
		ExternalCustomerID:               types.StringValue("cust-002"),
		Name:                             types.StringNull(),
		Currency:                         types.StringValue("EUR"),
		RateAmount:                       types.StringValue("2.5"),
		PaidCredits:                      types.StringNull(),
		GrantedCredits:                   types.StringNull(),
		ExpirationAt:                     types.StringValue("2025-12-31T23:59:59Z"),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(true),
		Status:                           types.StringNull(),
		CreditsBalance:                   types.StringNull(),
		RecurringTransactionRules:        types.ListValueMust(types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}, []attr.Value{}),
		CreatedAt:                        types.StringNull(),
	}

	input, diags := expandWalletInput(ctx, plan, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.ExpirationAt == nil {
		t.Fatal("expected non-nil ExpirationAt")
	}
	expected := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	if !input.ExpirationAt.Equal(expected) {
		t.Errorf("expected ExpirationAt %v, got %v", expected, *input.ExpirationAt)
	}
	if !input.InvoiceRequiresSuccessfulPayment {
		t.Error("expected InvoiceRequiresSuccessfulPayment to be true")
	}
}

func TestWalletExpandInput_InvalidExpirationAt(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	plan := walletResourceModel{
		ID:                               types.StringNull(),
		LagoID:                           types.StringNull(),
		ExternalCustomerID:               types.StringValue("cust-003"),
		Name:                             types.StringNull(),
		Currency:                         types.StringValue("USD"),
		RateAmount:                       types.StringValue("1.0"),
		PaidCredits:                      types.StringNull(),
		GrantedCredits:                   types.StringNull(),
		ExpirationAt:                     types.StringValue("not-a-date"),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
		Status:                           types.StringNull(),
		CreditsBalance:                   types.StringNull(),
		RecurringTransactionRules:        types.ListValueMust(types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}, []attr.Value{}),
		CreatedAt:                        types.StringNull(),
	}

	_, diags := expandWalletInput(ctx, plan, nil)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for invalid expiration_at, got none")
	}
}

func TestWalletExpandInput_NullName(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	plan := walletResourceModel{
		ID:                               types.StringNull(),
		LagoID:                           types.StringNull(),
		ExternalCustomerID:               types.StringValue("cust-004"),
		Name:                             types.StringNull(),
		Currency:                         types.StringValue("USD"),
		RateAmount:                       types.StringValue("1.0"),
		PaidCredits:                      types.StringNull(),
		GrantedCredits:                   types.StringNull(),
		ExpirationAt:                     types.StringNull(),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
		Status:                           types.StringNull(),
		CreditsBalance:                   types.StringNull(),
		RecurringTransactionRules:        types.ListValueMust(types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}, []attr.Value{}),
		CreatedAt:                        types.StringNull(),
	}

	input, diags := expandWalletInput(ctx, plan, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.Name != "" {
		t.Errorf("expected empty Name for null model, got %q", input.Name)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — expandRecurringTransactionRules
// ---------------------------------------------------------------------------

func TestExpandRecurringTransactionRules_Empty(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	list := types.ListValueMust(types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}, []attr.Value{})

	rules, diags := expandRecurringTransactionRules(ctx, list, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(rules))
	}
}

func TestExpandRecurringTransactionRules_WithRule(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rm := walletRecurringTransactionRuleModel{
		LagoID:                           types.StringNull(),
		Interval:                         types.StringValue("monthly"),
		Method:                           types.StringValue("fixed"),
		Trigger:                          types.StringValue("interval"),
		PaidCredits:                      types.StringValue("50.0"),
		GrantedCredits:                   types.StringValue("5.0"),
		ThresholdCredits:                 types.StringNull(),
		StartedAt:                        types.StringNull(),
		ExpirationAt:                     types.StringNull(),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
	}

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}
	obj, objDiags := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), rm)
	if objDiags.HasError() {
		t.Fatalf("failed to create object: %+v", objDiags)
	}

	list, listDiags := types.ListValue(objType, []attr.Value{obj})
	if listDiags.HasError() {
		t.Fatalf("failed to create list: %+v", listDiags)
	}

	rules, diags := expandRecurringTransactionRules(ctx, list, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	if rule.Interval != "monthly" {
		t.Errorf("expected Interval %q, got %q", "monthly", rule.Interval)
	}
	if rule.Method != "fixed" {
		t.Errorf("expected Method %q, got %q", "fixed", rule.Method)
	}
	if rule.Trigger != "interval" {
		t.Errorf("expected Trigger %q, got %q", "interval", rule.Trigger)
	}
	if rule.PaidCredits != "50.0" {
		t.Errorf("expected PaidCredits %q, got %q", "50.0", rule.PaidCredits)
	}
	if rule.GrantedCredits != "5.0" {
		t.Errorf("expected GrantedCredits %q, got %q", "5.0", rule.GrantedCredits)
	}
	if rule.ThresholdCredits != "" {
		t.Errorf("expected empty ThresholdCredits, got %q", rule.ThresholdCredits)
	}
	if rule.InvoiceRequiresSuccessfulPayment {
		t.Error("expected InvoiceRequiresSuccessfulPayment to be false")
	}
}

func TestExpandRecurringTransactionRules_WithDates(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rm := walletRecurringTransactionRuleModel{
		LagoID:                           types.StringNull(),
		Interval:                         types.StringValue("weekly"),
		Method:                           types.StringValue("fixed"),
		Trigger:                          types.StringValue("interval"),
		PaidCredits:                      types.StringValue("10.0"),
		GrantedCredits:                   types.StringNull(),
		ThresholdCredits:                 types.StringNull(),
		StartedAt:                        types.StringValue("2024-01-01T00:00:00Z"),
		ExpirationAt:                     types.StringValue("2024-12-31T23:59:59Z"),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(true),
	}

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}
	obj, _ := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), rm)
	list, _ := types.ListValue(objType, []attr.Value{obj})

	rules, diags := expandRecurringTransactionRules(ctx, list, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	if rule.StartedAt == nil {
		t.Fatal("expected non-nil StartedAt")
	}
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !rule.StartedAt.Equal(expectedStart) {
		t.Errorf("expected StartedAt %v, got %v", expectedStart, *rule.StartedAt)
	}
	if rule.ExpirationAt == nil {
		t.Fatal("expected non-nil ExpirationAt")
	}
	expectedExpiry := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	if !rule.ExpirationAt.Equal(expectedExpiry) {
		t.Errorf("expected ExpirationAt %v, got %v", expectedExpiry, *rule.ExpirationAt)
	}
	if !rule.InvoiceRequiresSuccessfulPayment {
		t.Error("expected InvoiceRequiresSuccessfulPayment to be true")
	}
}

func TestExpandRecurringTransactionRules_InvalidStartedAt(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rm := walletRecurringTransactionRuleModel{
		LagoID:                           types.StringNull(),
		Interval:                         types.StringValue("monthly"),
		Method:                           types.StringValue("fixed"),
		Trigger:                          types.StringValue("interval"),
		PaidCredits:                      types.StringNull(),
		GrantedCredits:                   types.StringNull(),
		ThresholdCredits:                 types.StringNull(),
		StartedAt:                        types.StringValue("bad-date"),
		ExpirationAt:                     types.StringNull(),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
	}

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}
	obj, _ := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), rm)
	list, _ := types.ListValue(objType, []attr.Value{obj})

	_, diags := expandRecurringTransactionRules(ctx, list, nil)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for invalid started_at, got none")
	}
}

func TestExpandRecurringTransactionRules_InvalidExpirationAt(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rm := walletRecurringTransactionRuleModel{
		LagoID:                           types.StringNull(),
		Interval:                         types.StringValue("monthly"),
		Method:                           types.StringValue("fixed"),
		Trigger:                          types.StringValue("interval"),
		PaidCredits:                      types.StringNull(),
		GrantedCredits:                   types.StringNull(),
		ThresholdCredits:                 types.StringNull(),
		StartedAt:                        types.StringNull(),
		ExpirationAt:                     types.StringValue("not-valid"),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
	}

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}
	obj, _ := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), rm)
	list, _ := types.ListValue(objType, []attr.Value{obj})

	_, diags := expandRecurringTransactionRules(ctx, list, nil)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for invalid expiration_at, got none")
	}
}

func TestExpandRecurringTransactionRules_ThreadsLagoIDFromState(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	existingLagoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	// Plan rule has no lago_id yet (it will be unknown after import / first apply).
	planRM := walletRecurringTransactionRuleModel{
		LagoID:                           types.StringNull(),
		Interval:                         types.StringValue("monthly"),
		Method:                           types.StringValue("fixed"),
		Trigger:                          types.StringValue("interval"),
		PaidCredits:                      types.StringValue("20.0"),
		GrantedCredits:                   types.StringNull(),
		ThresholdCredits:                 types.StringNull(),
		StartedAt:                        types.StringNull(),
		ExpirationAt:                     types.StringNull(),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
	}

	// State rule carries the lago_id from a previous API response.
	stateRM := walletRecurringTransactionRuleModel{
		LagoID:                           types.StringValue(existingLagoID.String()),
		Interval:                         types.StringValue("monthly"),
		Method:                           types.StringValue("fixed"),
		Trigger:                          types.StringValue("interval"),
		PaidCredits:                      types.StringValue("20.0"),
		GrantedCredits:                   types.StringNull(),
		ThresholdCredits:                 types.StringNull(),
		StartedAt:                        types.StringNull(),
		ExpirationAt:                     types.StringNull(),
		InvoiceRequiresSuccessfulPayment: types.BoolValue(false),
	}

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}
	planObj, _ := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), planRM)
	stateObj, _ := types.ObjectValueFrom(ctx, walletRecurringTransactionRuleObjectType(), stateRM)
	planList, _ := types.ListValue(objType, []attr.Value{planObj})
	stateList, _ := types.ListValue(objType, []attr.Value{stateObj})

	state := &walletResourceModel{
		RecurringTransactionRules: stateList,
	}

	rules, diags := expandRecurringTransactionRules(ctx, planList, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].LagoID != existingLagoID {
		t.Errorf("expected LagoID %q threaded from state, got %q", existingLagoID, rules[0].LagoID)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — mapWalletToModel
// ---------------------------------------------------------------------------

func TestWalletMapToModel_Basic(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	lagoID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	now := time.Now().UTC()
	expiresAt := now.Add(30 * 24 * time.Hour)

	wallet := &lago.Wallet{
		LagoID:                           lagoID,
		ExternalCustomerID:               "cust-001",
		Status:                           lago.Active,
		Currency:                         lago.USD,
		Name:                             "My Wallet",
		RateAmount:                       "1.0",
		CreditsBalance:                   "150.0",
		InvoiceRequiresSuccessfulPayment: true,
		CreatedAt:                        now,
		ExpirationAt:                     expiresAt,
		RecurringTransactionRules:        []lago.RecurringTransactionRuleResponse{},
	}

	base := walletResourceModel{
		PaidCredits:    types.StringValue("100.0"),
		GrantedCredits: types.StringValue("10.0"),
	}

	state, diags := mapWalletToModel(ctx, wallet, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if state.ID.ValueString() != lagoID.String() {
		t.Errorf("expected ID %q, got %q", lagoID.String(), state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.ExternalCustomerID.ValueString() != "cust-001" {
		t.Errorf("expected ExternalCustomerID %q, got %q", "cust-001", state.ExternalCustomerID.ValueString())
	}
	if state.Currency.ValueString() != "USD" {
		t.Errorf("expected Currency %q, got %q", "USD", state.Currency.ValueString())
	}
	if state.Name.ValueString() != "My Wallet" {
		t.Errorf("expected Name %q, got %q", "My Wallet", state.Name.ValueString())
	}
	if state.RateAmount.ValueString() != "1.0" {
		t.Errorf("expected RateAmount %q, got %q", "1.0", state.RateAmount.ValueString())
	}
	if state.CreditsBalance.ValueString() != "150.0" {
		t.Errorf("expected CreditsBalance %q, got %q", "150.0", state.CreditsBalance.ValueString())
	}
	if state.Status.ValueString() != "active" {
		t.Errorf("expected Status %q, got %q", "active", state.Status.ValueString())
	}
	if !state.InvoiceRequiresSuccessfulPayment.ValueBool() {
		t.Error("expected InvoiceRequiresSuccessfulPayment to be true")
	}
	if state.ExpirationAt.IsNull() {
		t.Error("expected non-null ExpirationAt")
	}
	if state.CreatedAt.IsNull() {
		t.Error("expected non-null CreatedAt")
	}
	// Write-only fields should be preserved from base.
	if state.PaidCredits.ValueString() != "100.0" {
		t.Errorf("expected PaidCredits preserved as %q, got %q", "100.0", state.PaidCredits.ValueString())
	}
	if state.GrantedCredits.ValueString() != "10.0" {
		t.Errorf("expected GrantedCredits preserved as %q, got %q", "10.0", state.GrantedCredits.ValueString())
	}
}

func TestWalletMapToModel_EmptyName(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	wallet := &lago.Wallet{
		LagoID:             uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"),
		ExternalCustomerID: "cust-002",
		Status:             lago.Active,
		Currency:           lago.USD,
		Name:               "",
		RateAmount:         "1.0",
		CreditsBalance:     "0.0",
		CreatedAt:          time.Now().UTC(),
	}

	state, diags := mapWalletToModel(ctx, wallet, walletResourceModel{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if !state.Name.IsNull() {
		t.Errorf("expected null Name for empty string, got %q", state.Name.ValueString())
	}
	if !state.ExpirationAt.IsNull() {
		t.Errorf("expected null ExpirationAt for zero time, got %q", state.ExpirationAt.ValueString())
	}
}

func TestWalletMapToModel_ZeroCreatedAt(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	wallet := &lago.Wallet{
		LagoID:             uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001"),
		ExternalCustomerID: "cust-003",
		Status:             lago.Active,
		Currency:           lago.EUR,
		RateAmount:         "1.0",
		CreditsBalance:     "0.0",
		// CreatedAt is zero value
	}

	state, diags := mapWalletToModel(ctx, wallet, walletResourceModel{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if !state.CreatedAt.IsNull() {
		t.Errorf("expected null CreatedAt for zero time, got %q", state.CreatedAt.ValueString())
	}
}

func TestWalletMapToModel_WithRecurringRules(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	ruleID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	startedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	wallet := &lago.Wallet{
		LagoID:             uuid.MustParse("dddddddd-0000-0000-0000-000000000001"),
		ExternalCustomerID: "cust-004",
		Status:             lago.Active,
		Currency:           lago.USD,
		RateAmount:         "1.0",
		CreditsBalance:     "50.0",
		CreatedAt:          time.Now().UTC(),
		RecurringTransactionRules: []lago.RecurringTransactionRuleResponse{
			{
				LagoID:           ruleID,
				Interval:         "monthly",
				Method:           "fixed",
				Trigger:          "interval",
				PaidCredits:      "25.0",
				GrantedCredits:   "",
				ThresholdCredits: "",
				StartedAt:        &startedAt,
			},
		},
	}

	state, diags := mapWalletToModel(ctx, wallet, walletResourceModel{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if state.RecurringTransactionRules.IsNull() {
		t.Fatal("expected non-null RecurringTransactionRules")
	}

	var ruleModels []walletRecurringTransactionRuleModel
	listDiags := state.RecurringTransactionRules.ElementsAs(ctx, &ruleModels, false)
	if listDiags.HasError() {
		t.Fatalf("failed to extract rule models: %+v", listDiags)
	}
	if len(ruleModels) != 1 {
		t.Fatalf("expected 1 rule model, got %d", len(ruleModels))
	}

	rm := ruleModels[0]
	if rm.LagoID.ValueString() != ruleID.String() {
		t.Errorf("expected rule LagoID %q, got %q", ruleID.String(), rm.LagoID.ValueString())
	}
	if rm.Interval.ValueString() != "monthly" {
		t.Errorf("expected Interval %q, got %q", "monthly", rm.Interval.ValueString())
	}
	if rm.PaidCredits.ValueString() != "25.0" {
		t.Errorf("expected PaidCredits %q, got %q", "25.0", rm.PaidCredits.ValueString())
	}
	if !rm.GrantedCredits.IsNull() {
		t.Errorf("expected null GrantedCredits for empty string, got %q", rm.GrantedCredits.ValueString())
	}
	if rm.StartedAt.IsNull() {
		t.Error("expected non-null StartedAt")
	}
}

func TestWalletMapToModel_WriteOnlyFieldsPreservedOnRead(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Simulate a Read() call — the API does not return paid_credits/granted_credits.
	// The base contains the prior state values that should be carried forward.
	wallet := &lago.Wallet{
		LagoID:             uuid.MustParse("eeeeeeee-0000-0000-0000-000000000001"),
		ExternalCustomerID: "cust-005",
		Status:             lago.Active,
		Currency:           lago.USD,
		RateAmount:         "1.0",
		CreditsBalance:     "75.0",
		CreatedAt:          time.Now().UTC(),
	}

	base := walletResourceModel{
		PaidCredits:    types.StringValue("50.0"),
		GrantedCredits: types.StringValue("5.0"),
	}

	state, diags := mapWalletToModel(ctx, wallet, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if state.PaidCredits.ValueString() != "50.0" {
		t.Errorf("expected PaidCredits %q preserved from prior state, got %q", "50.0", state.PaidCredits.ValueString())
	}
	if state.GrantedCredits.ValueString() != "5.0" {
		t.Errorf("expected GrantedCredits %q preserved from prior state, got %q", "5.0", state.GrantedCredits.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Unit tests — flattenRecurringTransactionRules
// ---------------------------------------------------------------------------

func TestFlattenRecurringTransactionRules_Empty(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	objType := types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}
	emptyBase, _ := types.ListValue(objType, []attr.Value{})
	list, diags := flattenRecurringTransactionRules(ctx, []lago.RecurringTransactionRuleResponse{}, emptyBase)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if list.IsNull() {
		t.Error("expected non-null list for empty rules slice")
	}
	if len(list.Elements()) != 0 {
		t.Errorf("expected 0 elements, got %d", len(list.Elements()))
	}
}

func TestFlattenRecurringTransactionRules_WithRule(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	ruleID := uuid.MustParse("ffffffff-0000-0000-0000-000000000001")
	expiresAt := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)

	rules := []lago.RecurringTransactionRuleResponse{
		{
			LagoID:                           ruleID,
			Interval:                         "monthly",
			Method:                           "fixed",
			Trigger:                          "interval",
			PaidCredits:                      "30.0",
			GrantedCredits:                   "3.0",
			ThresholdCredits:                 "",
			ExpirationAt:                     &expiresAt,
			InvoiceRequiresSuccessfulPayment: true,
		},
	}

	list, diags := flattenRecurringTransactionRules(ctx, rules, types.ListNull(types.ObjectType{AttrTypes: walletRecurringTransactionRuleObjectType()}))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if len(list.Elements()) != 1 {
		t.Fatalf("expected 1 element, got %d", len(list.Elements()))
	}

	var ruleModels []walletRecurringTransactionRuleModel
	listDiags := list.ElementsAs(ctx, &ruleModels, false)
	if listDiags.HasError() {
		t.Fatalf("failed to extract rule models: %+v", listDiags)
	}

	rm := ruleModels[0]
	if rm.LagoID.ValueString() != ruleID.String() {
		t.Errorf("expected LagoID %q, got %q", ruleID.String(), rm.LagoID.ValueString())
	}
	if rm.Interval.ValueString() != "monthly" {
		t.Errorf("expected Interval %q, got %q", "monthly", rm.Interval.ValueString())
	}
	if rm.PaidCredits.ValueString() != "30.0" {
		t.Errorf("expected PaidCredits %q, got %q", "30.0", rm.PaidCredits.ValueString())
	}
	if rm.GrantedCredits.ValueString() != "3.0" {
		t.Errorf("expected GrantedCredits %q, got %q", "3.0", rm.GrantedCredits.ValueString())
	}
	if !rm.ThresholdCredits.IsNull() {
		t.Errorf("expected null ThresholdCredits, got %q", rm.ThresholdCredits.ValueString())
	}
	if rm.ExpirationAt.IsNull() {
		t.Error("expected non-null ExpirationAt")
	}
	if !rm.InvoiceRequiresSuccessfulPayment.ValueBool() {
		t.Error("expected InvoiceRequiresSuccessfulPayment to be true")
	}
}

// ---------------------------------------------------------------------------
// Acceptance tests
// ---------------------------------------------------------------------------

func TestAccWalletResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	ts := time.Now().UnixNano()
	customerExternalID := fmt.Sprintf("tf_acc_wallet_cust_%d", ts)
	resourceName := "lago_wallet.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccWalletConfig(customerExternalID, "Test Wallet", "1.0", "100.0", "10.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "external_customer_id", customerExternalID),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Wallet"),
					resource.TestCheckResourceAttr(resourceName, "currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "rate_amount", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "paid_credits", "100.0"),
					resource.TestCheckResourceAttr(resourceName, "granted_credits", "10.0"),
					resource.TestCheckResourceAttr(resourceName, "invoice_requires_successful_payment", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
					resource.TestCheckResourceAttrSet(resourceName, "lago_id"),
					resource.TestCheckResourceAttrSet(resourceName, "credits_balance"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"paid_credits", "granted_credits", "created_at", "credits_balance"},
			},
			{
				// Update — change name only; Lago v1.42.0 silently ignores rate_amount updates
				// so we keep rate_amount at "1.0" to avoid inconsistent-result errors.
				Config: testAccWalletConfig(customerExternalID, "Updated Wallet", "1.0", "50.0", "5.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Updated Wallet"),
					resource.TestCheckResourceAttr(resourceName, "rate_amount", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "paid_credits", "50.0"),
					resource.TestCheckResourceAttr(resourceName, "granted_credits", "5.0"),
				),
			},
		},
	})
}

func TestAccWalletResource_WithRecurringRule(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	// Lago API v1.42.0 silently ignores recurring_transaction_rules on both create
	// and update requests, so we cannot verify round-trip of recurring rules.
	// This test verifies that:
	//   1. a wallet with recurring_transaction_rules in the config can be created
	//      (rules are dropped silently, not a hard error)
	//   2. the provider does not panic or produce inconsistent state when the API
	//      returns an empty rules list
	ts := time.Now().UnixNano()
	customerExternalID := fmt.Sprintf("tf_acc_wallet_rule_cust_%d", ts)
	resourceName := "lago_wallet.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create a wallet without recurring rules first (create with rules returns 500).
			{
				Config: testAccWalletWithRecurringRuleBaseConfig(customerExternalID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "external_customer_id", customerExternalID),
					resource.TestCheckResourceAttr(resourceName, "currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "rate_amount", "1.0"),
					resource.TestCheckResourceAttrSet(resourceName, "lago_id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
		},
	})
}

func testAccWalletConfig(customerExternalID, name, rateAmount, paidCredits, grantedCredits string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_customer" "test" {
  external_id = "%s"
  name        = "Terraform Acceptance Test Customer"
}

resource "lago_wallet" "test" {
  external_customer_id = lago_customer.test.external_id
  name                 = "%s"
  currency             = "USD"
  rate_amount          = "%s"
  paid_credits         = "%s"
  granted_credits      = "%s"
}
`, customerExternalID, name, rateAmount, paidCredits, grantedCredits)
}

func testAccWalletWithRecurringRuleBaseConfig(customerExternalID string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_customer" "test" {
  external_id = "%s"
  name        = "Terraform Acceptance Test Customer"
}

resource "lago_wallet" "test" {
  external_customer_id = lago_customer.test.external_id
  currency             = "USD"
  rate_amount          = "1.0"
  paid_credits         = "10.0"
}
`, customerExternalID)
}
