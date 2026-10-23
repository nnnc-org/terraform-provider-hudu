# terraform-provider-hudu

A Terraform / OpenTofu provider for [Hudu](https://www.hudu.com/), the IT
documentation platform. Built with the
[terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework).

## Supported resources

| Resource | Description |
|---|---|
| `hudu_company` | Companies (full CRUD). |
| `hudu_asset_layout` | Asset layouts with nested custom fields. **Note:** Hudu's API has no DELETE endpoint for layouts, so `terraform destroy` sets `active = false` instead. |
| `hudu_asset` | Assets with dynamic `custom_fields`. |
| `hudu_network` | Networks. |
| `hudu_vlan` | VLANs. |
| `hudu_vlan_zone` | VLAN zones (logical groupings of VLANs). |

## Provider configuration

```hcl
terraform {
  required_providers {
    hudu = {
      source = "nnnc-org/hudu"
    }
  }
}

provider "hudu" {
  host    = "https://hudu.example.com"   # or HUDU_API_HOST
  api_key = var.hudu_api_key             # or HUDU_API_KEY
}
```

## Local development

```sh
# Build & install into Terraform's local plugin cache
make install
```

Then point Terraform/OpenTofu at the local build via `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "nnnc-org/hudu" = "/home/<you>/.terraform.d/plugins/registry.terraform.io/nnnc-org/hudu/0.1.0/<os>_<arch>"
  }
  direct {}
}
```

With dev_overrides active, `terraform init` is unnecessary (and prints a
warning) — just `terraform plan` / `terraform apply`.

## Testing

```sh
make test         # unit tests
make testacc      # acceptance tests (hits a live Hudu tenant)
```

Acceptance tests expect `HUDU_API_HOST` and `HUDU_API_KEY` env vars.

## Known issues

### `hudu_asset_layout` is soft-deleted on destroy

The Hudu API has no `DELETE /asset_layouts/{id}` endpoint, so `terraform
destroy` issues `PUT /asset_layouts/{id}` with `active: false` (soft delete)
and removes the resource from state. The layout is hidden from the UI but not
permanently removed.

> **Historical note:** earlier Hudu versions returned HTTP 500 on *every* write
> to `PUT /asset_layouts/{id}` (observed against `huducloud.com` on 2026-06-05),
> which broke updates and soft-deletes. This was fixed upstream (re-verified
> 2026-10-23). If you hit 500s on layout updates/destroys with an older Hudu
> release, that's the cause — deactivate the layout manually via **Hudu Admin →
> Asset Layouts → ⋯ → Deactivate**.

### Sidebar folders (asset folders) can be referenced but not created

Asset layouts are grouped in the Hudu sidebar by "sidebar folders" (the
`sidebar_folders` resource in the admin UI). Set `sidebar_folder_id` on a
`hudu_asset_layout` to place the layout in a folder:

```hcl
resource "hudu_asset_layout" "computer" {
  name              = "Computer"
  sidebar_folder_id = 12   # an existing folder's numeric ID
  # ...
}
```

Hudu's public API exposes `sidebar_folder_id` as a foreign key on the layout
but provides **no endpoint to create, list, or delete the sidebar folders
themselves** (every `/sidebar_folders`-style route returns 404). Create the
folder once in the Hudu UI, note its numeric ID, and reference it here. Leaving
`sidebar_folder_id` unset keeps the layout at the top level.

### Asset CheckBox values must be native bools

Hudu's API docs say CheckBox custom fields accept string values (`"true"`),
but the GET response returns real booleans. To avoid plan/state drift, write
checkboxes as bools in HCL:

```hcl
custom_fields = {
  is_active = true   # NOT "true"
}
```

### API key scope

Creating an API key in Hudu requires explicit scoping. To use this provider
end-to-end, the key must have:

- **Destructive actions** enabled (otherwise `DELETE` on companies, networks,
  and VLANs fails with `401 Key does not have delete permissions!`).
- **Whitelisted IPs** that include the host running Terraform/OpenTofu (or no
  IP allowlist).

## Custom fields on `hudu_asset`

`custom_fields` is a `dynamic` attribute. Keys must match the layout field
labels converted to snake_case (lowercase, spaces → underscores). Values can be
any type the layout permits — strings, numbers, booleans, lists, or nested
objects for Address fields. Example:

```hcl
resource "hudu_asset" "example" {
  company_id      = hudu_company.acme.id
  asset_layout_id = hudu_asset_layout.computer.id
  name            = "Example"

  custom_fields = {
    operating_system = "Linux"
    score            = 5
    is_active        = true
    categories       = ["Item 1", "Item 2"]
    linked_assets    = [456, 789]
    office_location = {
      address_line_1 = "123 Main St"
      city           = "Denver"
      state          = "CO"
      zip            = "80202"
      country_name   = "USA"
    }
  }
}
```
