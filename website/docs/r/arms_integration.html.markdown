---
subcategory: "Application Real-Time Monitoring Service (ARMS)"
layout: "alicloud"
page_title: "Alicloud: alicloud_arms_integration"
sidebar_current: "docs-alicloud-resource-arms-integration"
description: |-
  Provides a Alicloud ARMS Integration resource.
---

# alicloud_arms_integration

Provides a ARMS Integration resource.

For information about ARMS Integration and how to use it, see [What is Integration](https://www.alibabacloud.com/help/en/arms/developer-reference/api-arms-2019-08-08-createintegration).

-> **NOTE:** Available since v1.212.0.

## Example Usage

Basic Usage

```terraform
resource "alicloud_arms_integration" "example" {
  integration_name = "example-webhook-integration"
  integration_type = "webhooks"
  config = jsonencode({
    url    = "https://example.com/webhook"
    secret = "webhook-secret-key"
  })
  description = "Example webhook integration for monitoring alerts"
  status      = "Active"
}
```

## Argument Reference

The following arguments are supported:

* `integration_name` - (Required, ForceNew) The name of the integration. The name must be unique within the region.
* `integration_type` - (Required, ForceNew) The type of the integration. Valid values: `cloudwatch`, `datadog`, `grafana`, `prometheus`, `webhooks`.
* `config` - (Required) The configuration of the integration. The configuration format varies by integration type. Should be a valid JSON string.
* `description` - (Optional) The description of the integration.
* `status` - (Optional) The status of the integration. Valid values: `Active`, `Inactive`. Default value: `Active`.

## Attributes Reference

The following attributes are exported:

* `id` - The resource ID in terraform of Integration. The value is same as `integration_id`.
* `integration_name` - The name of the integration.
* `integration_type` - The type of the integration.
* `config` - The configuration of the integration.
* `description` - The description of the integration.
* `status` - The status of the integration.
* `create_time` - The creation time of the integration.
* `update_time` - The last update time of the integration.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration-0-11/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 3 mins) Used when creating the Integration.
* `update` - (Defaults to 3 mins) Used when updating the Integration.
* `delete` - (Defaults to 3 mins) Used when deleting the Integration.

## Import

ARMS Integration can be imported using the id, e.g.

```shell
$ terraform import alicloud_arms_integration.example <id>
```