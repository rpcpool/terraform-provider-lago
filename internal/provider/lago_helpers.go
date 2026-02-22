package provider

import (
	"net/http"

	lago "github.com/getlago/lago-go-client"
)

// isNotFound returns true when the Lago API returned a 404 Not Found.
// Used in Read() to remove resources that no longer exist remotely.
func isNotFound(err *lago.Error) bool {
	return err != nil && err.HTTPStatusCode == http.StatusNotFound
}
