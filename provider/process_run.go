package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
	"time"
)

var processRunSchema = mergeSchemas(commandSchema, inputOutputCopySchema, map[string]*schema.Schema{
	"tries": {
		Description:  "The number of tries",
		Type:         schema.TypeInt,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Default:      1,
	},
	"retry_interval": {
		Description:  "The number of milliseconds to wait between the tries",
		Type:         schema.TypeInt,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Default:      500,
	},
	"timeout": {
		Description:  "The maximum number of milliseconds after the process gets killed, or no timeout if 0",
		Type:         schema.TypeInt,
		Default:      500,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.IntAtLeast(0),
	},
	"stdout": {
		Description: "The output in stdout of the last process",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"stderr": {
		Description: "The output in stderr of the last process",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"error": {
		Description: "True if the last process exited with an exit code not equal 0",
		Type:        schema.TypeBool,
		Computed:    true,
	},
	"timed_out": {
		Description: "True if the last process was killed because it took longer than the timeout",
		Type:        schema.TypeBool,
		Computed:    true,
	},
})

func dataSourceProcessRun() *schema.Resource {
	return &schema.Resource{
		ReadContext: runProcessRun,
		Schema:      processRunSchema,
	}
}

func resourceProcessRun() *schema.Resource {
	return &schema.Resource{
		CreateContext: runProcessRun,
		DeleteContext: processResourceDeleteFunc,
		Schema:        processRunSchema,
	}
}

func runProcessRun(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("static")

	var diagnostics diag.Diagnostics

	tries := d.Get("tries").(int)
	retryInterval := time.Duration(d.Get("retry_interval").(int)) * time.Millisecond
	log.Printf("[DEBUG] Run process with %d tries in a interval of %dms\n", tries, retryInterval/time.Millisecond)

	for tries > 0 {
		stdout, stderr, hasError, timedOut, err := doRunTry(d)
		if err != nil {
			return diag.FromErr(err)
		}

		if !hasError || tries == 1 {
			if err := d.Set("stdout", stdout); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("stderr", stderr); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("error", hasError); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("timed_out", timedOut); err != nil {
				return diag.FromErr(err)
			}
			if timedOut {
				diagnostics = append(diagnostics, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Process timed out",
					Detail:   fmt.Sprintf("The process took longer than %dms, so it was killed before ending.", d.Get("timeout")),
				})
			}
			break
		}
		log.Printf("[INFO] Running process failed! Retries left: %d\n", tries)

		time.Sleep(retryInterval)

		tries--
	}

	if err := copyInputOutput(d); err != nil {
		return diag.FromErr(err)
	}

	return diagnostics
}

func doRunTry(d *schema.ResourceData) (string, string, bool, bool, error) {
	cmd, err := createCommand(d)
	if err != nil {
		return "", "", false, false, err
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	args := strings.Join(cmd.Args, ", ")
	log.Printf("[INFO] Run command: {%s}\n", args)
	err = cmd.Start()
	var timedOut = false

	if err == nil {
		var timer *time.Timer

		if timeout := d.Get("timeout").(int); timeout != 0 {
			log.Printf("[DEBUG] Killing process when not done for %dms\n", timeout)
			timer = time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
				log.Printf("[WARN] Command timed out after %dms: {%s}\n", timeout, args)
				timedOut = true
				timer.Stop()
				cmd.Process.Kill()
			})
		}

		err = cmd.Wait()
		log.Printf("[DEBUG] Process exited: {%s}, stdout: %s, stderr: %s\n", args, stdout.String(), stderr.String())
		if timer != nil {
			timer.Stop()
		}
	} else {
		log.Printf("[ERROR] Could not start process {%s}: %s\n", args, err.Error())
	}

	return stdout.String(), stderr.String(), err != nil, timedOut, nil
}
