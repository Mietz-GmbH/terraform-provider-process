package provider

import (
	"bytes"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
	"time"
)

var processRunSchema = mergeSchemas(phasedSchema, commandSchema, inputOutputCopySchema, map[string]*schema.Schema{
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
		Description:  "The maximum number of milliseconds after the process gets killed",
		Type:         schema.TypeInt,
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
		Description: "True if the last process exited with a exit code not equal 0",
		Type:        schema.TypeBool,
		Computed:    true,
	},
})

func dataSourceProcessRun() *schema.Resource {
	return &schema.Resource{
		//ReadContext: runProcessRun,
		Schema: processRunSchema,
	}
}

func resourceProcessRun() *schema.Resource {
	return &schema.Resource{
		ReadContext:   phasedFunc("plan", runProcessRun),
		CreateContext: phasedFunc("apply", runProcessRun),
		DeleteContext: processResourceDeleteFunc,
		Schema:        processRunSchema,
	}
}

func runProcessRun(_ string, _ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	tries := d.Get("tries").(int)
	retryInterval := time.Duration(d.Get("retry_interval").(int)) * time.Millisecond
	log.Printf("[DEBUG] Run process with %d tries in a interval of %dms\n", tries, retryInterval/time.Millisecond)

	for tries > 0 {
		stdout, stderr, hasError := doRunTry(d)

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
			break
		}
		log.Printf("[INFO] Running process failed! Retries left: %d\n", tries)

		time.Sleep(retryInterval)

		tries--
	}

	if err := copyInputOutput(d); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func doRunTry(d *schema.ResourceData) (string, string, bool) {
	cmd := createCommand(d)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	args := strings.Join(cmd.Args, ", ")
	log.Printf("[INFO] Run command: {%s}\n", args)
	err := cmd.Start()

	if err == nil {
		timeout := d.Get("timeout")
		var timer *time.Timer
		if timeout != nil {
			log.Printf("[DEBUG] Killing process when not done for %dms\n", timeout)
			timer = time.AfterFunc(time.Duration(timeout.(int))*time.Millisecond, func() {
				log.Printf("[WARN] Command timed out after %dms: {%s}\n", timeout, args)
				timer.Stop()
				cmd.Process.Kill()
			})
		}

		err = cmd.Wait()
		log.Printf("[DEBUG] Process exited: {%s}\n", args)
		if timer != nil {
			timer.Stop()
		}
	} else {
		log.Printf("[ERROR] Could not start process {%s}: %s\n", args, err.Error())
	}

	return stdout.String(), stderr.String(), err != nil
}
