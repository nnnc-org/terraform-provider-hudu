package client

import (
	"context"
	"fmt"
	"net/url"
)

// VlanZone mirrors the API's VlanZone schema. A VLAN Zone is a logical grouping
// of related VLANs (e.g. a datacenter, building, or rack range).
type VlanZone struct {
	ID           int64   `json:"id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Slug         string  `json:"slug,omitempty"`
	Description  *string `json:"description,omitempty"`
	VlanIDRanges string  `json:"vlan_id_ranges,omitempty"`
	CompanyID    int64   `json:"company_id,omitempty"`
	ArchivedAt   *string `json:"archived_at,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
	VlansCount   int64   `json:"vlans_count,omitempty"`
	URL          string  `json:"url,omitempty"`
	// Archived is write-only — accepted on create/update; on read the API
	// reports state via archived_at.
	Archived *bool `json:"archived,omitempty"`
}

// CreateVlanZone POSTs /vlan_zones.
func (c *Client) CreateVlanZone(ctx context.Context, in *VlanZone) (*VlanZone, error) {
	var resp VlanZone
	if err := c.do(ctx, "POST", "/vlan_zones", nil, in, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVlanZone GETs /vlan_zones/{id}.
func (c *Client) GetVlanZone(ctx context.Context, id int64) (*VlanZone, error) {
	var resp VlanZone
	if err := c.do(ctx, "GET", fmt.Sprintf("/vlan_zones/%d", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateVlanZone PUTs /vlan_zones/{id}.
func (c *Client) UpdateVlanZone(ctx context.Context, id int64, in *VlanZone) (*VlanZone, error) {
	var resp VlanZone
	if err := c.do(ctx, "PUT", fmt.Sprintf("/vlan_zones/%d", id), nil, in, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVlanZone DELETEs /vlan_zones/{id}.
func (c *Client) DeleteVlanZone(ctx context.Context, id int64) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/vlan_zones/%d", id), nil, nil, nil)
}

// ListVlanZones returns the first page (25) matching the filters.
func (c *Client) ListVlanZones(ctx context.Context, companyID *int64, name string) ([]VlanZone, error) {
	q := url.Values{}
	if companyID != nil {
		q.Set("company_id", fmt.Sprintf("%d", *companyID))
	}
	if name != "" {
		q.Set("name", name)
	}
	var resp []VlanZone
	if err := c.do(ctx, "GET", "/vlan_zones", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
