resource "hudu_company" "acme" {
  name           = "Acme Corp"
  nickname       = "Acme"
  company_type   = "Customer"
  address_line_1 = "123 Main St"
  city           = "Denver"
  state          = "CO"
  zip            = "80202"
  country_name   = "USA"
  phone_number   = "+1-555-0100"
  website        = "https://acme.example.com"
}
