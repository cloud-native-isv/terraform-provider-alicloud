# Quickstart: Split AliKafka Instance Resource

## Overview

This feature splits the `alicloud_alikafka_instance` resource into two:
1. `alicloud_alikafka_instance`: Creates the instance (order).
2. `alicloud_alikafka_deployment`: Deploys (starts) the instance.

## Usage

### Creating an Instance (Order Only)

```hcl
resource "alicloud_alikafka_instance" "default" {
  name          = "tf-alikafka-instance"
  partition_num = 50
  disk_type     = "1"
  disk_size     = 500
  deploy_type   = 5
}
```

### Deploying the Instance

```hcl
resource "alicloud_alikafka_deployment" "default" {
  instance_id = alicloud_alikafka_instance.default.id
  vswitch_id  = "vsw-123456"
  zone_id     = "cn-hangzhou-b"
}
```

## Migration

Existing users of `alicloud_alikafka_instance` will need to:
1. Add `alicloud_alikafka_deployment` resource to their configuration.
2. Move deployment-related arguments (`vswitch_id`, `zone_id`, etc.) to the deployment resource.
3. Run `terraform apply` to adopt the new structure.
