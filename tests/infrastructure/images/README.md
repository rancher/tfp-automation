# Update Images

Boots a VM from an image that already exists on a provider, runs a script on it, snapshots the result into a **new** image, then destroys the VM. The new image is kept.

Currently supported providers: `aws` (creates an AMI via `aws_ami_from_instance`).

## Usage

```go
imageID, err := images.UpdateImage(t, "aws", "/path/to/script.sh", "my-image")
```

- The script is uploaded to the VM and run with `sudo`. If it exits non-zero, no image is created and the VM is cleaned up.
- A UTC timestamp is appended to the image name so every run is unique.
- The builder VM is created from the existing provider config, e.g. for AWS:

```yaml
terraform:
  provider: "aws"
  privateKeyPath: ""
  awsCredentials: {awsAccessKey: "", awsSecretKey: ""}
  awsConfig:
    ami: ""                  # the image to start from
    awsInstanceType: ""
    awsUser: ""              # ssh user for the base image
    awsKeyName: ""
    awsSubnetID: ""
    awsSecurityGroups: [""]
    awsRootSize: 100
    region: ""
    timeout: "5m"
```

Before the VM is destroyed, the image is removed from Terraform state (`terraform state rm`) because destroying an `aws_ami_from_instance` would deregister the AMI. Cleanup honors the `rancher.cleanup` setting like other infrastructure tests.

## Adding a provider

1. Implement `images.ImageResourceFunc` in `framework/set/resources/images/<provider>/` (run the script on the instance, create the image, emit an output named `image_id`).
2. Register it in `TunnelToImageProvider` in `framework/set/resources/images/images.go`.
3. Add `modules/images/<provider>/main.tf` (placeholder, like the other modules).
