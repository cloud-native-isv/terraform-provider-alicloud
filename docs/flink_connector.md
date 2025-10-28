# Flink Connector Resource

The `alicloud_flink_connector` resource allows you to manage custom Flink connectors in Alibaba Cloud.

## Example Usage

```hcl
resource "alicloud_flink_connector" "example" {
  workspace_id     = "your-workspace-id"
  namespace_name   = "your-namespace-name"
  connector_name   = "example-connector"
  connector_type   = "custom"
  jar_url          = "https://example.com/connectors/example-connector.jar"
  description      = "An example custom Flink connector"
  source           = true
  sink             = false
  lookup           = false
  supported_formats = ["json", "csv"]
  dependencies     = ["com.example:dependency:1.0.0"]
}
```

## Argument Reference

The following arguments are supported:

* `workspace_id` - (Required, ForceNew) The workspace ID where the Flink connector will be registered.
* `namespace_name` - (Required, ForceNew) The namespace name where the Flink connector will be registered.
* `connector_name` - (Required, ForceNew) The name of the Flink connector.
* `connector_type` - (Required, ForceNew) The type of the Flink connector.
* `jar_url` - (Required) The URL to the JAR file for the connector.
* `description` - (Optional) Description of the connector.
* `source` - (Optional) Whether the connector is a source connector. Defaults to `false`.
* `sink` - (Optional) Whether the connector is a sink connector. Defaults to `false`.
* `lookup` - (Optional) Whether the connector is a lookup connector. Defaults to `false`.
* `supported_formats` - (Optional) List of supported formats for the connector.
* `dependencies` - (Optional) List of dependencies for the connector.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the connector in the format `workspace_id:namespace_name:connector_name`.

## Import

Flink connectors can be imported using the ID in the format `workspace_id:namespace_name:connector_name`, e.g.

```bash
terraform import alicloud_flink_connector.example your-workspace-id:your-namespace-name:example-connector
```