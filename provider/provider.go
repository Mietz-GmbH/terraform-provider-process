package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		//DataSourcesMap: map[string]*schema.Resource{
		//	"process_start": dataSourceProcessStart(),
		//	"process_end":   dataSourceProcessEnd(),
		//	"process_run":   dataSourceProcessRun(),
		//},
		ResourcesMap: map[string]*schema.Resource{
			"process_start": resourceProcessStart(),
			"process_end":   resourceProcessEnd(),
			"process_run":   resourceProcessRun(),
		},
	}
}
