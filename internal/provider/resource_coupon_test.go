package provider

import (
	"context"
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

func TestCouponExpandInput_FixedAmount(t *testing.T) {
	t.Parallel()

	plan := couponResourceModel{
		ID:                  types.StringNull(),
		LagoID:              types.StringNull(),
		Name:                types.StringValue("10 USD Off"),
		Code:                types.StringValue("ten_off"),
		Description:         types.StringValue("Ten dollar discount"),
		CouponType:          types.StringValue("fixed_amount"),
		AmountCents:         types.Int64Value(1000),
		AmountCurrency:      types.StringValue("USD"),
		PercentageRate:      types.Float64Null(),
		Expiration:          types.StringValue("no_expiration"),
		ExpirationAt:        types.StringNull(),
		Frequency:           types.StringValue("once"),
		FrequencyDuration:   types.Int64Null(),
		Reusable:            types.BoolValue(false),
		PlanCodes:           types.SetNull(types.StringType),
		BillableMetricCodes: types.SetNull(types.StringType),
		CreatedAt:           types.StringNull(),
	}

	input, diags := expandCouponInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.Name != "10 USD Off" {
		t.Errorf("expected Name %q, got %q", "10 USD Off", input.Name)
	}
	if input.Code != "ten_off" {
		t.Errorf("expected Code %q, got %q", "ten_off", input.Code)
	}
	if input.CouponType != lago.CouponTypeFixedAmount {
		t.Errorf("expected CouponType %q, got %q", lago.CouponTypeFixedAmount, input.CouponType)
	}
	if input.AmountCents != 1000 {
		t.Errorf("expected AmountCents 1000, got %d", input.AmountCents)
	}
	if string(input.AmountCurrency) != "USD" {
		t.Errorf("expected AmountCurrency USD, got %s", input.AmountCurrency)
	}
	if input.PercentageRate != 0 {
		t.Errorf("expected PercentageRate 0, got %f", input.PercentageRate)
	}
	if input.Expiration != lago.CouponExpirationNoExpiration {
		t.Errorf("expected Expiration %q, got %q", lago.CouponExpirationNoExpiration, input.Expiration)
	}
	if input.ExpirationAt != nil {
		t.Errorf("expected nil ExpirationAt, got %v", input.ExpirationAt)
	}
	if input.Frequency != lago.CouponFrequencyOnce {
		t.Errorf("expected Frequency %q, got %q", lago.CouponFrequencyOnce, input.Frequency)
	}
	if input.FrequencyDuration != 0 {
		t.Errorf("expected FrequencyDuration 0, got %d", input.FrequencyDuration)
	}
	if input.Reusable != false {
		t.Errorf("expected Reusable false, got true")
	}
}

func TestCouponExpandInput_Percentage(t *testing.T) {
	t.Parallel()

	plan := couponResourceModel{
		ID:                  types.StringNull(),
		LagoID:              types.StringNull(),
		Name:                types.StringValue("15% Off"),
		Code:                types.StringValue("fifteen_pct"),
		Description:         types.StringNull(),
		CouponType:          types.StringValue("percentage"),
		AmountCents:         types.Int64Null(),
		AmountCurrency:      types.StringNull(),
		PercentageRate:      types.Float64Value(15.0),
		Expiration:          types.StringValue("time_limit"),
		ExpirationAt:        types.StringValue("2025-12-31T23:59:59Z"),
		Frequency:           types.StringValue("recurring"),
		FrequencyDuration:   types.Int64Value(3),
		Reusable:            types.BoolValue(true),
		PlanCodes:           types.SetValueMust(types.StringType, []attr.Value{types.StringValue("pro_plan")}),
		BillableMetricCodes: types.SetNull(types.StringType),
		CreatedAt:           types.StringNull(),
	}

	input, diags := expandCouponInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.CouponType != lago.CouponTypePercentage {
		t.Errorf("expected CouponType %q, got %q", lago.CouponTypePercentage, input.CouponType)
	}
	if input.PercentageRate != 15.0 {
		t.Errorf("expected PercentageRate 15.0, got %f", input.PercentageRate)
	}
	if input.AmountCents != 0 {
		t.Errorf("expected AmountCents 0, got %d", input.AmountCents)
	}
	if input.Expiration != lago.CouponExpirationTimeLimit {
		t.Errorf("expected Expiration %q, got %q", lago.CouponExpirationTimeLimit, input.Expiration)
	}
	if input.ExpirationAt == nil {
		t.Fatal("expected non-nil ExpirationAt")
	}
	if input.Frequency != lago.CouponFrequencyRecurring {
		t.Errorf("expected Frequency %q, got %q", lago.CouponFrequencyRecurring, input.Frequency)
	}
	if input.FrequencyDuration != 3 {
		t.Errorf("expected FrequencyDuration 3, got %d", input.FrequencyDuration)
	}
	if input.Reusable != true {
		t.Errorf("expected Reusable true, got false")
	}
	if len(input.AppliesTo.PlanCodes) != 1 || input.AppliesTo.PlanCodes[0] != "pro_plan" {
		t.Errorf("expected AppliesTo.PlanCodes [pro_plan], got %v", input.AppliesTo.PlanCodes)
	}
}

func TestCouponExpandInput_MissingAmountCentsForFixedAmount(t *testing.T) {
	t.Parallel()

	plan := couponResourceModel{
		ID:                  types.StringNull(),
		LagoID:              types.StringNull(),
		Name:                types.StringValue("Bad Coupon"),
		Code:                types.StringValue("bad"),
		Description:         types.StringNull(),
		CouponType:          types.StringValue("fixed_amount"),
		AmountCents:         types.Int64Null(), // missing!
		AmountCurrency:      types.StringValue("USD"),
		PercentageRate:      types.Float64Null(),
		Expiration:          types.StringValue("no_expiration"),
		ExpirationAt:        types.StringNull(),
		Frequency:           types.StringValue("once"),
		FrequencyDuration:   types.Int64Null(),
		Reusable:            types.BoolValue(false),
		PlanCodes:           types.SetNull(types.StringType),
		BillableMetricCodes: types.SetNull(types.StringType),
		CreatedAt:           types.StringNull(),
	}

	_, diags := expandCouponInput(context.Background(), plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for missing amount_cents, got none")
	}
}

func TestCouponExpandInput_MissingPercentageRateForPercentage(t *testing.T) {
	t.Parallel()

	plan := couponResourceModel{
		ID:                  types.StringNull(),
		LagoID:              types.StringNull(),
		Name:                types.StringValue("Bad Coupon"),
		Code:                types.StringValue("bad"),
		Description:         types.StringNull(),
		CouponType:          types.StringValue("percentage"),
		AmountCents:         types.Int64Null(),
		AmountCurrency:      types.StringNull(),
		PercentageRate:      types.Float64Null(), // missing!
		Expiration:          types.StringValue("no_expiration"),
		ExpirationAt:        types.StringNull(),
		Frequency:           types.StringValue("once"),
		FrequencyDuration:   types.Int64Null(),
		Reusable:            types.BoolValue(false),
		PlanCodes:           types.SetNull(types.StringType),
		BillableMetricCodes: types.SetNull(types.StringType),
		CreatedAt:           types.StringNull(),
	}

	_, diags := expandCouponInput(context.Background(), plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for missing percentage_rate, got none")
	}
}

func TestCouponExpandInput_MissingExpirationAt(t *testing.T) {
	t.Parallel()

	plan := couponResourceModel{
		ID:                  types.StringNull(),
		LagoID:              types.StringNull(),
		Name:                types.StringValue("Time Limited"),
		Code:                types.StringValue("time_limited"),
		Description:         types.StringNull(),
		CouponType:          types.StringValue("fixed_amount"),
		AmountCents:         types.Int64Value(500),
		AmountCurrency:      types.StringValue("USD"),
		PercentageRate:      types.Float64Null(),
		Expiration:          types.StringValue("time_limit"),
		ExpirationAt:        types.StringNull(), // missing!
		Frequency:           types.StringValue("once"),
		FrequencyDuration:   types.Int64Null(),
		Reusable:            types.BoolValue(false),
		PlanCodes:           types.SetNull(types.StringType),
		BillableMetricCodes: types.SetNull(types.StringType),
		CreatedAt:           types.StringNull(),
	}

	_, diags := expandCouponInput(context.Background(), plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for missing expiration_at, got none")
	}
}

func TestCouponExpandInput_MissingFrequencyDuration(t *testing.T) {
	t.Parallel()

	plan := couponResourceModel{
		ID:                  types.StringNull(),
		LagoID:              types.StringNull(),
		Name:                types.StringValue("Recurring"),
		Code:                types.StringValue("recurring"),
		Description:         types.StringNull(),
		CouponType:          types.StringValue("fixed_amount"),
		AmountCents:         types.Int64Value(500),
		AmountCurrency:      types.StringValue("USD"),
		PercentageRate:      types.Float64Null(),
		Expiration:          types.StringValue("no_expiration"),
		ExpirationAt:        types.StringNull(),
		Frequency:           types.StringValue("recurring"),
		FrequencyDuration:   types.Int64Null(), // missing!
		Reusable:            types.BoolValue(false),
		PlanCodes:           types.SetNull(types.StringType),
		BillableMetricCodes: types.SetNull(types.StringType),
		CreatedAt:           types.StringNull(),
	}

	_, diags := expandCouponInput(context.Background(), plan)
	if !diags.HasError() {
		t.Fatal("expected diagnostics error for missing frequency_duration, got none")
	}
}

func TestCouponMapToModel_FixedAmount(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-000000000001")
	now := time.Now().UTC()

	coupon := &lago.Coupon{
		LagoID:         lagoID,
		Name:           "10 USD Off",
		Code:           "ten_off",
		Description:    "Ten dollar discount",
		CouponType:     lago.CouponTypeFixedAmount,
		AmountCents:    1000,
		AmountCurrency: lago.USD,
		PercentageRate: 0,
		Expiration:     lago.CouponExpirationNoExpiration,
		ExpirationAt:   nil,
		Frequency:      lago.CouponFrequencyOnce,
		Reusable:       false,
		CreatedAt:      now,
	}

	base := couponResourceModel{
		PlanCodes:           types.SetNull(types.StringType),
		BillableMetricCodes: types.SetNull(types.StringType),
	}
	state, diags := mapCouponToModel(coupon, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if state.ID.ValueString() != "ten_off" {
		t.Errorf("expected ID %q, got %q", "ten_off", state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.CouponType.ValueString() != "fixed_amount" {
		t.Errorf("expected CouponType %q, got %q", "fixed_amount", state.CouponType.ValueString())
	}
	if state.AmountCents.ValueInt64() != 1000 {
		t.Errorf("expected AmountCents 1000, got %d", state.AmountCents.ValueInt64())
	}
	if state.AmountCurrency.ValueString() != "USD" {
		t.Errorf("expected AmountCurrency USD, got %q", state.AmountCurrency.ValueString())
	}
	// percentage_rate should be null for fixed_amount coupons
	if !state.PercentageRate.IsNull() {
		t.Errorf("expected null PercentageRate for fixed_amount coupon, got %f", state.PercentageRate.ValueFloat64())
	}
	if state.ExpirationAt.IsNull() != true {
		t.Errorf("expected null ExpirationAt, got %q", state.ExpirationAt.ValueString())
	}
	// frequency_duration should be null for once frequency
	if !state.FrequencyDuration.IsNull() {
		t.Errorf("expected null FrequencyDuration for once frequency, got %d", state.FrequencyDuration.ValueInt64())
	}
	if state.CreatedAt.IsNull() {
		t.Error("expected non-null CreatedAt")
	}
}

func TestCouponMapToModel_Percentage(t *testing.T) {
	t.Parallel()

	expAt := time.Now().Add(24 * time.Hour).UTC()
	lagoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-000000000002")

	coupon := &lago.Coupon{
		LagoID:            lagoID,
		Name:              "15% Off",
		Code:              "fifteen_pct",
		CouponType:        lago.CouponTypePercentage,
		AmountCents:       0,
		PercentageRate:    15.0,
		Expiration:        lago.CouponExpirationTimeLimit,
		ExpirationAt:      &expAt,
		Frequency:         lago.CouponFrequencyRecurring,
		FrequencyDuration: 3,
		Reusable:          true,
		PlanCodes:         []string{"pro_plan"},
		CreatedAt:         time.Now().UTC(),
	}

	base := couponResourceModel{
		PlanCodes:           types.SetValueMust(types.StringType, []attr.Value{types.StringValue("pro_plan")}),
		BillableMetricCodes: types.SetNull(types.StringType),
	}
	state, diags := mapCouponToModel(coupon, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if state.CouponType.ValueString() != "percentage" {
		t.Errorf("expected CouponType %q, got %q", "percentage", state.CouponType.ValueString())
	}
	// amount_cents and amount_currency should be null for percentage coupons
	if !state.AmountCents.IsNull() {
		t.Errorf("expected null AmountCents for percentage coupon, got %d", state.AmountCents.ValueInt64())
	}
	if !state.AmountCurrency.IsNull() {
		t.Errorf("expected null AmountCurrency for percentage coupon, got %q", state.AmountCurrency.ValueString())
	}
	if state.PercentageRate.ValueFloat64() != 15.0 {
		t.Errorf("expected PercentageRate 15.0, got %f", state.PercentageRate.ValueFloat64())
	}
	if state.ExpirationAt.IsNull() {
		t.Error("expected non-null ExpirationAt")
	}
	if state.FrequencyDuration.ValueInt64() != 3 {
		t.Errorf("expected FrequencyDuration 3, got %d", state.FrequencyDuration.ValueInt64())
	}
	if state.Reusable.ValueBool() != true {
		t.Error("expected Reusable true")
	}
}

func TestAccCouponResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	randomCode := fmt.Sprintf("tf_acc_coupon_%d", time.Now().UnixNano())
	resourceName := "lago_coupon.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCouponFixedAmountConfig(randomCode, "Test Coupon", 1000, "USD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "code", randomCode),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Coupon"),
					resource.TestCheckResourceAttr(resourceName, "coupon_type", "fixed_amount"),
					resource.TestCheckResourceAttr(resourceName, "amount_cents", "1000"),
					resource.TestCheckResourceAttr(resourceName, "amount_currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "expiration", "no_expiration"),
					resource.TestCheckResourceAttr(resourceName, "frequency", "once"),
					resource.TestCheckResourceAttrSet(resourceName, "lago_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at"},
			},
			{
				Config: testAccCouponFixedAmountConfig(randomCode, "Updated Test Coupon", 2000, "EUR"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Updated Test Coupon"),
					resource.TestCheckResourceAttr(resourceName, "amount_cents", "2000"),
					resource.TestCheckResourceAttr(resourceName, "amount_currency", "EUR"),
				),
			},
		},
	})
}

func testAccCouponFixedAmountConfig(code, name string, amountCents int, currency string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_coupon" "test" {
  name            = "%s"
  code            = "%s"
  coupon_type     = "fixed_amount"
  amount_cents    = %d
  amount_currency = "%s"
  expiration      = "no_expiration"
  frequency       = "once"
  reusable        = false
}
`, name, code, amountCents, currency)
}
