package provider

import (
	"net/http"
	"time"

	lago "github.com/getlago/lago-go-client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// isNotFound returns true when the Lago API returned a 404 Not Found.
// Used in Read() to remove resources that no longer exist remotely.
func isNotFound(err *lago.Error) bool {
	return err != nil && err.HTTPStatusCode == http.StatusNotFound
}

// timeOrNull converts a *time.Time to a types.String (RFC3339) or types.StringNull if nil or zero.
func timeOrNull(t *time.Time) types.String {
	if t == nil || t.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}
