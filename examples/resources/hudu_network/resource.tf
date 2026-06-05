resource "hudu_network" "office" {
  name        = "Office Network"
  address     = "192.168.1.0/24"
  company_id  = hudu_company.acme.id
  description = "Main office local network"
  vlan_id     = hudu_vlan.production.id
}
