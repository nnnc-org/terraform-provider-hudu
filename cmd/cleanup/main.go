// cleanup is a one-off helper that deletes Hudu records matching a name
// pattern. Use it to remove orphans created by partial smoketest runs.
//
// Usage:
//
//	HUDU_API_HOST=... HUDU_API_KEY=... go run ./cmd/cleanup -name tfprov-smoketest
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nnnc-org/terraform-provider-hudu/internal/client"
)

func main() {
	prefix := flag.String("name", "tfprov-smoketest", "match resources whose name starts with this prefix")
	flag.Parse()

	host := os.Getenv("HUDU_API_HOST")
	apiKey := os.Getenv("HUDU_API_KEY")
	if host == "" || apiKey == "" {
		log.Fatal("HUDU_API_HOST and HUDU_API_KEY must be set")
	}
	c, err := client.New(client.Config{Host: host, APIKey: apiKey})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Networks
	nets, err := c.ListNetworks(ctx, nil, "")
	if err == nil {
		for _, n := range nets {
			if strings.HasPrefix(n.Name, *prefix) {
				fmt.Printf("deleting network %d %q\n", n.ID, n.Name)
				if err := c.DeleteNetwork(ctx, n.ID); err != nil {
					log.Printf("  delete failed: %v", err)
				}
			}
		}
	}

	// VLANs
	vlans, err := c.ListVlans(ctx, nil, "")
	if err == nil {
		for _, v := range vlans {
			if strings.HasPrefix(v.Name, *prefix) {
				fmt.Printf("deleting vlan %d %q\n", v.ID, v.Name)
				if err := c.DeleteVlan(ctx, v.ID); err != nil {
					log.Printf("  delete failed: %v", err)
				}
			}
		}
	}

	// Companies (and assets under them - DELETE cascades on company removal in Hudu)
	companies, err := c.ListCompanies(ctx, client.CompanyListFilters{Search: *prefix})
	if err == nil {
		for _, co := range companies {
			if strings.HasPrefix(co.Name, *prefix) {
				fmt.Printf("deleting company %d %q\n", co.ID, co.Name)
				if err := c.DeleteCompany(ctx, co.ID); err != nil {
					log.Printf("  delete failed: %v", err)
				}
			}
		}
	}

	// Asset layouts: no DELETE - deactivate.
	layouts, err := c.ListAssetLayouts(ctx, client.AssetLayoutListFilters{})
	if err == nil {
		for _, l := range layouts {
			if strings.HasPrefix(l.Name, *prefix) {
				fmt.Printf("deactivating layout %d %q\n", l.ID, l.Name)
				if err := c.DeactivateAssetLayout(ctx, l.ID); err != nil {
					log.Printf("  deactivate failed: %v", err)
				}
			}
		}
	}
}
