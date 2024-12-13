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

```
$ cd ~/terraform/apps
$ tfscp "module.foobar_service" --target-dir="./foobar-service"
Moving resource module.foobar_service:
  From state /Users/thomasvendetta/code/example-terraform-repo/apps
  To state   /Users/thomasvendetta/code/example-terraform-repo/apps/foobar-service

[dry-run] enabled. Not actually making moves.

==> Initializing source state
[dry-run] Would run 'terraform init':
  from /Users/thomasvendetta/code/example-terraform-repo/apps

==> Pulling source state
[dry-run] Would run 'terraform state pull':
  from /Users/thomasvendetta/code/example-terraform-repo/apps
  write to /operating-dir/terraform.tfstate

==> Initializing target state
[dry-run] Would run 'terraform init':
  from /Users/thomasvendetta/code/example-terraform-repo/apps/foobar-service

==> Pulling target state
[dry-run] Would run 'terraform state pull':
  from /Users/thomasvendetta/code/example-terraform-repo/apps/foobar-service
  write to /operating-dir/target-state.tfstate

==> Moving resource from source to target in local state
[dry-run] Would run 'terraform state mv --state-out=/operating-dir/target-state.tfstate'

==> Copying updated state to target directory
[dry-run] Would copy mutated state file over to target state directory
  from /operating-dir/target-state.tfstate
  to /Users/thomasvendetta/code/example-terraform-repo/apps/foobar-service/target-state.tfstate

==> Pushing updated target state.
[dry-run] Would run 'terraform state push target-state.tfstate' from /Users/thomasvendetta/code/example-terraform-repo/apps/foobar-service

==> Successfully moved!
```

If the dry-run output looks good, execute the migration by disabling the dry-run mode with `--dry-run=false`:

```
$ tfscp "module.foobar_service" --target-dir="./foobar-service" --dry-run=false             
Moving resource module.foobar_service:
  From state /Users/thomasvendetta/code/example-terraform-repo/apps
  To state   /Users/thomasvendetta/code/example-terraform-repo/apps/foobar-service
Do you want to continue? (yes/no): yes
==> Initializing source state
==> Pulling source state
==> Initializing target state
==> Pulling target state
==> Moving resource from source to target in local state
Move "module.foobar_service" to "module.foobar_service"
Successfully moved 1 object(s).
==> Copying updated state to target directory
==> Pushing updated target state.
==> Successfully moved!
```
