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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Unit tests
// ---------------------------------------------------------------------------

func TestCustomerExpandInput_BasicFields(t *testing.T) {
	t.Parallel()

	plan := customerResourceModel{
		ID:                        types.StringNull(),
		LagoID:                    types.StringNull(),
		ExternalID:                types.StringValue("cust-001"),
		Name:                      types.StringValue("Acme Corp"),
		Email:                     types.StringValue("billing@acme.com"),
		Phone:                     types.StringValue("+1-555-000-0001"),
		URL:                       types.StringValue("https://acme.example.com"),
		CustomerType:              types.StringValue("company"),
		Currency:                  types.StringValue("USD"),
		Timezone:                  types.StringValue("America/New_York"),
		AddressLine1:              types.StringValue("123 Main St"),
		AddressLine2:              types.StringValue("Suite 1"),
		City:                      types.StringValue("New York"),
		State:                     types.StringValue("NY"),
		Zipcode:                   types.StringValue("10001"),
		Country:                   types.StringValue("US"),
		LegalName:                 types.StringValue("Acme Corporation Inc."),
		LegalNumber:               types.StringValue("12-3456789"),
		TaxIdentificationNumber:   types.StringValue("US123456789"),
		NetPaymentTerm:            types.Int64Value(30),
		FinalizeZeroAmountInvoice: types.StringValue("skip"),
		TaxCodes:                  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("vat_20")}),
		BillingConfiguration:      types.ObjectNull(customerBillingConfigObjectType().AttrTypes),
		ShippingAddress:           types.ObjectNull(customerShippingAddressObjectType().AttrTypes),
		Metadata:                  types.ListNull(customerMetadataObjectType()),
		CreatedAt:                 types.StringNull(),
		UpdatedAt:                 types.StringNull(),
	}

	input, diags := expandCustomerInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.ExternalID != "cust-001" {
		t.Errorf("expected ExternalID %q, got %q", "cust-001", input.ExternalID)
	}
	if input.Name != "Acme Corp" {
		t.Errorf("expected Name %q, got %q", "Acme Corp", input.Name)
	}
	if input.Email != "billing@acme.com" {
		t.Errorf("expected Email %q, got %q", "billing@acme.com", input.Email)
	}
	if input.Phone != "+1-555-000-0001" {
		t.Errorf("expected Phone %q, got %q", "+1-555-000-0001", input.Phone)
	}
	if input.URL != "https://acme.example.com" {
		t.Errorf("expected URL %q, got %q", "https://acme.example.com", input.URL)
	}
	if string(input.CustomerType) != "company" {
		t.Errorf("expected CustomerType %q, got %q", "company", input.CustomerType)
	}
	if string(input.Currency) != "USD" {
		t.Errorf("expected Currency %q, got %q", "USD", input.Currency)
	}
	if input.Timezone != "America/New_York" {
		t.Errorf("expected Timezone %q, got %q", "America/New_York", input.Timezone)
	}
	if input.AddressLine1 != "123 Main St" {
		t.Errorf("expected AddressLine1 %q, got %q", "123 Main St", input.AddressLine1)
	}
	if input.City != "New York" {
		t.Errorf("expected City %q, got %q", "New York", input.City)
	}
	if input.Country != "US" {
		t.Errorf("expected Country %q, got %q", "US", input.Country)
	}
	if input.LegalName != "Acme Corporation Inc." {
		t.Errorf("expected LegalName %q, got %q", "Acme Corporation Inc.", input.LegalName)
	}
	if input.NetPaymentTerm != 30 {
		t.Errorf("expected NetPaymentTerm 30, got %d", input.NetPaymentTerm)
	}
	if string(input.FinalizeZeroAmountInvoice) != "skip" {
		t.Errorf("expected FinalizeZeroAmountInvoice %q, got %q", "skip", input.FinalizeZeroAmountInvoice)
	}
	if len(input.TaxCodes) != 1 || input.TaxCodes[0] != "vat_20" {
		t.Errorf("expected TaxCodes [vat_20], got %v", input.TaxCodes)
	}
}

