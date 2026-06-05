resource "hudu_vlan" "production" {
  name        = "Production"
  vlan_id     = 100
  company_id  = hudu_company.acme.id
  description = "Primary production VLAN"
}
