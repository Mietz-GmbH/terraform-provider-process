---
layout: "process"
page_title: "Process: process_end"
sidebar_current: "docs-process-process_end"
description: |-
Ends a background process
---

# process_end

The ``process_end`` resource ends a process which was spawned by ``process_start``. It is highly
recommended using this resource. Not using it can cause Terraform to stop the provider early when
there is nothing left in the plan for it to do. This will kill all process which are spawned by
this provider as well. When placing a ``process_end`` resource which will be executed after the
process is not needed anymore, Terraform will be forced to keep this provider running and not
killing it and its child processes.

**Note:** It is not possible to mix resources and data sources in a pair of `process_start` and
`process_end`. Doing so, will cause in an error when creating the `process_end` instance.

## Usage

### Forwarding a TCP port
```hcl
resource "process_start" "forward_tcp" {
   command {
      platforms = ["linux", "darwin"]
      command   = "socat tcp-listen:443,reuseaddr,fork tcp:localhost:4000"
   }
}

resource "some_resource" "uses_tcp_port" {
   # ...
   depends_on = [process_start.forward_tcp]
}

resource "process_end" "forward_tcp_end" {
   process_id = process_start.forward_tcp.process_id

   depends_on = [some_resource.uses_tcp_port]
}
```

## Argument Reference

The following arguments are supported:

* `process_id` - (Required) The ID of the process from `process_start`.
* `kill` - (Optional, Default: `true`) If set to `true`, the process will be killed rather than waiting 
for the process to exit.
* `inputs` - (Optional) Copied into `outputs` (documented on main page)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `error` - True if the process exited with an exit code not equal 0
* `outputs` - Copied from `inputs` (documented on main page)