func TestCustomerExpandInput_NullOptionals(t *testing.T) {
	t.Parallel()

	plan := customerResourceModel{
		ID:                        types.StringNull(),
		LagoID:                    types.StringNull(),
		ExternalID:                types.StringValue("min-cust"),
		Name:                      types.StringNull(),
		Email:                     types.StringNull(),
		Phone:                     types.StringNull(),
		URL:                       types.StringNull(),
		CustomerType:              types.StringNull(),
		Currency:                  types.StringNull(),
		Timezone:                  types.StringNull(),
		AddressLine1:              types.StringNull(),
		AddressLine2:              types.StringNull(),
		City:                      types.StringNull(),
		State:                     types.StringNull(),
		Zipcode:                   types.StringNull(),
		Country:                   types.StringNull(),
		LegalName:                 types.StringNull(),
		LegalNumber:               types.StringNull(),
		TaxIdentificationNumber:   types.StringNull(),
		NetPaymentTerm:            types.Int64Null(),
		FinalizeZeroAmountInvoice: types.StringNull(),
		TaxCodes:                  types.SetNull(types.StringType),
		BillingConfiguration:      types.ObjectNull(customerBillingConfigObjectType().AttrTypes),
		ShippingAddress:           types.ObjectNull(customerShippingAddressObjectType().AttrTypes),
		Metadata:                  types.ListNull(customerMetadataObjectType()),
		CreatedAt:                 types.StringNull(),
		UpdatedAt:                 types.StringNull(),
	}

	input, diags := expandCustomerInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if input.ExternalID != "min-cust" {
		t.Errorf("expected ExternalID %q, got %q", "min-cust", input.ExternalID)
	}
	if input.Name != "" {
		t.Errorf("expected empty Name for null input, got %q", input.Name)
	}
	if input.NetPaymentTerm != 0 {
		t.Errorf("expected NetPaymentTerm 0 for null, got %d", input.NetPaymentTerm)
	}
	if len(input.TaxCodes) != 0 {
		t.Errorf("expected no tax codes, got %v", input.TaxCodes)
	}
}

func TestCustomerExpandBillingConfiguration(t *testing.T) {
	t.Parallel()

	billingObj := types.ObjectValueMust(customerBillingConfigObjectType().AttrTypes, map[string]attr.Value{
		"invoice_grace_period":     types.Int64Value(3),
		"payment_provider":         types.StringValue("stripe"),
		"payment_provider_code":    types.StringValue("stripe_us"),
		"provider_customer_id":     types.StringValue("cus_abc123"),
		"sync_with_provider":       types.BoolValue(true),
		"document_locale":          types.StringValue("en"),
		"provider_payment_methods": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("card"), types.StringValue("us_bank_account")}),
	})

	cfg, diags := expandCustomerBillingConfiguration(context.Background(), billingObj)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if cfg.InvoiceGracePeriod != 3 {
		t.Errorf("expected InvoiceGracePeriod 3, got %d", cfg.InvoiceGracePeriod)
	}
	if string(cfg.PaymentProvider) != "stripe" {
		t.Errorf("expected PaymentProvider %q, got %q", "stripe", cfg.PaymentProvider)
	}
	if cfg.PaymentProviderCode != "stripe_us" {
		t.Errorf("expected PaymentProviderCode %q, got %q", "stripe_us", cfg.PaymentProviderCode)
	}
	if cfg.ProviderCustomerID != "cus_abc123" {
		t.Errorf("expected ProviderCustomerID %q, got %q", "cus_abc123", cfg.ProviderCustomerID)
	}
	if cfg.SyncWithProvider != true {
		t.Errorf("expected SyncWithProvider true")
	}
	if cfg.DocumentLocale != "en" {
		t.Errorf("expected DocumentLocale %q, got %q", "en", cfg.DocumentLocale)
	}
	if len(cfg.ProviderPaymentMethods) != 2 {
		t.Errorf("expected 2 provider payment methods, got %d", len(cfg.ProviderPaymentMethods))
	}
}

func TestCustomerExpandBillingConfiguration_Null(t *testing.T) {
	t.Parallel()

	cfg, diags := expandCustomerBillingConfiguration(context.Background(), types.ObjectNull(customerBillingConfigObjectType().AttrTypes))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	// All zero values expected when null
	if string(cfg.PaymentProvider) != "" {
		t.Errorf("expected empty PaymentProvider for null config, got %q", cfg.PaymentProvider)
	}
}

