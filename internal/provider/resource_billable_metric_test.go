package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestValidateWeightedInterval(t *testing.T) {
	t.Parallel()

	diags := validateWeightedInterval(types.StringValue("sum_agg"), types.StringValue("daily"))
	if len(diags) == 0 {
		t.Fatal("expected diagnostics when weighted_interval is set for non-weighted aggregation")
	}

	diags = validateWeightedInterval(types.StringValue("weighted_sum_agg"), types.StringValue("daily"))
	if len(diags) > 0 {
		t.Fatalf("expected no diagnostics for weighted_sum_agg, got %d", len(diags))
	}
}

func TestFilterExpandFlatten(t *testing.T) {
	t.Parallel()

	objectType := filterObjectType()
	filterList := types.SetValueMust(objectType, []attr.Value{
		types.ObjectValueMust(objectType.AttrTypes, map[string]attr.Value{
			"key": types.StringValue("backend"),
			"values": types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("rpc"),
				types.StringValue("grpc"),
			}),
		}),
	})

	expanded, diags := expandFilters(context.Background(), filterList)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics expanding filters: %+v", diags)
	}

	if len(expanded) != 1 {
		t.Fatalf("expected one filter, got %d", len(expanded))
	}
	if expanded[0].Key != "backend" {
		t.Fatalf("expected backend key, got %q", expanded[0].Key)
	}

	flattened, flattenDiags := flattenFilters([]lago.BillableMetricFilter{
		{
			Key:    "historical",
			Values: []string{"true", "false"},
		},
	})
	if flattenDiags.HasError() {
		t.Fatalf("unexpected diagnostics flattening filters: %+v", flattenDiags)
	}
	if flattened.IsNull() || flattened.IsUnknown() {
		t.Fatal("expected non-null flattened filters")
	}
}

func TestAccBillableMetricResource(t *testing.T) {
	if os.Getenv("LAGO_ACC") != "1" {
		t.Skip("set LAGO_ACC=1 to run acceptance tests")
	}

	if os.Getenv("LAGO_API_ENDPOINT") == "" || os.Getenv("LAGO_API_KEY") == "" {
		t.Fatal("set LAGO_API_ENDPOINT and LAGO_API_KEY for acceptance tests")
	}

	randomCode := fmt.Sprintf("tf_acc_metric_%d", time.Now().UnixNano())
	resourceName := "lago_billable_metric.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lago": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccBillableMetricConfig(randomCode, "sum_agg", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "code", randomCode),
					resource.TestCheckResourceAttr(resourceName, "aggregation_type", "sum_agg"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
			{
				Config: testAccBillableMetricConfig(randomCode, "max_agg", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "aggregation_type", "max_agg"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
			{
				Config:      testAccInvalidWeightedIntervalConfig(randomCode),
				ExpectError: regexp.MustCompile("weighted_interval"),
			},
		},
	})
}

func testAccBillableMetricConfig(code, aggregationType, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_billable_metric" "test" {
  name             = "Terraform Acceptance Test"
  code             = "%s"
  aggregation_type = "%s"
  field_name       = "request_count"
  description      = "%s"
}
`, code, aggregationType, description)
}

func testAccInvalidWeightedIntervalConfig(code string) string {
	return providerConfig() + fmt.Sprintf(`
resource "lago_billable_metric" "test" {
  name              = "Terraform Acceptance Test"
  code              = "%s"
  aggregation_type  = "sum_agg"
  weighted_interval = "daily"
}
`, code)
}
