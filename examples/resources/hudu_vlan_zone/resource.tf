resource "hudu_vlan_zone" "east_coast_dc" {
  name           = "East Coast Datacenter"
  company_id     = hudu_company.acme.id
  vlan_id_ranges = "100-500,1000-1500"
  description    = "Primary DC zones"
}
