package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSubscriptionExpandInput_Basic(t *testing.T) {
	t.Parallel()

	plan := subscriptionResourceModel{
		ID:                      types.StringNull(),
		LagoID:                  types.StringNull(),
		ExternalID:              types.StringValue("sub-001"),
		ExternalCustomerID:      types.StringValue("cust-001"),
		PlanCode:                types.StringValue("pro_plan"),
		Name:                    types.StringValue("My Subscription"),
		BillingTime:             types.StringValue("anniversary"),
		SubscriptionAt:          types.StringNull(),
		EndingAt:                types.StringNull(),
		Status:                  types.StringNull(),
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
		CreatedAt:               types.StringNull(),
		StartedAt:               types.StringNull(),
	}

	input, diags := expandSubscriptionInput(plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.ExternalID != "sub-001" {
		t.Errorf("expected ExternalID %q, got %q", "sub-001", input.ExternalID)
	}
	if input.ExternalCustomerID != "cust-001" {
		t.Errorf("expected ExternalCustomerID %q, got %q", "cust-001", input.ExternalCustomerID)
	}
	if input.PlanCode != "pro_plan" {
		t.Errorf("expected PlanCode %q, got %q", "pro_plan", input.PlanCode)
	}
	if input.Name != "My Subscription" {
		t.Errorf("expected Name %q, got %q", "My Subscription", input.Name)
	}
	if input.BillingTime != lago.Anniversary {
		t.Errorf("expected BillingTime %q, got %q", lago.Anniversary, input.BillingTime)
	}
	if input.SubscriptionAt != nil {
		t.Errorf("expected nil SubscriptionAt, got %v", input.SubscriptionAt)
	}
	if input.EndingAt != nil {
		t.Errorf("expected nil EndingAt, got %v", input.EndingAt)
	}
}

func TestSubscriptionExpandInput_WithDates(t *testing.T) {
	t.Parallel()

	plan := subscriptionResourceModel{
		ID:                      types.StringNull(),
		LagoID:                  types.StringNull(),
		ExternalID:              types.StringValue("sub-002"),
		ExternalCustomerID:      types.StringValue("cust-002"),
		PlanCode:                types.StringValue("starter_plan"),
		Name:                    types.StringNull(),
		BillingTime:             types.StringNull(),
		SubscriptionAt:          types.StringValue("2024-01-15T00:00:00Z"),
		EndingAt:                types.StringValue("2025-01-15T00:00:00Z"),
		Status:                  types.StringNull(),
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
		CreatedAt:               types.StringNull(),
		StartedAt:               types.StringNull(),
	}

	input, diags := expandSubscriptionInput(plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.SubscriptionAt == nil {
		t.Fatal("expected non-nil SubscriptionAt")
	}
	expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !input.SubscriptionAt.Equal(expected) {
		t.Errorf("expected SubscriptionAt %v, got %v", expected, *input.SubscriptionAt)
	}

	if input.EndingAt == nil {
		t.Fatal("expected non-nil EndingAt")
	}
	expectedEnding := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	if !input.EndingAt.Equal(expectedEnding) {
		t.Errorf("expected EndingAt %v, got %v", expectedEnding, *input.EndingAt)
	}
}

func TestSubscriptionExpandInput_InvalidSubscriptionAt(t *testing.T) {
	t.Parallel()

	plan := subscriptionResourceModel{
		ID:                      types.StringNull(),
		LagoID:                  types.StringNull(),
		ExternalID:              types.StringValue("sub-003"),
		ExternalCustomerID:      types.StringValue("cust-003"),
		PlanCode:                types.StringValue("pro_plan"),
		Name:                    types.StringNull(),
		BillingTime:             types.StringNull(),
		SubscriptionAt:          types.StringValue("not-a-date"),
		EndingAt:                types.StringNull(),
		Status:                  types.StringNull(),
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
		CreatedAt:               types.StringNull(),
		StartedAt:               types.StringNull(),
	}

	_, diags := expandSubscriptionInput(plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for invalid subscription_at, got none")
	}
}

func TestSubscriptionExpandInput_InvalidEndingAt(t *testing.T) {
	t.Parallel()

	plan := subscriptionResourceModel{
		ID:                      types.StringNull(),
		LagoID:                  types.StringNull(),
		ExternalID:              types.StringValue("sub-004"),
		ExternalCustomerID:      types.StringValue("cust-004"),
		PlanCode:                types.StringValue("pro_plan"),
		Name:                    types.StringNull(),
		BillingTime:             types.StringNull(),
		SubscriptionAt:          types.StringNull(),
		EndingAt:                types.StringValue("bad-date"),
		Status:                  types.StringNull(),
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
		CreatedAt:               types.StringNull(),
		StartedAt:               types.StringNull(),
	}

	_, diags := expandSubscriptionInput(plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for invalid ending_at, got none")
	}
}

func TestSubscriptionMapToModel_Basic(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	now := time.Now().UTC()
	startedAt := now.Add(-time.Hour)
	endingAt := now.Add(30 * 24 * time.Hour)

	subscription := &lago.Subscription{
		LagoID:             lagoID,
		ExternalID:         "sub-001",
		ExternalCustomerID: "cust-001",
		PlanCode:           "pro_plan",
		Name:               "My Subscription",
		Status:             lago.SubscriptionStatusActive,
		BillingTime:        lago.Anniversary,
		SubscriptionAt:     now,
		EndingAt:           &endingAt,
		StartedAt:          &startedAt,
		CreatedAt:          now,
	}

	base := subscriptionResourceModel{
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
	}
	state := mapSubscriptionToModel(subscription, base)

	if state.ID.ValueString() != "sub-001" {
		t.Errorf("expected ID %q, got %q", "sub-001", state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.ExternalID.ValueString() != "sub-001" {
		t.Errorf("expected ExternalID %q, got %q", "sub-001", state.ExternalID.ValueString())
	}
	if state.ExternalCustomerID.ValueString() != "cust-001" {
		t.Errorf("expected ExternalCustomerID %q, got %q", "cust-001", state.ExternalCustomerID.ValueString())
	}
	if state.PlanCode.ValueString() != "pro_plan" {
		t.Errorf("expected PlanCode %q, got %q", "pro_plan", state.PlanCode.ValueString())
	}
	if state.Name.ValueString() != "My Subscription" {
		t.Errorf("expected Name %q, got %q", "My Subscription", state.Name.ValueString())
	}
	if state.Status.ValueString() != "active" {
		t.Errorf("expected Status %q, got %q", "active", state.Status.ValueString())
	}
	if state.BillingTime.ValueString() != "anniversary" {
		t.Errorf("expected BillingTime %q, got %q", "anniversary", state.BillingTime.ValueString())
	}
	if state.SubscriptionAt.IsNull() {
		t.Error("expected non-null SubscriptionAt")
	}
	if state.EndingAt.IsNull() {
		t.Error("expected non-null EndingAt")
	}
	if state.StartedAt.IsNull() {
		t.Error("expected non-null StartedAt")
	}
	if state.CreatedAt.IsNull() {
		t.Error("expected non-null CreatedAt")
	}
}

func TestSubscriptionMapToModel_PendingStatus(t *testing.T) {
	t.Parallel()

	subscription := &lago.Subscription{
		LagoID:             uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001"),
		ExternalID:         "sub-pending",
		ExternalCustomerID: "cust-001",
		PlanCode:           "starter_plan",
		Status:             lago.SubscriptionStatusPending,
		CreatedAt:          time.Now().UTC(),
	}

	base := subscriptionResourceModel{
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
	}
	state := mapSubscriptionToModel(subscription, base)

	if state.Status.ValueString() != "pending" {
		t.Errorf("expected Status %q, got %q", "pending", state.Status.ValueString())
	}
	// nil *time.Time fields should be null
	if !state.EndingAt.IsNull() {
		t.Errorf("expected null EndingAt, got %q", state.EndingAt.ValueString())
	}
	if !state.StartedAt.IsNull() {
		t.Errorf("expected null StartedAt, got %q", state.StartedAt.ValueString())
	}
}

func TestSubscriptionMapToModel_EmptyName(t *testing.T) {
	t.Parallel()

	subscription := &lago.Subscription{
		LagoID:             uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001"),
		ExternalID:         "sub-noname",
		ExternalCustomerID: "cust-001",
		PlanCode:           "pro_plan",
		Name:               "",
		Status:             lago.SubscriptionStatusActive,
		CreatedAt:          time.Now().UTC(),
	}

	base := subscriptionResourceModel{
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
	}
	state := mapSubscriptionToModel(subscription, base)

	if !state.Name.IsNull() {
		t.Errorf("expected null Name for empty string, got %q", state.Name.ValueString())
	}
}

func TestSubscriptionMapToModel_OnTerminationFieldsPreserved(t *testing.T) {
	t.Parallel()

	// Simulate API returning empty on_termination fields — state should preserve
	// whatever was previously configured.
	subscription := &lago.Subscription{
		LagoID:                  uuid.MustParse("cccccccc-0000-0000-0000-000000000001"),
		ExternalID:              "sub-term",
		ExternalCustomerID:      "cust-001",
		PlanCode:                "pro_plan",
		Status:                  lago.SubscriptionStatusActive,
		OnTerminationCreditNote: "", // API returned empty
		OnTerminationInvoice:    "", // API returned empty
		CreatedAt:               time.Now().UTC(),
	}

	base := subscriptionResourceModel{
		OnTerminationCreditNote: types.StringValue("credit"),
		OnTerminationInvoice:    types.StringValue("generate"),
	}
	state := mapSubscriptionToModel(subscription, base)

	// Should preserve prior state values since API returned empty.
	if state.OnTerminationCreditNote.ValueString() != "credit" {
		t.Errorf("expected OnTerminationCreditNote %q preserved from state, got %q", "credit", state.OnTerminationCreditNote.ValueString())
	}
	if state.OnTerminationInvoice.ValueString() != "generate" {
		t.Errorf("expected OnTerminationInvoice %q preserved from state, got %q", "generate", state.OnTerminationInvoice.ValueString())
	}
}

func TestSubscriptionMapToModel_OnTerminationFieldsFromAPI(t *testing.T) {
	t.Parallel()

	// Simulate API returning non-empty on_termination fields.
	subscription := &lago.Subscription{
		LagoID:                  uuid.MustParse("dddddddd-0000-0000-0000-000000000001"),
		ExternalID:              "sub-term2",
		ExternalCustomerID:      "cust-001",
		PlanCode:                "pro_plan",
		Status:                  lago.SubscriptionStatusActive,
		OnTerminationCreditNote: lago.OnTerminationCreditNoteRefund,
		OnTerminationInvoice:    lago.OnTerminationInvoiceSkip,
		CreatedAt:               time.Now().UTC(),
	}

	base := subscriptionResourceModel{
		OnTerminationCreditNote: types.StringNull(),
		OnTerminationInvoice:    types.StringNull(),
	}
	state := mapSubscriptionToModel(subscription, base)

	if state.OnTerminationCreditNote.ValueString() != "refund" {
		t.Errorf("expected OnTerminationCreditNote %q, got %q", "refund", state.OnTerminationCreditNote.ValueString())
	}
	if state.OnTerminationInvoice.ValueString() != "skip" {
		t.Errorf("expected OnTerminationInvoice %q, got %q", "skip", state.OnTerminationInvoice.ValueString())
	}
}

func TestAccSubscriptionResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	ts := time.Now().UnixNano()
	customerCode := fmt.Sprintf("tf_acc_cust_%d", ts)
	planCode := fmt.Sprintf("tf_acc_plan_%d", ts)
	subExternalID := fmt.Sprintf("tf_acc_sub_%d", ts)
	resourceName := "lago_subscription.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionConfig(customerCode, planCode, subExternalID, "anniversary"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "external_id", subExternalID),
					resource.TestCheckResourceAttr(resourceName, "external_customer_id", customerCode),
					resource.TestCheckResourceAttr(resourceName, "plan_code", planCode),
					resource.TestCheckResourceAttr(resourceName, "billing_time", "anniversary"),
					resource.TestCheckResourceAttrSet(resourceName, "lago_id"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "started_at", "subscription_at"},
			},
		},
	})
}

func testAccSubscriptionConfig(customerExternalID, planCode, subExternalID, billingTime string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_customer" "test" {
  external_id = "%s"
  name        = "Terraform Acceptance Test Customer"
}

resource "lago_billable_metric" "test" {
  name             = "API Calls"
  code             = "%s_bm"
  aggregation_type = "count_agg"
}

resource "lago_plan" "test" {
  name            = "Test Plan"
  code            = "%s"
  interval        = "monthly"
  amount_cents    = 0
  amount_currency = "USD"
  pay_in_advance  = false
}

resource "lago_subscription" "test" {
  external_id          = "%s"
  external_customer_id = lago_customer.test.external_id
  plan_code            = lago_plan.test.code
  billing_time         = "%s"
}
`, customerExternalID, planCode, planCode, subExternalID, billingTime)
}
