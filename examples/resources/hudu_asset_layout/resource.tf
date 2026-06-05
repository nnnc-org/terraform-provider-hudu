resource "hudu_asset_layout" "computer" {
  name              = "Computer Assets"
  icon              = "fas fa-laptop"
  color             = "#1f77b4"
  icon_color        = "#ffffff"
  include_passwords = true
  include_photos    = true
  include_comments  = true
  include_files     = true

  fields = [
    {
      label        = "Operating System"
      field_type   = "Text"
      position     = 1
      required     = true
      show_in_list = true
    },
    {
      label      = "Serial Number"
      field_type = "Text"
      position   = 2
    },
    {
      label      = "Purchase Date"
      field_type = "Date"
      position   = 3
      expiration = false
    },
  ]
}
