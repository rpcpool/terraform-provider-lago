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

func TestAddOnExpandFlatten(t *testing.T) {
	t.Parallel()

	plan := addOnResourceModel{
		ID:                 types.StringNull(),
		LagoID:             types.StringNull(),
		Name:               types.StringValue("Setup Fee"),
		Code:               types.StringValue("setup_fee"),
		Description:        types.StringValue("A one-time setup fee"),
		AmountCents:        types.Int64Value(10000),
		AmountCurrency:     types.StringValue("USD"),
		InvoiceDisplayName: types.StringValue("Setup Fee (display)"),
		TaxCodes:           types.SetValueMust(types.StringType, nil),
		CreatedAt:          types.StringNull(),
	}

	input, diags := expandAddOnInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics expanding add-on: %+v", diags)
	}

	if input.Name != "Setup Fee" {
		t.Errorf("expected Name %q, got %q", "Setup Fee", input.Name)
	}
	if input.Code != "setup_fee" {
		t.Errorf("expected Code %q, got %q", "setup_fee", input.Code)
	}
	if input.AmountCents != 10000 {
		t.Errorf("expected AmountCents 10000, got %d", input.AmountCents)
	}
	if string(input.AmountCurrency) != "USD" {
		t.Errorf("expected AmountCurrency USD, got %s", input.AmountCurrency)
	}
	if input.Description != "A one-time setup fee" {
		t.Errorf("expected Description %q, got %q", "A one-time setup fee", input.Description)
	}
	if input.InvoiceDisplayName != "Setup Fee (display)" {
		t.Errorf("expected InvoiceDisplayName %q, got %q", "Setup Fee (display)", input.InvoiceDisplayName)
	}
}

func TestAddOnExpandFlattenWithTaxCodes(t *testing.T) {
	t.Parallel()

	plan := addOnResourceModel{
		ID:                 types.StringNull(),
		LagoID:             types.StringNull(),
		Name:               types.StringValue("Setup Fee"),
		Code:               types.StringValue("setup_fee"),
		Description:        types.StringNull(),
		AmountCents:        types.Int64Value(5000),
		AmountCurrency:     types.StringValue("EUR"),
		InvoiceDisplayName: types.StringNull(),
		TaxCodes: types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue("vat_20"),
			types.StringValue("local_tax"),
		}),
		CreatedAt: types.StringNull(),
	}

	input, diags := expandAddOnInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics expanding add-on with tax codes: %+v", diags)
	}

	if len(input.TaxCodes) != 2 {
		t.Fatalf("expected 2 tax codes, got %d", len(input.TaxCodes))
	}
}

func TestAddOnMapToModel(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("00000000-0000-0000-0000-000000000042")
	now := time.Now().UTC()

	addOn := &lago.AddOn{
		LagoID:             lagoID,
		Name:               "Setup Fee",
		Code:               "setup_fee",
		Description:        "A one-time setup fee",
		AmountCents:        10000,
		AmountCurrency:     lago.EUR,
		InvoiceDisplayName: "Setup Fee Invoice",
		Taxes: []lago.Tax{
			{Code: "vat_20"},
			{Code: "local_tax"},
		},
		CreatedAt: now,
	}

	base := addOnResourceModel{}
	state, diags := mapAddOnToModel(addOn, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics mapping add-on to model: %+v", diags)
	}

	if state.ID.ValueString() != "setup_fee" {
		t.Errorf("expected ID %q, got %q", "setup_fee", state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.Name.ValueString() != "Setup Fee" {
		t.Errorf("expected Name %q, got %q", "Setup Fee", state.Name.ValueString())
	}
	if state.Code.ValueString() != "setup_fee" {
		t.Errorf("expected Code %q, got %q", "setup_fee", state.Code.ValueString())
	}
	if state.AmountCents.ValueInt64() != 10000 {
		t.Errorf("expected AmountCents 10000, got %d", state.AmountCents.ValueInt64())
	}
	if state.AmountCurrency.ValueString() != "EUR" {
		t.Errorf("expected AmountCurrency EUR, got %q", state.AmountCurrency.ValueString())
	}
	if state.Description.ValueString() != "A one-time setup fee" {
		t.Errorf("expected Description %q, got %q", "A one-time setup fee", state.Description.ValueString())
	}
	if state.InvoiceDisplayName.ValueString() != "Setup Fee Invoice" {
		t.Errorf("expected InvoiceDisplayName %q, got %q", "Setup Fee Invoice", state.InvoiceDisplayName.ValueString())
	}
	if state.TaxCodes.IsNull() {
		t.Fatal("expected non-null TaxCodes")
	}
	if state.CreatedAt.IsNull() {
		t.Fatal("expected non-null CreatedAt")
	}
}

func TestAddOnMapToModelEmptyOptionals(t *testing.T) {
	t.Parallel()

	addOn := &lago.AddOn{
		LagoID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:           "Basic Add-On",
		Code:           "basic_add_on",
		AmountCents:    100,
		AmountCurrency: lago.USD,
	}

	base := addOnResourceModel{}
	state, diags := mapAddOnToModel(addOn, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if !state.Description.IsNull() {
		t.Errorf("expected null Description, got %q", state.Description.ValueString())
	}
	if !state.InvoiceDisplayName.IsNull() {
		t.Errorf("expected null InvoiceDisplayName, got %q", state.InvoiceDisplayName.ValueString())
	}
	if !state.TaxCodes.IsNull() {
		t.Errorf("expected null TaxCodes when taxes are empty")
	}
	if !state.CreatedAt.IsNull() {
		t.Errorf("expected null CreatedAt for zero time, got %q", state.CreatedAt.ValueString())
	}
}

func TestAccAddOnResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	randomCode := fmt.Sprintf("tf_acc_addon_%d", time.Now().UnixNano())
	resourceName := "lago_add_on.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccAddOnConfig(randomCode, "Initial Add-On", 5000, "USD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "code", randomCode),
					resource.TestCheckResourceAttr(resourceName, "name", "Initial Add-On"),
					resource.TestCheckResourceAttr(resourceName, "amount_cents", "5000"),
					resource.TestCheckResourceAttr(resourceName, "amount_currency", "USD"),
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
				Config: testAccAddOnConfig(randomCode, "Updated Add-On", 9999, "EUR"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Updated Add-On"),
					resource.TestCheckResourceAttr(resourceName, "amount_cents", "9999"),
					resource.TestCheckResourceAttr(resourceName, "amount_currency", "EUR"),
				),
			},
		},
	})
}

func testAccAddOnConfig(code, name string, amountCents int, currency string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_add_on" "test" {
  name            = "%s"
  code            = "%s"
  amount_cents    = %d
  amount_currency = "%s"
  description     = "Terraform acceptance test add-on"
}
`, name, code, amountCents, currency)
}
