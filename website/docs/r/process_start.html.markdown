---
layout: "process"
page_title: "Process: process_start"
sidebar_current: "docs-process-process_start"
description: |-
Starts a process which runs in background.
---

# process_start

The ``process_start`` resource starts a process, which will keep running in background.

**Note:** check the ``process_end`` resources as well for best use of this resource.

## Usage

### Forwarding a TCP port
```hcl
resource "process_start" "forward_tcp" {
   command {
      platforms = ["linux", "darwin"]
      command   = "socat tcp-listen:443,reuseaddr,fork tcp:localhost:4000"
   }
}
```

## Argument Reference

The following arguments are supported:

* `triggers` - (Optional) When anything in this object changed, the process will
   be executed again.
* `command` - (Required) A configuration of a possible command (documented in `process_run`)
* `inputs` - (Optional) Copied into `outputs` (documented on main page)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `process_id` - A ID of the spawned process which can be passed to `process_end`.
* `pid` - The PID (Process Identifier) of the spawned process.
* `outputs` - Copied from `inputs` (documented on main page)
