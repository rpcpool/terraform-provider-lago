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

func TestWebhookEndpointExpandInput(t *testing.T) {
	t.Parallel()

	plan := webhookEndpointResourceModel{
		ID:            types.StringNull(),
		LagoID:        types.StringNull(),
		WebhookURL:    types.StringValue("https://example.com/webhooks/lago"),
		SignatureAlgo: types.StringValue("hmac"),
		CreatedAt:     types.StringNull(),
	}

	input := expandWebhookEndpointInput(plan)

	if input.WebhookURL != "https://example.com/webhooks/lago" {
		t.Errorf("expected WebhookURL %q, got %q", "https://example.com/webhooks/lago", input.WebhookURL)
	}
	if input.SignatureAlgo != lago.HMac {
		t.Errorf("expected SignatureAlgo %q, got %q", lago.HMac, input.SignatureAlgo)
	}
}

func TestWebhookEndpointExpandInput_NullSignatureAlgo(t *testing.T) {
	t.Parallel()

	plan := webhookEndpointResourceModel{
		ID:            types.StringNull(),
		LagoID:        types.StringNull(),
		WebhookURL:    types.StringValue("https://example.com/hooks"),
		SignatureAlgo: types.StringNull(),
		CreatedAt:     types.StringNull(),
	}

	input := expandWebhookEndpointInput(plan)

	if input.WebhookURL != "https://example.com/hooks" {
		t.Errorf("expected WebhookURL %q, got %q", "https://example.com/hooks", input.WebhookURL)
	}
	if input.SignatureAlgo != "" {
		t.Errorf("expected empty SignatureAlgo, got %q", input.SignatureAlgo)
	}
}

func TestWebhookEndpointMapToModel(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	now := time.Now().UTC()

	endpoint := &lago.WebhookEndpoint{
		LagoID:        lagoID,
		WebhookURL:    "https://example.com/webhooks/lago",
		SignatureAlgo: lago.JWT,
		CreatedAt:     now,
	}

	base := webhookEndpointResourceModel{}
	state := mapWebhookEndpointToModel(endpoint, base)

	if state.ID.ValueString() != lagoID.String() {
		t.Errorf("expected ID %q, got %q", lagoID.String(), state.ID.ValueString())
	}
	if state.LagoID.ValueString() != lagoID.String() {
		t.Errorf("expected LagoID %q, got %q", lagoID.String(), state.LagoID.ValueString())
	}
	if state.WebhookURL.ValueString() != "https://example.com/webhooks/lago" {
		t.Errorf("expected WebhookURL %q, got %q", "https://example.com/webhooks/lago", state.WebhookURL.ValueString())
	}
	if state.SignatureAlgo.ValueString() != "jwt" {
		t.Errorf("expected SignatureAlgo %q, got %q", "jwt", state.SignatureAlgo.ValueString())
	}
	if state.CreatedAt.IsNull() {
		t.Error("expected non-null CreatedAt")
	}
}

func TestWebhookEndpointMapToModel_EmptySignatureAlgo(t *testing.T) {
	t.Parallel()

	lagoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	endpoint := &lago.WebhookEndpoint{
		LagoID:        lagoID,
		WebhookURL:    "https://example.com/hooks",
		SignatureAlgo: "",
	}

	base := webhookEndpointResourceModel{}
	state := mapWebhookEndpointToModel(endpoint, base)

	if !state.SignatureAlgo.IsNull() {
		t.Errorf("expected null SignatureAlgo, got %q", state.SignatureAlgo.ValueString())
	}
	if !state.CreatedAt.IsNull() {
		t.Errorf("expected null CreatedAt for zero time, got %q", state.CreatedAt.ValueString())
	}
}

func TestAccWebhookEndpointResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	resourceName := "lago_webhook_endpoint.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookEndpointConfig("https://example.com/webhooks/lago", "hmac"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "webhook_url", "https://example.com/webhooks/lago"),
					resource.TestCheckResourceAttr(resourceName, "signature_algo", "hmac"),
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
				Config: testAccWebhookEndpointConfig("https://example.com/webhooks/lago", "jwt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "signature_algo", "jwt"),
				),
			},
		},
	})
}

func testAccWebhookEndpointConfig(webhookURL, signatureAlgo string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_webhook_endpoint" "test" {
  webhook_url    = "%s"
  signature_algo = "%s"
}
`, webhookURL, signatureAlgo)
}
