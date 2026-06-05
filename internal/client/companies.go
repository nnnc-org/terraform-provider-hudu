package client

import (
	"context"
	"fmt"
	"net/url"
)

// Company mirrors the API's Company schema. Pointer fields capture nullability
// so we can distinguish "unset" from "zero value".
type Company struct {
	ID                int64   `json:"id,omitempty"`
	Slug              string  `json:"slug,omitempty"`
	Name              string  `json:"name,omitempty"`
	Nickname          *string `json:"nickname,omitempty"`
	CompanyType       *string `json:"company_type,omitempty"`
	AddressLine1      *string `json:"address_line_1,omitempty"`
	AddressLine2      *string `json:"address_line_2,omitempty"`
	City              *string `json:"city,omitempty"`
	State             *string `json:"state,omitempty"`
	Zip               *string `json:"zip,omitempty"`
	CountryName       *string `json:"country_name,omitempty"`
	PhoneNumber       *string `json:"phone_number,omitempty"`
	FaxNumber         *string `json:"fax_number,omitempty"`
	Website           *string `json:"website,omitempty"`
	IDNumber          *string `json:"id_number,omitempty"`
	ParentCompanyID   *int64  `json:"parent_company_id,omitempty"`
	ParentCompanyName *string `json:"parent_company_name,omitempty"`
	Notes             *string `json:"notes,omitempty"`
	Archived          bool    `json:"archived,omitempty"`
	URL               string  `json:"url,omitempty"`
	FullURL           string  `json:"full_url,omitempty"`
	CreatedAt         string  `json:"created_at,omitempty"`
	UpdatedAt         string  `json:"updated_at,omitempty"`
}

type companyEnvelope struct {
	Company Company `json:"company"`
}

// CreateCompany POSTs /companies. Despite the OpenAPI spec showing a bare
// Company response, the live API wraps the result in {"company": {...}}.
func (c *Client) CreateCompany(ctx context.Context, in *Company) (*Company, error) {
	var env companyEnvelope
	if err := c.do(ctx, "POST", "/companies", nil, in, &env); err != nil {
		return nil, err
	}
	return &env.Company, nil
}

// GetCompany GETs /companies/{id}.
func (c *Client) GetCompany(ctx context.Context, id int64) (*Company, error) {
	var env companyEnvelope
	if err := c.do(ctx, "GET", fmt.Sprintf("/companies/%d", id), nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Company, nil
}

// UpdateCompany PUTs /companies/{id}.
func (c *Client) UpdateCompany(ctx context.Context, id int64, in *Company) (*Company, error) {
	var env companyEnvelope
	if err := c.do(ctx, "PUT", fmt.Sprintf("/companies/%d", id), nil, in, &env); err != nil {
		return nil, err
	}
	return &env.Company, nil
}

// DeleteCompany DELETEs /companies/{id}.
func (c *Client) DeleteCompany(ctx context.Context, id int64) error {
	return c.do(ctx, "DELETE", fmt.Sprintf("/companies/%d", id), nil, nil, nil)
}

// ArchiveCompany PUTs /companies/{id}/archive.
func (c *Client) ArchiveCompany(ctx context.Context, id int64) (*Company, error) {
	var env companyEnvelope
	if err := c.do(ctx, "PUT", fmt.Sprintf("/companies/%d/archive", id), nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Company, nil
}

// UnarchiveCompany PUTs /companies/{id}/unarchive.
func (c *Client) UnarchiveCompany(ctx context.Context, id int64) (*Company, error) {
	var env companyEnvelope
	if err := c.do(ctx, "PUT", fmt.Sprintf("/companies/%d/unarchive", id), nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Company, nil
}

// CompanyListFilters narrows /companies queries.
type CompanyListFilters struct {
	Name   string
	Slug   string
	Search string
}

// ListCompanies returns the first page (25) matching the filters. Pagination
// across all pages is the caller's responsibility — Terraform reads always go
// through GetCompany by ID.
func (c *Client) ListCompanies(ctx context.Context, f CompanyListFilters) ([]Company, error) {
	q := url.Values{}
	if f.Name != "" {
		q.Set("name", f.Name)
	}
	if f.Slug != "" {
		q.Set("slug", f.Slug)
	}
	if f.Search != "" {
		q.Set("search", f.Search)
	}
	var resp struct {
		Companies []Company `json:"companies"`
	}
	if err := c.do(ctx, "GET", "/companies", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Companies, nil
}
