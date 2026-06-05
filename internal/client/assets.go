package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Asset mirrors the API's Asset schema, with custom_fields handled separately
// since the wire shape differs between request and response.
type Asset struct {
	ID                  int64        `json:"id,omitempty"`
	CompanyID           int64        `json:"company_id,omitempty"`
	AssetLayoutID       int64        `json:"asset_layout_id,omitempty"`
	Slug                string       `json:"slug,omitempty"`
	Name                string       `json:"name,omitempty"`
	PrimarySerial       string       `json:"primary_serial,omitempty"`
	PrimaryMail         string       `json:"primary_mail,omitempty"`
	PrimaryModel        string       `json:"primary_model,omitempty"`
	PrimaryManufacturer string       `json:"primary_manufacturer,omitempty"`
	CompanyName         string       `json:"company_name,omitempty"`
	AssetType           string       `json:"asset_type,omitempty"`
	Archived            bool         `json:"archived,omitempty"`
	URL                 string       `json:"url,omitempty"`
	CreatedAt           string       `json:"created_at,omitempty"`
	UpdatedAt           string       `json:"updated_at,omitempty"`
	Fields              []AssetField `json:"fields,omitempty"`
}

// AssetField is the read-side representation of a custom field on an asset.
// Value is left as raw JSON so it can be any type the layout permits.
type AssetField struct {
	ID       int64           `json:"id,omitempty"`
	Label    string          `json:"label,omitempty"`
	Position int64           `json:"position,omitempty"`
	Value    json.RawMessage `json:"value,omitempty"`
}

// AssetCreateRequest is what we POST. custom_fields is an array of single-key
// objects, keyed by the layout field's snake_case label.
type AssetCreateRequest struct {
	Name                string           `json:"name,omitempty"`
	AssetLayoutID       int64            `json:"asset_layout_id,omitempty"`
	PrimarySerial       string           `json:"primary_serial,omitempty"`
	PrimaryMail         string           `json:"primary_mail,omitempty"`
	PrimaryModel        string           `json:"primary_model,omitempty"`
	PrimaryManufacturer string           `json:"primary_manufacturer,omitempty"`
	CustomFields        []map[string]any `json:"custom_fields,omitempty"`
}

// AssetUpdateRequest is what we PUT. Sent under an "asset" envelope per the API
// examples.
type AssetUpdateRequest struct {
	Asset AssetUpdateBody `json:"asset"`
}

// AssetUpdateBody is the inner payload of an AssetUpdateRequest.
type AssetUpdateBody struct {
	Name                *string          `json:"name,omitempty"`
	PrimarySerial       *string          `json:"primary_serial,omitempty"`
	PrimaryMail         *string          `json:"primary_mail,omitempty"`
	PrimaryModel        *string          `json:"primary_model,omitempty"`
	PrimaryManufacturer *string          `json:"primary_manufacturer,omitempty"`
	CustomFields        []map[string]any `json:"custom_fields,omitempty"`
}

type assetEnvelope struct {
	Asset Asset `json:"asset"`
}

// CreateAsset POSTs /companies/{company_id}/assets. The response wraps the
// asset in {"asset": {...}}.
func (c *Client) CreateAsset(ctx context.Context, companyID int64, in *AssetCreateRequest) (*Asset, error) {
	var env assetEnvelope
	path := fmt.Sprintf("/companies/%d/assets", companyID)
	if err := c.do(ctx, "POST", path, nil, in, &env); err != nil {
		return nil, err
	}
	return &env.Asset, nil
}

// GetAsset GETs /companies/{company_id}/assets/{id}.
func (c *Client) GetAsset(ctx context.Context, companyID, id int64) (*Asset, error) {
	var env assetEnvelope
	path := fmt.Sprintf("/companies/%d/assets/%d", companyID, id)
	if err := c.do(ctx, "GET", path, nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Asset, nil
}

// UpdateAsset PUTs /companies/{company_id}/assets/{id} with the {"asset": ...} envelope.
func (c *Client) UpdateAsset(ctx context.Context, companyID, id int64, in *AssetUpdateRequest) (*Asset, error) {
	var env assetEnvelope
	path := fmt.Sprintf("/companies/%d/assets/%d", companyID, id)
	if err := c.do(ctx, "PUT", path, nil, in, &env); err != nil {
		return nil, err
	}
	return &env.Asset, nil
}

// DeleteAsset DELETEs /companies/{company_id}/assets/{id}.
func (c *Client) DeleteAsset(ctx context.Context, companyID, id int64) error {
	path := fmt.Sprintf("/companies/%d/assets/%d", companyID, id)
	return c.do(ctx, "DELETE", path, nil, nil, nil)
}
