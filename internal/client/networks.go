package client

import (
	"context"
	"fmt"
	"net/url"
)

// Network mirrors the API's Network schema.
type Network struct {
	ID               int64   `json:"id,omitempty"`
	Name             string  `json:"name,omitempty"`
	Address          string  `json:"address,omitempty"`
	NetworkType      *int64  `json:"network_type,omitempty"`
	Slug             string  `json:"slug,omitempty"`
	CompanyID        int64   `json:"company_id,omitempty"`
	LocationID       *int64  `json:"location_id,omitempty"`
	Description      *string `json:"description,omitempty"`
	Notes            *string `json:"notes,omitempty"`
	Ancestry         *string `json:"ancestry,omitempty"`
	SyncIdentifier   *string `json:"sync_identifier,omitempty"`
	IsRadar          *bool   `json:"is_radar,omitempty"`
	StatusListItemID *int64  `json:"status_list_item_id,omitempty"`
	RoleListItemID   *int64  `json:"role_list_item_id,omitempty"`
	VlanID           *int64  `json:"vlan_id,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
	URL              string  `json:"url,omitempty"`
	ArchivedAt       *string `json:"archived_at,omitempty"`
	// Archived is write-only — accepted on create/update; on read the API
	// reports state via archived_at.
	Archived *bool `json:"archived,omitempty"`
}

// CreateNetwork POSTs /networks.
func (c *Client) CreateNetwork(ctx context.Context, in *Network) (*Network, error) {
	var resp Network
	if err := c.do(ctx, "POST", "/networks", nil, in, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetNetwork GETs /networks/{id}.
func (c *Client) GetNetwork(ctx context.Context, id int64) (*Network, error) {
	var resp Network
	if err := c.do(ctx, "GET", fmt.Sprintf("/networks/%d", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateNetwork PUTs /networks/{id}.
func (c *Client) UpdateNetwork(ctx context.Context, id int64, in *Network) (*Network, error) {
	var resp Network
	if err := c.do(ctx, "PUT", fmt.Sprintf("/networks/%d", id), nil, in, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteNetwork DELETEs /networks/{id}.
func (c *Client) DeleteNetwork(ctx context.Context, id int64) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/networks/%d", id), nil, nil, nil)
}

// ListNetworks returns the first page (25) matching the filters.
func (c *Client) ListNetworks(ctx context.Context, companyID *int64, name string) ([]Network, error) {
	q := url.Values{}
	if companyID != nil {
		q.Set("company_id", fmt.Sprintf("%d", *companyID))
	}
	if name != "" {
		q.Set("name", name)
	}
	var resp []Network
	if err := c.do(ctx, "GET", "/networks", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
