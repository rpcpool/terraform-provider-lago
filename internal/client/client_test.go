package client

import "testing"

func TestNewClientValidation(t *testing.T) {
	t.Parallel()

	_, err := NewClient("", "key")
	if err == nil {
		t.Fatal("expected error for empty endpoint")
	}

	_, err = NewClient("https://api.getlago.com/api/v1", "")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}

	_, err = NewClient("not-a-url", "key")
	if err == nil {
		t.Fatal("expected error for invalid endpoint")
	}

	_, err = NewClient("https://api.getlago.com/api/v1", "key")
	if err != nil {
		t.Fatalf("expected valid client, got error: %v", err)
	}
}
