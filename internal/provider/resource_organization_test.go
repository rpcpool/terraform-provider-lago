package provider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Unit tests — expandOrganizationInput
// ---------------------------------------------------------------------------

func TestOrganizationExpandInput_Basic(t *testing.T) {
	t.Parallel()

	plan := organizationResourceModel{
		ID:                        types.StringValue(organizationSingletonID),
		Name:                      types.StringValue("Acme Corp"),
		Email:                     types.StringValue("billing@acme.example.com"),
		AddressLine1:              types.StringValue("123 Main St"),
		AddressLine2:              types.StringNull(),
		City:                      types.StringValue("New York"),
		State:                     types.StringValue("NY"),
		Zipcode:                   types.StringValue("10001"),
		Country:                   types.StringValue("US"),
		DefaultCurrency:           types.StringValue("USD"),
		Timezone:                  types.StringValue("America/New_York"),
		LegalName:                 types.StringValue("Acme Corporation Ltd."),
		LegalNumber:               types.StringValue("123456789"),
		TaxIdentificationNumber:   types.StringValue("US123456789"),
		NetPaymentTerm:            types.Int64Value(30),
		DocumentNumbering:         types.StringValue("per_organization"),
		DocumentNumberPrefix:      types.StringValue("ACME"),
		FinalizeZeroAmountInvoice: types.BoolValue(true),
		EmailSettings:             types.SetNull(types.StringType),
		BillingConfiguration:      types.ObjectNull(organizationBillingConfigObjectType().AttrTypes),
		CreatedAt:                 types.StringNull(),
	}

	input, diags := expandOrganizationInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if input.Name != "Acme Corp" {
		t.Errorf("expected Name %q, got %q", "Acme Corp", input.Name)
	}
	if input.Email != "billing@acme.example.com" {
		t.Errorf("expected Email %q, got %q", "billing@acme.example.com", input.Email)
	}
	if input.AddressLine1 != "123 Main St" {
		t.Errorf("expected AddressLine1 %q, got %q", "123 Main St", input.AddressLine1)
	}
	if input.AddressLine2 != "" {
		t.Errorf("expected empty AddressLine2 for null input, got %q", input.AddressLine2)
	}
	if input.City != "New York" {
		t.Errorf("expected City %q, got %q", "New York", input.City)
	}
	if input.State != "NY" {
		t.Errorf("expected State %q, got %q", "NY", input.State)
	}
	if input.Zipcode != "10001" {
		t.Errorf("expected Zipcode %q, got %q", "10001", input.Zipcode)
	}
	if input.Country != "US" {
		t.Errorf("expected Country %q, got %q", "US", input.Country)
	}
	if string(input.DefaultCurrency) != "USD" {
		t.Errorf("expected DefaultCurrency %q, got %q", "USD", input.DefaultCurrency)
	}
	if input.Timezone != "America/New_York" {
		t.Errorf("expected Timezone %q, got %q", "America/New_York", input.Timezone)
	}
	if input.LegalName != "Acme Corporation Ltd." {
		t.Errorf("expected LegalName %q, got %q", "Acme Corporation Ltd.", input.LegalName)
	}
	if input.LegalNumber != "123456789" {
		t.Errorf("expected LegalNumber %q, got %q", "123456789", input.LegalNumber)
	}
	if input.TaxIdentificationNumber != "US123456789" {
		t.Errorf("expected TaxIdentificationNumber %q, got %q", "US123456789", input.TaxIdentificationNumber)
	}
	if input.NetPaymentTerm != 30 {
		t.Errorf("expected NetPaymentTerm 30, got %d", input.NetPaymentTerm)
	}
	if string(input.DocumentNumbering) != "per_organization" {
		t.Errorf("expected DocumentNumbering %q, got %q", "per_organization", input.DocumentNumbering)
	}
	if input.DocumentNumberPrefix != "ACME" {
		t.Errorf("expected DocumentNumberPrefix %q, got %q", "ACME", input.DocumentNumberPrefix)
	}
	if !input.FinalizeZeroAmountInvoice {
		t.Error("expected FinalizeZeroAmountInvoice true, got false")
	}
	if len(input.EmailSettings) != 0 {
		t.Errorf("expected empty EmailSettings for null input, got %v", input.EmailSettings)
	}
}

