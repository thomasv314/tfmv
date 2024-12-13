# tfscp (terraform state copy)
tfscp is a tool designed for migrating Terraform resources across remote states. It simplifies the process by leveraging Terraform's existing capabilities.

### How it works

tfscp automates the following steps:

1. Creates a temporary operating directory.
1. Validates that `terraform init` succeeds in both source and target state directories.
1. Downloads the source and target states locally.
1. Uses `terraform state mv` to move the specified resource from the source to the target state.
1. Pushes the updated target state back to the remote location using `terraform state push`.
1. Cleans up the temporary operating directory.

### Installation
To install tfscp, run the following commands:

```
wget https://github.com/thomasv314/tfscp/releases/download/v1.0.2/tfscp-darwin-amd64
mv tfscp-darwin-amd64 /usr/local/bin/tfscp
chmod u+x /usr/local/bin/tfscp
```

## Usage

```
Usage:
  tfscp [terraform resource address] [flags]

Flags:
      --dry-run             If true, do not actually move the resource (default true)
  -h, --help                help for tfscp
      --source-dir string   Source directory to move a terraform resource from (optional, defaults to CWD)
      --target-dir string   Target directory to move a terraform resource to
```

## Example scenario

Suppose you have the following Terraform states:

```
└── apps
    ├── applications.tf
    ├── config.tf       # Uses s3://terraform-states/apps.tfstate
    └── foobar-service
        ├── config.tf   # Uses s3://terraform-states/foobar-service.tfstate
        └── service.tf 
2 directories, 4 files
```

Here, the apps directory contains a module for your app: `module.foobar_service`.

By default, tfscp operates in a safe "dry-run" mode. To preview actions, run tfscp from the source directory:

```
tfscp "module.foobar_service" --target-dir="./foobar-service"
```

For example:

![Example of a dry run preparing to move/copy terraform resources across remote state files](https://github.com/user-attachments/assets/aaf33a0a-b894-4e38-952c-4e8ad0533068)

If the dry-run output looks good, execute the migration by disabling the dry-run mode:

```
 tfscp "module.foobar_service" --target-dir="./foobar-service" --dry-run=false
```

![Example of moving terraform resources across remote state files](https://github.com/user-attachments/assets/8edc8cfe-c161-40e8-b2a1-e97f7b2a58c4)
