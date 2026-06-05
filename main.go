package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nnnc-org/terraform-provider-hudu/internal/provider"
)

// version is set by goreleaser at build time via -ldflags.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run provider with support for delve debugging")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/nnnc-org/hudu",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
