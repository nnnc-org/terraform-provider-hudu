terraform {
  required_providers {
    hudu = {
      source = "nnnc-org/hudu"
    }
  }
}

# host and api_key come from HUDU_API_HOST / HUDU_API_KEY in the environment.
provider "hudu" {}

locals {
  prefix = "tfprov-smoketest"
}

resource "hudu_company" "smoke" {
  name           = "${local.prefix}-company"
  nickname       = "Smoketest"
  company_type   = "Internal"
  address_line_1 = "1 Test Way"
  city           = "Boulder"
  state          = "CO"
  zip            = "80301"
  country_name   = "USA"
  phone_number   = "+1-555-0100"
  website        = "https://example.invalid"
  notes          = "Created by the terraform-provider-hudu smoketest. Safe to delete."
}

resource "hudu_asset_layout" "smoke" {
  name              = "${local.prefix}-layout"
  icon              = "fas fa-laptop"
  color             = "#1f77b4"
  icon_color        = "#ffffff"
  include_passwords = false
  include_photos    = false
  include_comments  = true
  include_files     = false

  fields = [
    {
      label        = "Operating System"
      field_type   = "Text"
      position     = 1
      required     = false
      show_in_list = true
    },
    {
      label      = "Score"
      field_type = "Number"
      position   = 2
    },
    {
      label      = "Is Active"
      field_type = "CheckBox"
      position   = 3
    },
  ]
}

resource "hudu_vlan_zone" "smoke" {
  name           = "${local.prefix}-vlan-zone"
  company_id     = tonumber(hudu_company.smoke.id)
  vlan_id_ranges = "3900-4094"
  description    = "Smoketest VLAN zone"
}

resource "hudu_vlan" "smoke" {
  name         = "${local.prefix}-vlan"
  vlan_id      = 4000
  company_id   = tonumber(hudu_company.smoke.id)
  vlan_zone_id = tonumber(hudu_vlan_zone.smoke.id)
  description  = "Smoketest VLAN"
}

resource "hudu_network" "smoke" {
  name        = "${local.prefix}-network"
  address     = "10.250.0.0/24"
  company_id  = tonumber(hudu_company.smoke.id)
  description = "Smoketest network"
  vlan_id     = tonumber(hudu_vlan.smoke.id)
}

resource "hudu_asset" "smoke" {
  company_id      = tonumber(hudu_company.smoke.id)
  asset_layout_id = tonumber(hudu_asset_layout.smoke.id)

  name                 = "${local.prefix}-asset"
  primary_serial       = "SMOKE-SN-1"
  primary_model        = "Test Model"
  primary_manufacturer = "Test Mfr"

  custom_fields = {
    operating_system = "Linux"
    score            = 7
    is_active        = true
  }
}

output "company_id" {
  value = hudu_company.smoke.id
}

output "asset_url" {
  value = hudu_asset.smoke.url
}

output "vlan_zone_id" {
  value = hudu_vlan_zone.smoke.id
}

output "vlan_zone_url" {
  value = hudu_vlan_zone.smoke.url
}
