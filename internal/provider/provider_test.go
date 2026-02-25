package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// providerConfig returns the lago provider block for use in acceptance test configs.
func providerConfig() string {
	return `
provider "lago" {
  api_endpoint = "` + os.Getenv("LAGO_API_ENDPOINT") + `"
  api_key      = "` + os.Getenv("LAGO_API_KEY") + `"
}
`
}

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	p := New("test")()
	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "lago" {
		t.Fatalf("expected type name lago, got %q", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Fatalf("expected version test, got %q", resp.Version)
	}
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	p := New("test")()
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["api_endpoint"]; !ok {
		t.Fatal("expected api_endpoint attribute")
	}
	if _, ok := resp.Schema.Attributes["api_key"]; !ok {
		t.Fatal("expected api_key attribute")
	}
}
