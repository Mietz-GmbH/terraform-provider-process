package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-providers/terraform-provider-process/provider"
)

func main() {
	defer provider.HandleExit()
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: provider.Provider})
}
