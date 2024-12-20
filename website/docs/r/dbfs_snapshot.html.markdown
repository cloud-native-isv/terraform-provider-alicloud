---
subcategory: "Database File System (DBFS)"
layout: "alicloud"
page_title: "Alicloud: alicloud_dbfs_snapshot"
sidebar_current: "docs-alicloud-resource-dbfs-snapshot"
description: |-
  Provides a Alicloud Database File System (DBFS) Snapshot resource.
---

# alicloud_dbfs_snapshot

Provides a Database File System (DBFS) Snapshot resource.

For information about Database File System (DBFS) Snapshot and how to use it, see [What is Snapshot](https://help.aliyun.com/zh/dbfs/developer-reference/api-dbfs-2020-04-18-createsnapshot).

-> **NOTE:** Available since v1.156.0.

## Example Usage

Basic Usage

<div style="display: block;margin-bottom: 40px;"><div class="oics-button" style="float: right;position: absolute;margin-bottom: 10px;">
  <a href="https://api.aliyun.com/api-tools/terraform?resource=alicloud_dbfs_snapshot&exampleId=830f5098-7bea-1db0-0d73-c8d0b342bef9e07158ee&activeTab=example&spm=docs.r.dbfs_snapshot.0.830f50987b&intl_lang=EN_US" target="_blank">
    <img alt="Open in AliCloud" src="https://img.alicdn.com/imgextra/i1/O1CN01hjjqXv1uYUlY56FyX_!!6000000006049-55-tps-254-36.svg" style="max-height: 44px; max-width: 100%;">
  </a>
</div></div>

```terraform
variable "name" {
  default = "terraform-example"
}

provider "alicloud" {
  region = "cn-hangzhou"
}

data "alicloud_dbfs_instances" "default" {
}

resource "alicloud_dbfs_snapshot" "example" {
  instance_id    = data.alicloud_dbfs_instances.default.instances.0.id
  retention_days = 50
  snapshot_name  = var.name
  description    = "DbfsSnapshot"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, ForceNew) The ID of the Database File System.
* `retention_days` - (Optional, ForceNew, Int) The retention period of the snapshot. Valid values: `1` to `65536`.
* `snapshot_name` - (Optional) The name of the snapshot. The `snapshot_name` must be `2` to `128` characters in length. It must start with a large or small letter or Chinese, and cannot start with `http://`, `https://`, `auto` or `dbfs-auto`. It can contain numbers, colons (:), underscores (_), or hyphens (-). **NOTE:** From version 1.234.0, `snapshot_name` can be modified.
* `description` - (Optional) The description of the snapshot. The `description` must be `2` to `256` characters in length. It cannot start with `http://` or `https://`. **NOTE:** From version 1.234.0, `description` can be modified.
* `force` - (Optional, Bool) Specifies whether to force delete the snapshot. Valid values:
  - `true`: Enable.
  - `false`: Disable.

## Attributes Reference

The following attributes are exported:

* `id` - The resource ID in terraform of Snapshot.
* `status` - The status of the Snapshot.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration-0-11/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when create the Snapshot.
* `update` - (Defaults to 5 mins) Used when update the Snapshot.
* `delete` - (Defaults to 1 mins) Used when delete the Snapshot.

## Import

Database File System (DBFS) Snapshot can be imported using the id, e.g.

```shell
$ terraform import alicloud_dbfs_snapshot.example <id>
```
