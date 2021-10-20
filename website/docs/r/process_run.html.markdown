---
layout: "process"
page_title: "Process: process_run"
sidebar_current: "docs-process-process_run"
description: |-
Runs a process.
---

# process_run

The ``process_run`` resource runs a process, and returns its output and status.

## Usage

### Create a file
```hcl
locals {
  filename = "test.txt"
}

resource "process_run" "create_file" {
  command {
    platforms = ["linux", "darwin"]
    command   = "touch ${local.filename}"
  }
  
  command {
    platforms = ["windows"]
    command   = "New-Item ${local.filename}"
  }
}
```

### Check if a port is open
```hcl
resource "process_run" "check_port" {
  command {
    platforms = ["darwin", "linux"]
    command   = "nc -z 127.0.0.1 8080"
  }
  tries    = 10
  timeout  = 100
}
```

### Write file
```hcl
resource "process_run" "write_text" {
  command {
    platforms = ["linux", "darwin"]
    command   = "cat > info.txt"
    stdin     = "Hello World!"
  }
}

resource "process_run" "write_image" {
  command {
    platforms    = ["linux", "darwin"]
    command      = "cat > image.jpg"
    stdin_base64 = var.image_base64
  }
}
```

### Read file
```hcl
resource "process_run" "read_file" {
  command {
    platforms = ["linux", "darwin"]
    command   = "cat info.txt"
  }
}

locals {
  file_content = process_run.read_file.stdout
}
```

## Argument Reference

The following arguments are supported:

* `triggers` - (Optional) When anything in this object changed, the process will
   be executed again.
* `command` - (Required) A configuration of a possible command (documented below)
* `tries` - (Optional) The number of tries, it will run the process again when failed.
* `retry_interval` - (Optional, Default: `500`) The number of milliseconds to wait between the tries
* `timeout` - (Optional, Default: `500`) The maximum number of milliseconds after the process gets killed. The
timeout can be disabled by settings this argument to `0`.
* `fail_on_error` - (Optional, Default: `false`) If set to true, the resource will fail to create when the process
exited an exit code not equal 0.
* `inputs` - (Optional) Copied into `outputs` (documented on main page)

The `command` object supports the following:

* `platforms` - (Required) A set of platforms which are valid for this configuration.
Valid values are `darwin`, `freebsd`, `linux`, `openbsd`, `solaris`, and `windows`.
* `command` - (Required) The command which will be executed.
* `environment` - (Optional) A map of environment variables.
* `sensitive_environment` - (Optional) A map of environment variables which are sensitive.
* `interpreter` - (Optional) The command line which spawns the process. When not given, on unix-like systems
`["/bin/bash", "-c"]` and on Windows `["powershell.exe"]` will be assumed.
* `working_directory` - (Optional) The working directory to start the process in. By default, the root of the
project.
* `stdin` - (Optional) When given, this data will be piped into the stdin of the process. Conflicts with `stdin_base64`.
* `stdin_base64` - (Optional) When given, this data will be decoded as base64 and be piped into
the stdin of the process. Conflicts with `stdin`.

The first `command` block which has a matching platform will be used as configuration. All other blocks
are ignored. When there is no matching command configuration, this resource will fail to create.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `stdout` - The output in stdout of the last process
* `stderr` - The output in stderr of the last process
* `error` - True if the last process exited with an exit code not equal 0
* `timed_out` - True if the last process was killed because it took longer than the timeout
* `outputs` - Copied from `inputs` (documented on main page)
