---
subcategory: "ECS"
layout: "alicloud"
page_title: "Alicloud: alicloud_ecs_storage_capacity_unit"
sidebar_current: "docs-alicloud-resource-ecs-storage-capacity-unit"
description: |-
  Provides a Alicloud ECS Storage Capacity Unit resource.
---

# alicloud_ecs_storage_capacity_unit

Provides a ECS Storage Capacity Unit resource.

For information about ECS Storage Capacity Unit and how to use it, see [What is Storage Capacity Unit](https://www.alibabacloud.com/help/en/doc-detail/161157.html).

-> **NOTE:** Available since v1.155.0.

## Example Usage

Basic Usage

<div style="display: block;margin-bottom: 40px;"><div class="oics-button" style="float: right;position: absolute;margin-bottom: 10px;">
  <a href="https://api.aliyun.com/terraform?resource=alicloud_ecs_storage_capacity_unit&exampleId=e8005869-d687-d6d9-709f-72d43e40f3e2cd384489&activeTab=example&spm=docs.r.ecs_storage_capacity_unit.0.e8005869d6&intl_lang=EN_US" target="_blank">
    <img alt="Open in AliCloud" src="https://img.alicdn.com/imgextra/i1/O1CN01hjjqXv1uYUlY56FyX_!!6000000006049-55-tps-254-36.svg" style="max-height: 44px; max-width: 100%;">
  </a>
</div></div>

```terraform
resource "alicloud_ecs_storage_capacity_unit" "default" {
  capacity                   = 20
  description                = "tftestdescription"
  storage_capacity_unit_name = "tftestname"
}
```

## Argument Reference

The following arguments are supported:

* `capacity` - (Required, ForceNew) The capacity of the Storage Capacity Unit. Unit: GiB. Valid values: `20`, `40`, `100`, `200`, `500`, `1024`, `2048`, `5120`, `10240`, `20480`, and `51200`.
* `description` - (Optional) The description of the Storage Capacity Unit. The description must be 2 to 256 characters in length and cannot start with `http://` or `https://`.
* `period` - (Optional, Computed) The validity period of the Storage Capacity Unit. Default value: `1`.
  * When PeriodUnit is set to Month, Valid values: `1`, `2`, `3`, `6`.
  * When PeriodUnit is set to Year, Valid values: `1`, `3`, `5`.
* `period_unit` - (Optional, Computed) The unit of the validity period of the Storage Capacity Unit. Default value: `Month`. Valid values: `Month`, `Year`.
* `start_time` - (Optional, ForceNew, Computed) The time when the Storage Capacity Unit takes effect. It cannot be earlier than or more than six months later than the time when the Storage Capacity Unit is created. Specify the time in the ISO 8601 standard in the `yyyy-MM-ddTHH:mm:ssZ` format. The time must be in UTC. **NOTE:** This parameter is empty by default. The Storage Capacity Unit immediately takes effect after it is created.
* `storage_capacity_unit_name` - (Optional, Computed) The name of the Storage Capacity Unit.

## Attributes Reference

The following attributes are exported:

* `id` - The resource ID in terraform of Storage Capacity Unit.
* `status` - The status of Storage Capacity Unit.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when create the Storage Capacity Unit.

## Import

ECS Storage Capacity Unit can be imported using the id, e.g.

```shell
$ terraform import alicloud_ecs_storage_capacity_unit.example <id>
```
