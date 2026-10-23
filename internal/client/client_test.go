package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(Config{
		Host:    srv.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestCreateCompany_SendsAPIKeyAndDecodesResponse(t *testing.T) {
	var gotKey, gotMethod, gotPath, gotCT string
	var gotBody map[string]any

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-api-key")
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":42,"name":"Acme","slug":"acme"}`))
	}))

	out, err := c.CreateCompany(context.Background(), &Company{Name: "Acme"})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	if gotKey != "test-key" {
		t.Errorf("api key header = %q, want test-key", gotKey)
	}
	if gotMethod != "POST" || gotPath != "/api/v1/companies" {
		t.Errorf("got %s %s, want POST /api/v1/companies", gotMethod, gotPath)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}
	if gotBody["name"] != "Acme" {
		t.Errorf("body name = %v, want Acme", gotBody["name"])
	}
	if out.ID != 42 || out.Name != "Acme" || out.Slug != "acme" {
		t.Errorf("decoded company = %+v", out)
	}
}

func TestGetCompany_UnwrapsEnvelope(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/companies/42" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"company":{"id":42,"name":"Acme"}}`))
	}))

	out, err := c.GetCompany(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if out.ID != 42 || out.Name != "Acme" {
		t.Errorf("got %+v", out)
	}
}

func TestGetCompany_404ReturnsErrNotFound(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	_, err := c.GetCompany(context.Background(), 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestAssetLayoutCreate_WrapsEnvelope(t *testing.T) {
	var got map[string]any
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"asset_layout":{"id":1,"name":"X"}}`))
	}))

	_, err := c.CreateAssetLayout(context.Background(), &AssetLayout{Name: "X"})
	if err != nil {
		t.Fatalf("CreateAssetLayout: %v", err)
	}
	envelope, ok := got["asset_layout"].(map[string]any)
	if !ok {
		t.Fatalf("body missing asset_layout key: %v", got)
	}
	if envelope["name"] != "X" {
		t.Errorf("name = %v", envelope["name"])
	}
}

func TestDeactivateAssetLayout_SendsActiveFalse(t *testing.T) {
	var got map[string]any
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("method = %s", r.Method)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	if err := c.DeactivateAssetLayout(context.Background(), 7); err != nil {
		t.Fatalf("DeactivateAssetLayout: %v", err)
	}
	env, _ := got["asset_layout"].(map[string]any)
	if env == nil || env["active"] != false {
		t.Errorf("body = %v, want active=false under asset_layout", got)
	}
}

func TestListNetworks_BareArrayResponse(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.RawQuery, "company_id=") {
			t.Errorf("missing company_id query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`[{"id":1,"name":"net1"},{"id":2,"name":"net2"}]`))
	}))
	companyID := int64(5)
	out, err := c.ListNetworks(context.Background(), &companyID, "")
	if err != nil {
		t.Fatalf("ListNetworks: %v", err)
	}
	if len(out) != 2 || out[0].Name != "net1" {
		t.Errorf("got %+v", out)
	}
}

func TestCreateVlanZone_BareRequestAndResponse(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"name":"East DC","vlan_id_ranges":"100-500","company_id":45,"slug":"east-dc"}`))
	}))

	out, err := c.CreateVlanZone(context.Background(), &VlanZone{
		Name:         "East DC",
		VlanIDRanges: "100-500",
		CompanyID:    45,
	})
	if err != nil {
		t.Fatalf("CreateVlanZone: %v", err)
	}
	if gotMethod != "POST" || gotPath != "/api/v1/vlan_zones" {
		t.Errorf("got %s %s, want POST /api/v1/vlan_zones", gotMethod, gotPath)
	}
	// Request body is sent bare (not wrapped in a "vlan_zone" envelope).
	if gotBody["name"] != "East DC" || gotBody["vlan_id_ranges"] != "100-500" {
		t.Errorf("body = %v", gotBody)
	}
	if out.ID != 7 || out.Name != "East DC" || out.VlanIDRanges != "100-500" || out.Slug != "east-dc" {
		t.Errorf("decoded vlan zone = %+v", out)
	}
}

func TestRetryOn500(t *testing.T) {
	calls := 0
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"company":{"id":1,"name":"X"}}`))
	}))
	_, err := c.GetCompany(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestAPIError_BodyIsPreserved(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"bad data"}`))
	}))
	_, err := c.CreateCompany(context.Background(), &Company{})
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("want *APIError, got %T: %v", err, err)
	}
	if ae.StatusCode != 422 || !strings.Contains(ae.Body, "bad data") {
		t.Errorf("APIError = %+v", ae)
	}
}

// drain ensures the test server's response bodies are fully read so
// keep-alive connections don't leak between subtests.
var _ = io.Discard
