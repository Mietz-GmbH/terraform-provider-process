package provider

import (
	"context"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os/exec"
)

var processEndSchema = mergeSchemas(inputOutputCopySchema, map[string]*schema.Schema{
	"process_id": {
		Description: "A ID of the executed process which is a output of the \"start_process\" data source.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"kill": {
		Description: "If set to true, the process will be killed rather than waiting for the process to exit.",
		Type:        schema.TypeBool,
		Default:     true,
		Optional:    true,
		ForceNew:    true,
	},
	"error": {
		Description: "If true, the process exited with a exit code not equal 0.",
		Type:        schema.TypeBool,
		Computed:    true,
	},
})

func dataSourceProcessEnd() *schema.Resource {
	return &schema.Resource{
		ReadContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return runProcessEnd("dataSource", ctx, data, i)
		},
		Schema: processEndSchema,
	}
}

func resourceProcessEnd() *schema.Resource {
	return &schema.Resource{
		ReadContext: noneContextAction,
		CreateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return runProcessEnd("resource", ctx, data, i)
		},
		DeleteContext: noneContextAction,
		Schema:        processEndSchema,
	}
}

func runProcessEnd(kind string, _ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("static")

	processId := d.Get("process_id").(string)
	log.Printf("[DEBUG] Ending process %s\n", processId)

	if err := copyInputOutput(d); err != nil {
		return diag.FromErr(err)
	}

	cmd := processes[kind+processId]
	if cmd == nil {
		return diag.Diagnostics{{
			Severity:      diag.Warning,
			Summary:       "Invalid process ID",
			Detail:        "The argument process_id must be the process_id output of a process_start block. It must be the same type as the process_end block.",
			AttributePath: cty.Path{cty.GetAttrStep{Name: "process_id"}},
		}}
	}

	var exitError error
	if d.Get("kill").(bool) {
		log.Printf("[DEBUG] Killing process %s\n", processId)
		exitError = cmd.Process.Kill()
	} else {
		log.Printf("[DEBUG] Waiting for process %s to exit\n", processId)
		_, exitError = cmd.Process.Wait()
		log.Printf("[DEBUG] Process %s has been exited\n", processId)
	}

	var processError bool
	if exitError == nil {
		log.Printf("[DEBUG] Process %s exited without an error\n", processId)
		processError = false
	} else if _, ok := exitError.(*exec.ExitError); ok {
		log.Printf("[DEBUG] Process %s exited with error: %s\n", processId, exitError.Error())
		processError = true
	} else {
		return diag.FromErr(exitError)
	}

	if err := d.Set("error", processError); err != nil {
		return diag.FromErr(err)
	}

	processes[processId] = nil

	return diag.Diagnostics{}
}
