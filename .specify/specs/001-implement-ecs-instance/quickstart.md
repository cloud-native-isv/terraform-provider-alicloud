# Quickstart: alicloud_ecs_instance

## Prerequisite
Ensure you have `ALICLOUD_ACCESS_KEY` and `ALICLOUD_SECRET_KEY` set.

## Usage

```hcl
resource "alicloud_ecs_instance" "example" {
  image_id        = "ubuntu_18_04_x64_20G_alibase_20240528.vhd"
  instance_type   = "ecs.t5-lc1m1.small"
  security_groups = ["sg-12345678"]
  vswitch_id      = "vsw-12345678"
  
  instance_name = "tf-test-instance"
  description   = "Created by alicloud_ecs_instance"
  
  tags = {
    Env = "Test"
  }
}
```

## Validate
Run `terraform apply` and check the console.
