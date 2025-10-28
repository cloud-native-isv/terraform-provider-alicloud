# Data Model: alicloud_flink_connector

## Entity: Flink Connector

Represents a custom connector in Alibaba Cloud Flink service with properties for data integration.

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| workspace_id | string | Yes | The workspace ID where the Flink connector will be registered |
| namespace_name | string | Yes | The namespace name where the Flink connector will be registered |
| connector_name | string | Yes | The name of the Flink connector |
| connector_type | string | Yes | The type of the Flink connector |
| jar_url | string | Yes | The URL to the JAR file for the connector |
| description | string | No | Description of the connector |
| source | boolean | No | Whether the connector is a source connector (default: false) |
| sink | boolean | No | Whether the connector is a sink connector (default: false) |
| lookup | boolean | No | Whether the connector is a lookup connector (default: false) |
| supported_formats | list of strings | No | List of supported formats for the connector |
| dependencies | list of strings | No | List of dependencies for the connector |

### Validation Rules

1. workspace_id: Must be a non-empty string
2. namespace_name: Must be a non-empty string
3. connector_name: Must be a non-empty string
4. connector_type: Must be a non-empty string
5. jar_url: Must be a non-empty string
6. supported_formats: Each item must be a string
7. dependencies: Each item must be a string

### State Transitions

1. Creating → Available
2. Available → Deleting
3. Available → Modifying → Available (for updates)
4. Any state → Failed (error state)

### ID Format

The resource ID follows the format: `workspace_id:namespace_name:connector_name`

This format is used for:
- Resource identification in Terraform state
- Import operations
- Internal API calls