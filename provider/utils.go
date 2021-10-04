package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"math/rand"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

var phasedSchema = map[string]*schema.Schema{
	"phase": {
		Description:  "Defines the phase in which the command should be executed",
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice([]string{"plan", "apply"}, false),
		Required:     true,
		ForceNew:     true,
	},
}

var commandSchema = map[string]*schema.Schema{
	"triggers": {
		Description: "A map of arbitrary strings that, when changed, will force the null resource to be replaced, re-running any associated provisioners.",
		Type:        schema.TypeMap,
		Optional:    true,
		ForceNew:    true,
	},
	"command": {
		Description: "A block which defines a possible command",
		Type:        schema.TypeList,
		MinItems:    1,
		Required:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"platforms": {
					Description: "The platforms which are valid for this configuration",
					Type:        schema.TypeSet,
					Required:    true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"darwin", "freebsd", "linux", "openbsd", "solaris", "windows"}, false),
					},
				},
				"command": {
					Description: "The command which will be executed.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"environment": {
					Description: "A map of environment variables.",
					Type:        schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
				},
				"sensitive_environment": {
					Description: "A map of environment variables which are sensitive.",
					Type:        schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional:  true,
					Sensitive: true,
				},
				"interpreter": {
					Description: "The command line which spawns the process.",
					Type:        schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
				},
				"working_directory": {
					Description: "The working directory to start the process in.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
	},
}

var inputOutputCopySchema = map[string]*schema.Schema{
	"inputs": {
		Description: "A map of arbitrary strings that is copied into the `outputs` attribute, and accessible directly for interpolation.",
		Type:        schema.TypeMap,
		ForceNew:    true,
		Optional:    true,
	},
	"outputs": {
		Description: "After the data source is \"read\", a copy of the `inputs` map.",
		Type:        schema.TypeMap,
		Computed:    true,
	},
}

func selectCommandConfiguration(d *schema.ResourceData) map[string]interface{} {
	for _, current := range d.Get("command").([]interface{}) {
		configuration := current.(map[string]interface{})
		platforms := configuration["platforms"].(*schema.Set)
		if platforms.Contains(runtime.GOOS) {
			return configuration
		}
	}

	return nil
}

func createCommand(d *schema.ResourceData) *exec.Cmd {
	configuration := selectCommandConfiguration(d)
	if configuration == nil {
		return nil
	}

	interpreter := configuration["interpreter"].([]interface{})

	if len(interpreter) == 0 {
		if runtime.GOOS == "windows" {
			interpreter = []interface{}{"powershell.exe"}
		} else {
			interpreter = []interface{}{"/bin/bash", "-c"}
		}
	}

	var args []string

	for _, value := range interpreter {
		args = append(args, value.(string))
	}

	log.Printf("[DEBUG] Create command with interpreter: {%s}\n", strings.Join(args, ", "))
	args = append(args, configuration["command"].(string))

	cmd := &exec.Cmd{
		Path: args[0],
		Args: args,
	}

	if environment := configuration["environment"].(map[string]interface{}); environment != nil {
		for key, value := range environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	if environment := configuration["sensitive_environment"].(map[string]interface{}); environment != nil {
		for key, value := range environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	if workingDirectory := configuration["working_directory"]; workingDirectory != nil {
		cmd.Dir = workingDirectory.(string)
	}

	log.Printf("[DEBUG] Initialized command: {%s}\n", strings.Join(cmd.Args, ", "))

	return cmd
}

func mergeSchemas(maps ...map[string]*schema.Schema) map[string]*schema.Schema {
	result := map[string]*schema.Schema{}

	for _, current := range maps {
		for key, value := range current {
			result[key] = value
		}
	}

	return result
}

func copyInputOutput(d *schema.ResourceData) error {
	return d.Set("outputs", d.Get("inputs"))
}

type PhasedContextFunc func(string, context.Context, *schema.ResourceData, interface{}) diag.Diagnostics

func phasedFunc(phase string, action PhasedContextFunc) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
		d.SetId("static")

		if phase == d.Get("phase") {
			return action(phase, ctx, d, i)
		}

		return diag.Diagnostics{}
	}
}

func processResourceDeleteFunc(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Diagnostics{}
}

func id() string {
	return strconv.FormatUint(rand.Uint64(), 16)
}
