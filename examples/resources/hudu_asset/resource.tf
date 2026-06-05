resource "hudu_asset" "laptop_001" {
  company_id      = hudu_company.acme.id
  asset_layout_id = hudu_asset_layout.computer.id

  name                 = "LAPTOP-001"
  primary_serial       = "SN123456"
  primary_model        = "X1 Carbon"
  primary_manufacturer = "Lenovo"

  custom_fields = {
    operating_system = "Linux"
    serial_number    = "SN123456"
    purchase_date    = "2024/01/15"
  }
}