func TestCustomerExpandShippingAddress(t *testing.T) {
	t.Parallel()

	addrObj := types.ObjectValueMust(customerShippingAddressObjectType().AttrTypes, map[string]attr.Value{
		"address_line1": types.StringValue("456 Warehouse Blvd"),
		"address_line2": types.StringNull(),
		"city":          types.StringValue("Newark"),
		"state":         types.StringValue("NJ"),
		"zipcode":       types.StringValue("07101"),
		"country":       types.StringValue("US"),
	})

	addr, diags := expandCustomerShippingAddress(context.Background(), addrObj)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if addr.AddressLine1 != "456 Warehouse Blvd" {
		t.Errorf("expected AddressLine1 %q, got %q", "456 Warehouse Blvd", addr.AddressLine1)
	}
	if addr.City != "Newark" {
		t.Errorf("expected City %q, got %q", "Newark", addr.City)
	}
	if addr.Country != "US" {
		t.Errorf("expected Country %q, got %q", "US", addr.Country)
	}
}

func TestCustomerExpandMetadata(t *testing.T) {
	t.Parallel()

	metaList := types.ListValueMust(customerMetadataObjectType(), []attr.Value{
		types.ObjectValueMust(customerMetadataObjectType().AttrTypes, map[string]attr.Value{
			"key":                types.StringValue("salesforce_id"),
			"value":              types.StringValue("0015000000XXXXAAA"),
			"display_in_invoice": types.BoolValue(false),
		}),
		types.ObjectValueMust(customerMetadataObjectType().AttrTypes, map[string]attr.Value{
			"key":                types.StringValue("plan_tier"),
			"value":              types.StringValue("enterprise"),
			"display_in_invoice": types.BoolValue(true),
		}),
	})

	metadata, diags := expandCustomerMetadata(context.Background(), metaList)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if len(metadata) != 2 {
		t.Fatalf("expected 2 metadata items, got %d", len(metadata))
	}
	if metadata[0].Key != "salesforce_id" {
		t.Errorf("expected Key %q, got %q", "salesforce_id", metadata[0].Key)
	}
	if metadata[0].Value != "0015000000XXXXAAA" {
		t.Errorf("expected Value %q, got %q", "0015000000XXXXAAA", metadata[0].Value)
	}
	if metadata[0].DisplayInInvoice != false {
		t.Errorf("expected DisplayInInvoice false for first item")
	}
	if metadata[1].DisplayInInvoice != true {
		t.Errorf("expected DisplayInInvoice true for second item")
	}
}