func TestOrganizationExpandInput_NullOptionals(t *testing.T) {
	t.Parallel()

	plan := organizationResourceModel{
		ID:                        types.StringValue(organizationSingletonID),
		Name:                      types.StringNull(),
		Email:                     types.StringNull(),
		AddressLine1:              types.StringNull(),
		AddressLine2:              types.StringNull(),
		City:                      types.StringNull(),
		State:                     types.StringNull(),
		Zipcode:                   types.StringNull(),
		Country:                   types.StringNull(),
		DefaultCurrency:           types.StringNull(),
		Timezone:                  types.StringNull(),
		LegalName:                 types.StringNull(),
		LegalNumber:               types.StringNull(),
		TaxIdentificationNumber:   types.StringNull(),
		NetPaymentTerm:            types.Int64Null(),
		DocumentNumbering:         types.StringNull(),
		DocumentNumberPrefix:      types.StringNull(),
		FinalizeZeroAmountInvoice: types.BoolValue(false),
		EmailSettings:             types.SetNull(types.StringType),
		BillingConfiguration:      types.ObjectNull(organizationBillingConfigObjectType().AttrTypes),
		CreatedAt:                 types.StringNull(),
	}

	input, diags := expandOrganizationInput(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if input.Name != "" {
		t.Errorf("expected empty Name for null input, got %q", input.Name)
	}
	if input.NetPaymentTerm != 0 {
		t.Errorf("expected NetPaymentTerm 0 for null input, got %d", input.NetPaymentTerm)
	}
	if len(input.EmailSettings) != 0 {
		t.Errorf("expected empty EmailSettings for null input, got %v", input.EmailSettings)
	}
}

func TestOrganizationExpandInput_BillingConfiguration(t *testing.T) {
	t.Parallel()

	bcObj, diags := types.ObjectValue(
		organizationBillingConfigObjectType().AttrTypes,
		map[string]attr.Value{
			"invoice_grace_period": types.Int64Value(5),
			"invoice_footer":       types.StringValue("Thank you"),
			"document_locale":      types.StringValue("en"),
		},
	)
	if diags.HasError() {
		t.Fatalf("failed to build billing config object: %v", diags)
	}

	plan := organizationResourceModel{
		ID:                        types.StringValue(organizationSingletonID),
		Name:                      types.StringNull(),
		Email:                     types.StringNull(),
		AddressLine1:              types.StringNull(),
		AddressLine2:              types.StringNull(),
		City:                      types.StringNull(),
		State:                     types.StringNull(),
		Zipcode:                   types.StringNull(),
		Country:                   types.StringNull(),
		DefaultCurrency:           types.StringNull(),
		Timezone:                  types.StringNull(),
		LegalName:                 types.StringNull(),
		LegalNumber:               types.StringNull(),
		TaxIdentificationNumber:   types.StringNull(),
		NetPaymentTerm:            types.Int64Null(),
		DocumentNumbering:         types.StringNull(),
		DocumentNumberPrefix:      types.StringNull(),
		FinalizeZeroAmountInvoice: types.BoolValue(false),
		EmailSettings:             types.SetNull(types.StringType),
		BillingConfiguration:      bcObj,
		CreatedAt:                 types.StringNull(),
	}

	input, d := expandOrganizationInput(context.Background(), plan)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	if input.BillingConfiguration.InvoiceGracePeriod != 5 {
		t.Errorf("expected InvoiceGracePeriod 5, got %d", input.BillingConfiguration.InvoiceGracePeriod)
	}
	if input.BillingConfiguration.InvoiceFooter != "Thank you" {
		t.Errorf("expected InvoiceFooter %q, got %q", "Thank you", input.BillingConfiguration.InvoiceFooter)
	}
	if input.BillingConfiguration.DocumentLocale != "en" {
		t.Errorf("expected DocumentLocale %q, got %q", "en", input.BillingConfiguration.DocumentLocale)
	}
}

// ---------------------------------------------------------------------------
// Unit tests — mapOrganizationToModel
// ---------------------------------------------------------------------------

func TestOrganizationMapToModel_Full(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	org := &lago.Organization{
		Name:                      "Acme Corp",
		Email:                     "billing@acme.example.com",
		AddressLine1:              "123 Main St",
		AddressLine2:              "Suite 400",
		City:                      "New York",
		State:                     "NY",
		Zipcode:                   "10001",
		Country:                   "US",
		DefaultCurrency:           lago.Currency("USD"),
		Timezone:                  "America/New_York",
		LegalName:                 "Acme Corporation Ltd.",
		LegalNumber:               "123456789",
		TaxIdentificationNumber:   "US123456789",
		NetPaymentTerm:            30,
		DocumentNumbering:         lago.DocumentNumberingPerOrganization,
		DocumentNumberPrefix:      "ACME",
		FinalizeZeroAmountInvoice: true,
		EmailSettings:             []string{"invoice.finalized", "credit_note.created"},
		BillingConfiguration: lago.OrganizationBillingConfiguration{
			InvoiceGracePeriod: 3,
			InvoiceFooter:      "Thank you for your business.",
			DocumentLocale:     "en",
		},
		CreatedAt: now,
	}

	base := organizationResourceModel{}
	state, diags := mapOrganizationToModel(org, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != organizationSingletonID {
		t.Errorf("expected ID %q, got %q", organizationSingletonID, state.ID.ValueString())
	}
	if state.Name.ValueString() != "Acme Corp" {
		t.Errorf("expected Name %q, got %q", "Acme Corp", state.Name.ValueString())
	}
	if state.Email.ValueString() != "billing@acme.example.com" {
		t.Errorf("expected Email %q, got %q", "billing@acme.example.com", state.Email.ValueString())
	}
	if state.AddressLine1.ValueString() != "123 Main St" {
		t.Errorf("expected AddressLine1 %q, got %q", "123 Main St", state.AddressLine1.ValueString())
	}
	if state.AddressLine2.ValueString() != "Suite 400" {
		t.Errorf("expected AddressLine2 %q, got %q", "Suite 400", state.AddressLine2.ValueString())
	}
	if state.DefaultCurrency.ValueString() != "USD" {
		t.Errorf("expected DefaultCurrency %q, got %q", "USD", state.DefaultCurrency.ValueString())
	}
	if state.Timezone.ValueString() != "America/New_York" {
		t.Errorf("expected Timezone %q, got %q", "America/New_York", state.Timezone.ValueString())
	}
	if state.NetPaymentTerm.ValueInt64() != 30 {
		t.Errorf("expected NetPaymentTerm 30, got %d", state.NetPaymentTerm.ValueInt64())
	}
	if state.DocumentNumbering.ValueString() != "per_organization" {
		t.Errorf("expected DocumentNumbering %q, got %q", "per_organization", state.DocumentNumbering.ValueString())
	}
	if state.DocumentNumberPrefix.ValueString() != "ACME" {
		t.Errorf("expected DocumentNumberPrefix %q, got %q", "ACME", state.DocumentNumberPrefix.ValueString())
	}
	if !state.FinalizeZeroAmountInvoice.ValueBool() {
		t.Error("expected FinalizeZeroAmountInvoice true, got false")
	}
	if state.EmailSettings.IsNull() {
		t.Fatal("expected non-null EmailSettings")
	}
	if state.BillingConfiguration.IsNull() {
		t.Fatal("expected non-null BillingConfiguration")
	}
	if state.CreatedAt.IsNull() {
		t.Fatal("expected non-null CreatedAt")
	}
}

func TestOrganizationMapToModel_EmptyOptionals(t *testing.T) {
	t.Parallel()

	org := &lago.Organization{
		Name: "Minimal Org",
	}

	base := organizationResourceModel{}
	state, diags := mapOrganizationToModel(org, base)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != organizationSingletonID {
		t.Errorf("expected ID %q, got %q", organizationSingletonID, state.ID.ValueString())
	}
	if state.Name.ValueString() != "Minimal Org" {
		t.Errorf("expected Name %q, got %q", "Minimal Org", state.Name.ValueString())
	}
	if !state.Email.IsNull() {
		t.Errorf("expected null Email for empty string, got %q", state.Email.ValueString())
	}
	if !state.AddressLine1.IsNull() {
		t.Errorf("expected null AddressLine1 for empty string, got %q", state.AddressLine1.ValueString())
	}
	if !state.DefaultCurrency.IsNull() {
		t.Errorf("expected null DefaultCurrency for empty string, got %q", state.DefaultCurrency.ValueString())
	}
	if !state.NetPaymentTerm.IsNull() {
		t.Errorf("expected null NetPaymentTerm for zero value, got %d", state.NetPaymentTerm.ValueInt64())
	}
	if !state.EmailSettings.IsNull() {
		t.Error("expected null EmailSettings for empty slice")
	}
	if !state.BillingConfiguration.IsNull() {
		t.Error("expected null BillingConfiguration when all billing fields are zero")
	}
	if !state.CreatedAt.IsNull() {
		t.Errorf("expected null CreatedAt for zero time, got %q", state.CreatedAt.ValueString())
	}
	if state.FinalizeZeroAmountInvoice.ValueBool() {
		t.Error("expected FinalizeZeroAmountInvoice false for zero value")
	}
}

func TestOrganizationMapToModel_BillingConfigPreservedWhenPlanHasIt(t *testing.T) {
	t.Parallel()

	// Simulate the case where the user configured billing_configuration but the API
	// returns all-zero values (e.g. after create before the values propagate).
	existingBCObj, diags := types.ObjectValue(
		organizationBillingConfigObjectType().AttrTypes,
		map[string]attr.Value{
			"invoice_grace_period": types.Int64Value(3),
			"invoice_footer":       types.StringValue("Footer"),
			"document_locale":      types.StringValue("en"),
		},
	)
	if diags.HasError() {
		t.Fatalf("failed to build billing config object: %v", diags)
	}

	base := organizationResourceModel{
		BillingConfiguration: existingBCObj,
	}

	org := &lago.Organization{
		Name: "Org",
		// BillingConfiguration is zero — all empty
	}

	state, d := mapOrganizationToModel(org, base)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %v", d)
	}

	// Because the base had a non-null billing_configuration, the state should
	// preserve a non-null object (with null field values) rather than becoming null.
	if state.BillingConfiguration.IsNull() {
		t.Error("expected BillingConfiguration to be non-null when plan had it configured")
	}
}

func TestOrganizationMapToModel_SingletonID(t *testing.T) {
	t.Parallel()

	org := &lago.Organization{Name: "Test"}
	state, diags := mapOrganizationToModel(org, organizationResourceModel{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.ID.ValueString() != "organization" {
		t.Errorf("singleton ID must always be %q, got %q", "organization", state.ID.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Acceptance tests
// ---------------------------------------------------------------------------

func TestAccOrganizationResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	resourceName := "lago_organization.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create (upsert) with a minimal configuration.
			{
				Config: testAccOrganizationConfig("USD", "UTC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "organization"),
					resource.TestCheckResourceAttr(resourceName, "default_currency", "USD"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			// Import — import ID is ignored; state is read from the API.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false, // computed fields may differ; singleton import is structural
				ImportStateId:     "organization",
			},
			// Update — change the currency.
			// Note: timezone is sent to the API but Lago v1.42.0 ignores it and
			// always returns "UTC", so we only assert on default_currency here.
			{
				Config: testAccOrganizationConfig("EUR", "UTC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "organization"),
					resource.TestCheckResourceAttr(resourceName, "default_currency", "EUR"),
				),
			},
		},
	})
}

func TestAccOrganizationResource_BillingConfiguration(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	resourceName := "lago_organization.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationBillingConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "organization"),
					resource.TestCheckResourceAttr(resourceName, "billing_configuration.invoice_footer", "Thank you for your business."),
					resource.TestCheckResourceAttr(resourceName, "billing_configuration.document_locale", "en"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers
// ---------------------------------------------------------------------------

func testAccOrganizationConfig(currency, timezone string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_organization" "this" {
  default_currency = "%s"
  timezone         = "%s"
}
`, currency, timezone)
}

func testAccOrganizationBillingConfig() string {
	return providerConfig() + `
resource "lago_organization" "this" {
  billing_configuration = {
    invoice_footer  = "Thank you for your business."
    document_locale = "en"
  }
}
`
}
