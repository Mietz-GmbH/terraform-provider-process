---
layout: "process"
page_title: "Provider: Process"
sidebar_current: "docs-process-index"
description: |-
A provider for spawning and using external processes.
---

# Process Provider

The process provider gives the ability to spawn processes. The input and output of the processes can be
used. As well as processes living over the lifetime of other tasks.

Use the navigation to the left to read about the available resources.

## Usage

There is no need for a configuration block for this provider at all.

## Data sources

Every resource from this provider has a data source, which does the same thing.
The only difference is, that the data source will execute the action when reading,
not when creating.

```hcl
resource "process_run" "as_resource" {
  # ...
}

data "process_run" "as_datasource" {
  # ...
}
```

## `inputs` / `outputs`

All resources and data sources from this provider have a `inputs` and a `outputs` attribute.

The content of the `inputs` will be copied over to `outputs` after the resource
has been executed.

Example:
```hcl
resource "process_run" "io" {
  inputs = {
    hello = "world"
  }
  # ...
}

locals {
  hello = process_run.io.outputs.hello # world
}
```

This can be used to make other provider configurations dependent on the execution of a process.
