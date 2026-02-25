package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

func TestPlanExpandFlattenCharges(t *testing.T) {
	t.Parallel()

	bmID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	objectType := chargeObjectType()
	setType := types.SetType{ElemType: types.StringType}
	chargeList := types.ListValueMust(objectType, []attr.Value{
		types.ObjectValueMust(objectType.AttrTypes, map[string]attr.Value{
			"billable_metric_id":   types.StringValue(bmID.String()),
			"charge_model":         types.StringValue("standard"),
			"invoiceable":          types.BoolValue(true),
			"invoice_display_name": types.StringValue("Requests"),
			"pay_in_advance":       types.BoolValue(false),
			"regroup_paid_fees":    types.BoolValue(false),
			"prorated":             types.BoolValue(false),
			"min_amount_cents":     types.Int64Value(0),
			"properties_json":      types.StringValue(`{"amount":"1"}`),
			"filters_json":         types.StringNull(),
			"tax_codes":            types.SetValueMust(types.StringType, []attr.Value{types.StringValue("tx_1")}),
		}),
	})

	charges, diags := expandCharges(context.Background(), chargeList)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics expanding charges: %+v", diags)
	}
	if len(charges) != 1 {
		t.Fatalf("expected one charge, got %d", len(charges))
	}
	if charges[0].BillableMetricID != bmID {
		t.Fatalf("unexpected billable_metric_id: %s", charges[0].BillableMetricID)
	}

	// flattenCharges takes []lago.Charge (response type); build one to test the flatten path
	responseCharges := []lago.Charge{
		{
			LagoBillableMetricID: bmID,
			ChargeModel:          lago.StandardChargeModel,
			Invoiceable:          true,
			PayInAdvance:         false,
			RegroupPaidFees:      "",
			Prorated:             false,
			MinAmountCents:       0,
			Properties:           map[string]interface{}{"amount": "1"},
		},
	}
	flattened, flattenDiags := flattenCharges(responseCharges)
	if flattenDiags.HasError() {
		t.Fatalf("unexpected diagnostics flattening charges: %+v", flattenDiags)
	}
	if flattened.IsNull() {
		t.Fatal("expected non-null flattened charges")
	}
	_ = setType
}

func TestPlanIntervalValidation(t *testing.T) {
	t.Parallel()

	resourceName := "lago_plan.test"
	code := fmt.Sprintf("tf_acc_plan_invalid_%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccPlanInvalidIntervalConfig(code),
				ExpectError: regexp.MustCompile("interval"),
				Check:       resource.TestCheckNoResourceAttr(resourceName, "id"),
			},
		},
	})
}

func TestAccPlanResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	code := fmt.Sprintf("tf_acc_plan_%d", time.Now().UnixNano())
	newCode := code + "_new"
	resourceName := "lago_plan.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccPlanConfig(code, "Starter", 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "code", code),
					resource.TestCheckResourceAttr(resourceName, "interval", "monthly"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at", "lago_id"},
			},
			{
				Config: testAccPlanConfig(code, "Starter Updated", 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Starter Updated"),
					resource.TestCheckResourceAttr(resourceName, "amount_cents", "2000"),
				),
			},
			{
				Config: testAccPlanConfig(newCode, "Starter Updated", 2000),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "code", newCode),
				),
			},
		},
	})
}

func testAccPlanConfig(code, name string, amountCents int64) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_plan" "test" {
  name            = "%s"
  code            = "%s"
  interval        = "monthly"
  amount_cents    = %d
  amount_currency = "USD"
  pay_in_advance  = false
}
`, name, code, amountCents)
}

func testAccPlanInvalidIntervalConfig(code string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_plan" "test" {
  name            = "Invalid"
  code            = "%s"
  interval        = "invalid"
  amount_cents    = 1000
  amount_currency = "USD"
}
`, code)
}
