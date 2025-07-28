package main

import (
	"log"
	"os"

	"github.com/aliyun/terraform-provider-alicloud/alicloud"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// Print panic info to stderr
			log.SetOutput(os.Stderr)
			log.Printf("PANIC occurred: %v", r)
		}
	}()
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: alicloud.Provider})
}