func TestCustomerMapToModel(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	now := time.Now().UTC()

	customer := &lago.Customer{
		LagoID:                    lagoID,
		ExternalID:                "cust-001",
		Name:                      "Acme Corp",
		Email:                     "billing@acme.com",
		Phone:                     "+1-555-000-0001",
		URL:                       "https://acme.example.com",
		CustomerType:              "company",
		Currency:                  lago.USD,
		Timezone:                  "America/New_York",
		AddressLine1:              "123 Main St",
		AddressLine2:              "Suite 1",
		City:                      "New York",
		State:                     "NY",
		Zipcode:                   "10001",
		Country:                   "US",
		LegalName:                 "Acme Corporation Inc.",
		LegalNumber:               "12-3456789",
		TaxIdentificationNumber:   "US123456789",
		NetPaymentTerm:            30,
		FinalizeZeroAmountInvoice: lago.SkipInvoice,
		Taxes: []lago.Tax{
			{Code: "vat_20"},
		},
		BillingConfiguration: lago.CustomerBillingConfiguration{
			InvoiceGracePeriod:  3,
			PaymentProvider:     lago.PaymentProviderStripe,
			PaymentProviderCode: "stripe_us",
			ProviderCustomerID:  "cus_abc123",
			DocumentLocale:      "en",
		},
		ShippingAddress: lago.Address{
			AddressLine1: "456 Warehouse Blvd",
			City:         "Newark",
			State:        "NJ",
			Zipcode:      "07101",
			Country:      "US",
		},
		Metadata: []lago.MetadataResponse{
			{Key: "salesforce_id", Value: "0015000000XXXXAAA", DisplayInInvoice: false},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	base := customerResourceModel{
		BillingConfiguration: types.ObjectNull(customerBillingConfigObjectType().AttrTypes),
		ShippingAddress:      types.ObjectNull(customerShippingAddressObjectType().AttrTypes),
		Metadata:             types.ListNull(customerMetadataObjectType()),
		TaxCodes:             types.SetNull(types.StringType),
	}

	state, diags := mapCustomerToModel(context.Background(), customer, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if state.ID.ValueString() != "cust-001" {
		t.Errorf("expected ID %q, got %q", "cust-001", state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.ExternalID.ValueString() != "cust-001" {
		t.Errorf("expected ExternalID %q, got %q", "cust-001", state.ExternalID.ValueString())
	}
	if state.Name.ValueString() != "Acme Corp" {
		t.Errorf("expected Name %q, got %q", "Acme Corp", state.Name.ValueString())
	}
	if state.Email.ValueString() != "billing@acme.com" {
		t.Errorf("expected Email %q, got %q", "billing@acme.com", state.Email.ValueString())
	}
	if state.NetPaymentTerm.ValueInt64() != 30 {
		t.Errorf("expected NetPaymentTerm 30, got %d", state.NetPaymentTerm.ValueInt64())
	}
	if state.FinalizeZeroAmountInvoice.ValueString() != "skip" {
		t.Errorf("expected FinalizeZeroAmountInvoice %q, got %q", "skip", state.FinalizeZeroAmountInvoice.ValueString())
	}
	if state.TaxCodes.IsNull() {
		t.Fatal("expected non-null TaxCodes")
	}
	if state.BillingConfiguration.IsNull() {
		t.Fatal("expected non-null BillingConfiguration")
	}
	if state.ShippingAddress.IsNull() {
		t.Fatal("expected non-null ShippingAddress")
	}
	if state.Metadata.IsNull() {
		t.Fatal("expected non-null Metadata")
	}
	if state.CreatedAt.IsNull() {
		t.Fatal("expected non-null CreatedAt")
	}
	if state.UpdatedAt.IsNull() {
		t.Fatal("expected non-null UpdatedAt")
	}
}

func TestCustomerMapToModel_EmptyOptionals(t *testing.T) {
	t.Parallel()

	customer := &lago.Customer{
		LagoID:     uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		ExternalID: "minimal-cust",
	}

	base := customerResourceModel{
		BillingConfiguration: types.ObjectNull(customerBillingConfigObjectType().AttrTypes),
		ShippingAddress:      types.ObjectNull(customerShippingAddressObjectType().AttrTypes),
		Metadata:             types.ListNull(customerMetadataObjectType()),
		TaxCodes:             types.SetNull(types.StringType),
	}

	state, diags := mapCustomerToModel(context.Background(), customer, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}

	if !state.Name.IsNull() {
		t.Errorf("expected null Name, got %q", state.Name.ValueString())
	}
	if !state.Email.IsNull() {
		t.Errorf("expected null Email, got %q", state.Email.ValueString())
	}
	if !state.NetPaymentTerm.IsNull() {
		t.Errorf("expected null NetPaymentTerm, got %d", state.NetPaymentTerm.ValueInt64())
	}
	if !state.FinalizeZeroAmountInvoice.IsNull() {
		t.Errorf("expected null FinalizeZeroAmountInvoice, got %q", state.FinalizeZeroAmountInvoice.ValueString())
	}
	if !state.TaxCodes.IsNull() {
		t.Errorf("expected null TaxCodes when no taxes")
	}
	if !state.BillingConfiguration.IsNull() {
		t.Errorf("expected null BillingConfiguration when empty API response")
	}
	if !state.ShippingAddress.IsNull() {
		t.Errorf("expected null ShippingAddress when empty API response")
	}
	if !state.Metadata.IsNull() {
		t.Errorf("expected null Metadata when empty API response")
	}
	if !state.CreatedAt.IsNull() {
		t.Errorf("expected null CreatedAt for zero time, got %q", state.CreatedAt.ValueString())
	}
}

func TestCustomerFlattenBillingConfiguration_SyncWithProviderPreserved(t *testing.T) {
	t.Parallel()

	// The base plan has sync_with_provider=true; the API does not echo it back.
	// It should be preserved from the base into state.
	baseBilling := types.ObjectValueMust(customerBillingConfigObjectType().AttrTypes, map[string]attr.Value{
		"invoice_grace_period":     types.Int64Value(0),
		"payment_provider":         types.StringNull(),
		"payment_provider_code":    types.StringNull(),
		"provider_customer_id":     types.StringNull(),
		"sync_with_provider":       types.BoolValue(true),
		"document_locale":          types.StringNull(),
		"provider_payment_methods": types.SetNull(types.StringType),
	})

	// API response has payment_provider so the object won't be null
	apiCfg := lago.CustomerBillingConfiguration{
		PaymentProvider: lago.PaymentProviderStripe,
	}

	obj, diags := flattenCustomerBillingConfiguration(context.Background(), apiCfg, baseBilling)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if obj.IsNull() {
		t.Fatal("expected non-null billing configuration object")
	}

	var model customerBillingConfigModel
	diags = obj.As(context.Background(), &model, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics extracting model: %+v", diags)
	}
	if !model.SyncWithProvider.ValueBool() {
		t.Errorf("expected sync_with_provider to be preserved as true, got false")
	}
}

func TestCustomerFlattenShippingAddress_EmptyIsNull(t *testing.T) {
	t.Parallel()

	obj, diags := flattenCustomerShippingAddress(lago.Address{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if !obj.IsNull() {
		t.Error("expected null object for empty shipping address")
	}
}

func TestCustomerFlattenMetadata_EmptyIsNull(t *testing.T) {
	t.Parallel()

	list, diags := flattenCustomerMetadata(nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if !list.IsNull() {
		t.Error("expected null list for empty metadata")
	}
}

// ---------------------------------------------------------------------------
// Acceptance tests
// ---------------------------------------------------------------------------

func TestAccCustomerResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	randomID := fmt.Sprintf("tf_acc_cust_%d", time.Now().UnixNano())
	resourceName := "lago_customer.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerConfig(randomID, "Acme Corp", "billing@acme.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "external_id", randomID),
					resource.TestCheckResourceAttr(resourceName, "name", "Acme Corp"),
					resource.TestCheckResourceAttr(resourceName, "email", "billing@acme.example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "lago_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccCustomerConfig(randomID, "Acme Corporation", "updated@acme.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "Acme Corporation"),
					resource.TestCheckResourceAttr(resourceName, "email", "updated@acme.example.com"),
				),
			},
		},
	})
}

func TestAccCustomerResourceWithBillingConfig(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	randomID := fmt.Sprintf("tf_acc_cust_bc_%d", time.Now().UnixNano())
	resourceName := "lago_customer.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerWithBillingConfig(randomID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "external_id", randomID),
					resource.TestCheckResourceAttr(resourceName, "currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "net_payment_term", "30"),
					resource.TestCheckResourceAttr(resourceName, "billing_configuration.invoice_grace_period", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "lago_id"),
				),
			},
		},
	})
}

func testAccCustomerConfig(externalID, name, email string) string {
	return fmt.Sprintf(`
provider "lago" {
  api_endpoint = "%s"
  api_key      = "%s"
}

resource "lago_customer" "test" {
  external_id   = "%s"
  name          = "%s"
  email         = "%s"
  customer_type = "company"
  currency      = "USD"
}
`, os.Getenv("LAGO_API_ENDPOINT"), os.Getenv("LAGO_API_KEY"), externalID, name, email)
}

func testAccCustomerWithBillingConfig(externalID string) string {
	return fmt.Sprintf(`
provider "lago" {
  api_endpoint = "%s"
  api_key      = "%s"
}

resource "lago_customer" "test" {
  external_id      = "%s"
  name             = "Terraform Billing Config Test"
  email            = "test@example.com"
  customer_type    = "company"
  currency         = "USD"
  net_payment_term = 30

  billing_configuration {
    invoice_grace_period = 3
    document_locale      = "en"
  }
}
`, os.Getenv("LAGO_API_ENDPOINT"), os.Getenv("LAGO_API_KEY"), externalID)
}
