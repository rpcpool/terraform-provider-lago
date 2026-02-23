resource "lago_organization" "this" {
  name             = "Acme Corp"
  email            = "billing@acme.example.com"
  default_currency = "USD"
  timezone         = "America/New_York"

  legal_name                = "Acme Corporation Ltd."
  legal_number              = "123456789"
  tax_identification_number = "US123456789"

  address_line1 = "123 Main Street"
  address_line2 = "Suite 400"
  city          = "New York"
  state         = "NY"
  zipcode       = "10001"
  country       = "US"

  net_payment_term             = 30
  document_numbering           = "per_organization"
  document_number_prefix       = "ACME"
  finalize_zero_amount_invoice = true

  email_settings = [
    "invoice.finalized",
    "credit_note.created",
  ]

  billing_configuration {
    invoice_grace_period = 3
    invoice_footer       = "Thank you for your business."
    document_locale      = "en"
  }
}
