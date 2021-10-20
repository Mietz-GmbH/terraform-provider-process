package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os/exec"
)

var processes = map[string]*exec.Cmd{}

var processStartSchema = mergeSchemas(commandSchema, inputOutputCopySchema, map[string]*schema.Schema{
	"process_id": {
		Description: "A ID of the started process which can be used to end the process with the `end_process` data source.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"pid": {
		Description: "The PID (Process Identifier) of the spawned process",
		Type:        schema.TypeInt,
		Computed:    true,
	},
})

func dataSourceProcessStart() *schema.Resource {
	return &schema.Resource{
		ReadContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return runProcessStart("dataSource", ctx, data, i)
		},
		Schema: processStartSchema,
	}
}

func resourceProcessStart() *schema.Resource {
	return &schema.Resource{
		CreateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return runProcessStart("resource", ctx, data, i)
		},
		DeleteContext: processResourceDeleteFunc,
		Schema:        processStartSchema,
	}
}

func runProcessStart(kind string, _ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("static")

	log.Printf("[DEBUG] Starting process as kind %s\n", kind)
	cmd, err := createCommand(d)
	if err != nil {
		return diag.FromErr(err)
	}

	processId := id()
	if err := d.Set("process_id", processId); err != nil {
		return diag.FromErr(err)
	}

	processes[kind+processId] = cmd
	log.Printf("[DEBUG] Stored new process with id: %d\n", processId)

	if err := cmd.Start(); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Started process with id %d\n", processId)

	if err := d.Set("pid", cmd.Process.Pid); err != nil {
		return diag.FromErr(err)
	}

	if err := copyInputOutput(d); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func HandleExit() {
	for _, cmd := range processes {
		if cmd != nil {
			_ = cmd.Process.Kill()
		}
	}
}
