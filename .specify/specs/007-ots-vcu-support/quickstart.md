# Quickstart: Using VCU Instances

## Requirements

- Terraform v0.12+
- `alicloud` provider v1.3.0+ (with this feature)

## Usage

```hcl
resource "alicloud_ots_instance" "default" {
  name                   = "tf-test-vcu-instance"
  description            = "terraform-test"
  instance_specification = "VCU"  # New value
  elastic_vcu_upper_limit = 2.5   # New field
  
  tags = {
    Created = "Terraform"
  }
}
```

## Update

Change `elastic_vcu_upper_limit` and apply.

```hcl
resource "alicloud_ots_instance" "default" {
  # ...
  elastic_vcu_upper_limit = 5.0
}
```
