package client

import (
	"context"
	"fmt"
	"net/url"
)

// Vlan mirrors the API's Vlan schema.
type Vlan struct {
	ID               int64   `json:"id,omitempty"`
	Name             string  `json:"name,omitempty"`
	Slug             string  `json:"slug,omitempty"`
	VlanID           int64   `json:"vlan_id,omitempty"`
	Description      *string `json:"description,omitempty"`
	Notes            *string `json:"notes,omitempty"`
	CompanyID        int64   `json:"company_id,omitempty"`
	VlanZoneID       *int64  `json:"vlan_zone_id,omitempty"`
	StatusListItemID *int64  `json:"status_list_item_id,omitempty"`
	RoleListItemID   *int64  `json:"role_list_item_id,omitempty"`
	ArchivedAt       *string `json:"archived_at,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
	NetworksCount    int64   `json:"networks_count,omitempty"`
	URL              string  `json:"url,omitempty"`
	// Archived is write-only.
	Archived *bool `json:"archived,omitempty"`
}

// CreateVlan POSTs /vlans.
func (c *Client) CreateVlan(ctx context.Context, in *Vlan) (*Vlan, error) {
	var resp Vlan
	if err := c.do(ctx, "POST", "/vlans", nil, in, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVlan GETs /vlans/{id}.
func (c *Client) GetVlan(ctx context.Context, id int64) (*Vlan, error) {
	var resp Vlan
	if err := c.do(ctx, "GET", fmt.Sprintf("/vlans/%d", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateVlan PUTs /vlans/{id}.
func (c *Client) UpdateVlan(ctx context.Context, id int64, in *Vlan) (*Vlan, error) {
	var resp Vlan
	if err := c.do(ctx, "PUT", fmt.Sprintf("/vlans/%d", id), nil, in, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVlan DELETEs /vlans/{id}.
func (c *Client) DeleteVlan(ctx context.Context, id int64) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/vlans/%d", id), nil, nil, nil)
}

// ListVlans returns the first page (25) matching the filters.
func (c *Client) ListVlans(ctx context.Context, companyID *int64, name string) ([]Vlan, error) {
	q := url.Values{}
	if companyID != nil {
		q.Set("company_id", fmt.Sprintf("%d", *companyID))
	}
	if name != "" {
		q.Set("name", name)
	}
	var resp []Vlan
	if err := c.do(ctx, "GET", "/vlans", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
