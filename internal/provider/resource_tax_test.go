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

func TestTaxExpandInput(t *testing.T) {
	t.Parallel()

	plan := taxResourceModel{
		ID:                    types.StringNull(),
		LagoID:                types.StringNull(),
		Name:                  types.StringValue("VAT"),
		Code:                  types.StringValue("vat_20"),
		Description:           types.StringValue("Standard 20% VAT"),
		Rate:                  types.Float64Value(20.0),
		AppliedToOrganization: types.BoolValue(false),
		CreatedAt:             types.StringNull(),
	}

	input := expandTaxInput(plan)

	if input.Name != "VAT" {
		t.Errorf("expected Name %q, got %q", "VAT", input.Name)
	}
	if input.Code != "vat_20" {
		t.Errorf("expected Code %q, got %q", "vat_20", input.Code)
	}
	if input.Rate == nil {
		t.Fatal("expected non-nil Rate")
	}
	if *input.Rate != float32(20.0) {
		t.Errorf("expected Rate 20.0, got %f", *input.Rate)
	}
	if input.Description != "Standard 20% VAT" {
		t.Errorf("expected Description %q, got %q", "Standard 20% VAT", input.Description)
	}
	if input.AppliedToOrganization != false {
		t.Errorf("expected AppliedToOrganization false, got true")
	}
}

func TestTaxExpandInputAppliedToOrg(t *testing.T) {
	t.Parallel()

	plan := taxResourceModel{
		ID:                    types.StringNull(),
		LagoID:                types.StringNull(),
		Name:                  types.StringValue("GST"),
		Code:                  types.StringValue("gst_10"),
		Description:           types.StringNull(),
		Rate:                  types.Float64Value(10.0),
		AppliedToOrganization: types.BoolValue(true),
		CreatedAt:             types.StringNull(),
	}

	input := expandTaxInput(plan)

	if input.AppliedToOrganization != true {
		t.Errorf("expected AppliedToOrganization true, got false")
	}
	// Null description should not be set in input
	if input.Description != "" {
		t.Errorf("expected empty Description for null input, got %q", input.Description)
	}
}

func TestTaxMapToModel(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	now := time.Now().UTC()

	tax := &lago.Tax{
		LagoID:                lagoID,
		Name:                  "VAT",
		Code:                  "vat_20",
		Rate:                  float32(20.0),
		Description:           "Standard 20% VAT",
		AppliedToOrganization: false,
		CreatedAt:             now,
	}

	base := taxResourceModel{}
	state := mapTaxToModel(tax, base)

	if state.ID.ValueString() != "vat_20" {
		t.Errorf("expected ID %q, got %q", "vat_20", state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.Name.ValueString() != "VAT" {
		t.Errorf("expected Name %q, got %q", "VAT", state.Name.ValueString())
	}
	if state.Code.ValueString() != "vat_20" {
		t.Errorf("expected Code %q, got %q", "vat_20", state.Code.ValueString())
	}
	if state.Rate.ValueFloat64() != float64(float32(20.0)) {
		t.Errorf("expected Rate %f, got %f", float64(float32(20.0)), state.Rate.ValueFloat64())
	}
	if state.Description.ValueString() != "Standard 20% VAT" {
		t.Errorf("expected Description %q, got %q", "Standard 20% VAT", state.Description.ValueString())
	}
	if state.AppliedToOrganization.ValueBool() != false {
		t.Errorf("expected AppliedToOrganization false, got true")
	}
	if state.CreatedAt.IsNull() {
		t.Fatal("expected non-null CreatedAt")
	}
}

func TestTaxMapToModelEmptyOptionals(t *testing.T) {
	t.Parallel()

	tax := &lago.Tax{
		LagoID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:   "Simple Tax",
		Code:   "simple_tax",
		Rate:   float32(5.0),
	}

	base := taxResourceModel{}
	state := mapTaxToModel(tax, base)

	if !state.Description.IsNull() {
		t.Errorf("expected null Description for empty string, got %q", state.Description.ValueString())
	}
	if !state.CreatedAt.IsNull() {
		t.Errorf("expected null CreatedAt for zero time, got %q", state.CreatedAt.ValueString())
	}
	if state.AppliedToOrganization.ValueBool() != false {
		t.Errorf("expected AppliedToOrganization false, got true")
	}
}

func TestTaxRateRoundTrip(t *testing.T) {
	t.Parallel()

	// Test that float64 -> float32 -> float64 round-trip is consistent
	originalRate := 7.5
	plan := taxResourceModel{
		ID:                    types.StringNull(),
		LagoID:                types.StringNull(),
		Name:                  types.StringValue("Tax"),
		Code:                  types.StringValue("tax_7_5"),
		Description:           types.StringNull(),
		Rate:                  types.Float64Value(originalRate),
		AppliedToOrganization: types.BoolValue(false),
		CreatedAt:             types.StringNull(),
	}

	input := expandTaxInput(plan)
	if input.Rate == nil {
		t.Fatal("expected non-nil Rate")
	}

	// Simulate what the API returns: float32
	responseTax := &lago.Tax{
		LagoID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:   "Tax",
		Code:   "tax_7_5",
		Rate:   *input.Rate,
	}

	state := mapTaxToModel(responseTax, taxResourceModel{})
	// float64 -> float32 -> float64 may have minor precision differences; just confirm it's close
	got := state.Rate.ValueFloat64()
	if got < 7.4 || got > 7.6 {
		t.Errorf("rate round-trip drifted too far: got %f", got)
	}
}

func TestAccTaxResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	randomCode := fmt.Sprintf("tf_acc_tax_%d", time.Now().UnixNano())
	resourceName := "lago_tax.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTaxConfig(randomCode, "Test Tax", 20.0, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "code", randomCode),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Tax"),
					resource.TestCheckResourceAttr(resourceName, "rate", "20"),
					resource.TestCheckResourceAttr(resourceName, "applied_to_organization", "false"),
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
				Config: testAccTaxConfig(randomCode, "Updated Tax", 15.0, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Updated Tax"),
					resource.TestCheckResourceAttr(resourceName, "rate", "15"),
				),
			},
		},
	})
}

func testAccTaxConfig(code, name string, rate float64, appliedToOrg bool) string {
	return fmt.Sprintf(`
provider "lago" {
  api_endpoint = "%s"
  api_key      = "%s"
}

resource "lago_tax" "test" {
  name                    = "%s"
  code                    = "%s"
  rate                    = %g
  applied_to_organization = %t
  description             = "Terraform acceptance test tax"
}
`, os.Getenv("LAGO_API_ENDPOINT"), os.Getenv("LAGO_API_KEY"), name, code, rate, appliedToOrg)
}
