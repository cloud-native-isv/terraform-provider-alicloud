---
subcategory: "Application Real-Time Monitoring Service (ARMS)"
layout: "alicloud"
page_title: "Alicloud: alicloud_arms_integrations"
sidebar_current: "docs-alicloud-datasource-arms-integrations"
description: |-
  Provides a list of ARMS Integrations to the user.
---

# alicloud_arms_integrations

This data source provides the ARMS Integrations of the current Alibaba Cloud user.

-> **NOTE:** Available since v1.212.0.

## Example Usage

Basic Usage

```terraform
data "alicloud_arms_integrations" "ids" {
  ids = ["example_id"]
}

output "arms_integration_id_1" {
  value = data.alicloud_arms_integrations.ids.integrations.0.id
}

data "alicloud_arms_integrations" "nameRegex" {
  name_regex = "^my-Integration"
}

output "arms_integration_id_2" {
  value = data.alicloud_arms_integrations.nameRegex.integrations.0.id
}

data "alicloud_arms_integrations" "type" {
  integration_type = "webhooks"
}

output "arms_integration_id_3" {
  value = data.alicloud_arms_integrations.type.integrations.0.id
}

data "alicloud_arms_integrations" "status" {
  status = "Active"
}

output "arms_integration_id_4" {
  value = data.alicloud_arms_integrations.status.integrations.0.id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, ForceNew, Computed) A list of Integration IDs.
* `name_regex` - (Optional, ForceNew) A regex string to filter results by Integration name.
* `integration_type` - (Optional, ForceNew) The type of the integration. Valid values: `cloudwatch`, `datadog`, `grafana`, `prometheus`, `webhooks`.
* `status` - (Optional, ForceNew) The status of the integration. Valid values: `Active`, `Inactive`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `names` - A list of Integration names.
* `integrations` - A list of ARMS Integrations. Each element contains the following attributes:
  * `id` - The ID of the Integration.
  * `integration_id` - The ID of the Integration.
  * `integration_name` - The name of the integration.
  * `integration_type` - The type of the integration.
  * `description` - The description of the integration.
  * `status` - The status of the integration.
  * `create_time` - The creation time of the integration.
  * `update_time` - The last update time of the integration.
  * `config` - The configuration of the integration.