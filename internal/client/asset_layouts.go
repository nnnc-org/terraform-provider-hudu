package client

import (
	"context"
	"fmt"
	"net/url"
)

// AssetLayout mirrors the API's Asset_Layout schema.
type AssetLayout struct {
	ID               int64              `json:"id,omitempty"`
	Slug             string             `json:"slug,omitempty"`
	Name             string             `json:"name,omitempty"`
	Icon             string             `json:"icon,omitempty"`
	Color            string             `json:"color,omitempty"`
	IconColor        string             `json:"icon_color,omitempty"`
	SidebarFolderID  *int64             `json:"sidebar_folder_id,omitempty"`
	Active           *bool              `json:"active,omitempty"`
	IncludePasswords *bool              `json:"include_passwords,omitempty"`
	IncludePhotos    *bool              `json:"include_photos,omitempty"`
	IncludeComments  *bool              `json:"include_comments,omitempty"`
	IncludeFiles     *bool              `json:"include_files,omitempty"`
	PasswordTypes    *string            `json:"password_types,omitempty"`
	CreatedAt        string             `json:"created_at,omitempty"`
	UpdatedAt        string             `json:"updated_at,omitempty"`
	Fields           []AssetLayoutField `json:"fields,omitempty"`
}

// AssetLayoutField mirrors the API's Asset_Layout_Field schema.
type AssetLayoutField struct {
	ID              int64  `json:"id,omitempty"`
	Label           string `json:"label,omitempty"`
	ShowInList      *bool  `json:"show_in_list,omitempty"`
	FieldType       string `json:"field_type,omitempty"`
	Required        *bool  `json:"required,omitempty"`
	Hint            string `json:"hint,omitempty"`
	Min             *int64 `json:"min,omitempty"`
	Max             *int64 `json:"max,omitempty"`
	LinkableID      *int64 `json:"linkable_id,omitempty"`
	Expiration      *bool  `json:"expiration,omitempty"`
	Options         string `json:"options,omitempty"`
	Position        int64  `json:"position,omitempty"`
	IsDestroyed     *bool  `json:"is_destroyed,omitempty"`
	ListID          *int64 `json:"list_id,omitempty"`
	MultipleOptions *bool  `json:"multiple_options,omitempty"`
}

type assetLayoutEnvelope struct {
	AssetLayout AssetLayout `json:"asset_layout"`
}

// CreateAssetLayout POSTs /asset_layouts with the {"asset_layout": ...} envelope.
func (c *Client) CreateAssetLayout(ctx context.Context, in *AssetLayout) (*AssetLayout, error) {
	body := assetLayoutEnvelope{AssetLayout: *in}
	var resp assetLayoutEnvelope
	if err := c.do(ctx, "POST", "/asset_layouts", nil, body, &resp); err != nil {
		return nil, err
	}
	return &resp.AssetLayout, nil
}

// GetAssetLayout GETs /asset_layouts/{id}.
func (c *Client) GetAssetLayout(ctx context.Context, id int64) (*AssetLayout, error) {
	var env assetLayoutEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/asset_layouts/%d", id), nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.AssetLayout, nil
}

// UpdateAssetLayout PUTs /asset_layouts/{id} with the envelope.
func (c *Client) UpdateAssetLayout(ctx context.Context, id int64, in *AssetLayout) (*AssetLayout, error) {
	body := assetLayoutEnvelope{AssetLayout: *in}
	var resp assetLayoutEnvelope
	if err := c.do(ctx, "PUT", fmt.Sprintf("/asset_layouts/%d", id), nil, body, &resp); err != nil {
		return nil, err
	}
	return &resp.AssetLayout, nil
}

// DeactivateAssetLayout is the closest thing to a delete the API exposes —
// it PUTs the layout with active:false. Hudu's PUT validates required fields
// and 500s if too few are supplied, so we GET the current layout first and
// echo it back with active toggled off.
func (c *Client) DeactivateAssetLayout(ctx context.Context, id int64) error {
	current, err := c.GetAssetLayout(ctx, id)
	if err != nil {
		return err
	}
	inactive := false
	current.Active = &inactive
	// Strip fields[] from the echo — sending them back unchanged can trip
	// Hudu validation (e.g. is_destroyed flag) and isn't necessary just to
	// flip the active bit.
	current.Fields = nil
	body := assetLayoutEnvelope{AssetLayout: *current}
	return c.do(ctx, "PUT", fmt.Sprintf("/asset_layouts/%d", id), nil, body, nil)
}

// AssetLayoutListFilters narrows /asset_layouts queries.
type AssetLayoutListFilters struct {
	Name string
	Slug string
}

// ListAssetLayouts returns the first page (25) matching filters.
func (c *Client) ListAssetLayouts(ctx context.Context, f AssetLayoutListFilters) ([]AssetLayout, error) {
	q := url.Values{}
	if f.Name != "" {
		q.Set("name", f.Name)
	}
	if f.Slug != "" {
		q.Set("slug", f.Slug)
	}
	var resp struct {
		AssetLayouts []AssetLayout `json:"asset_layouts"`
	}
	if err := c.do(ctx, "GET", "/asset_layouts", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp.AssetLayouts, nil
}
