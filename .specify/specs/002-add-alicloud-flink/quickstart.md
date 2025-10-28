# Quickstart: alicloud_flink_connector

## Overview

This guide shows how to use the `alicloud_flink_connector` resource to manage custom Flink connectors in Alibaba Cloud.

## Prerequisites

1. Terraform 0.12+ installed
2. Alibaba Cloud account with appropriate permissions
3. Access Key ID and Secret Access Key
4. An existing Flink workspace and namespace

## Basic Usage

### 1. Provider Configuration

```hcl
provider "alicloud" {
  access_key = "your-access-key"
  secret_key = "your-secret-key"
  region     = "cn-hangzhou"
}
```

### 2. Basic Connector Registration

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

### 3. Importing an Existing Connector

```bash
terraform import alicloud_flink_connector.example your-workspace-id:your-namespace-name:example-connector
```

## Common Operations

### Creating a Connector

1. Define the resource in your Terraform configuration
2. Run `terraform apply` to create the connector

### Updating a Connector

1. Modify the resource properties in your Terraform configuration
2. Run `terraform apply` to update the connector
   Note: Only non-force-new properties can be updated without recreation

### Deleting a Connector

1. Remove the resource from your Terraform configuration or use `terraform destroy`
2. Run `terraform apply` or `terraform destroy` to delete the connector

## Attributes Reference

The following attributes are exported:

- `workspace_id` - The workspace ID
- `namespace_name` - The namespace name
- `connector_name` - The connector name
- `connector_type` - The connector type
- `jar_url` - The JAR URL
- `description` - The connector description
- `source` - Whether it's a source connector
- `sink` - Whether it's a sink connector
- `lookup` - Whether it's a lookup connector
- `supported_formats` - List of supported formats
- `dependencies` - List of dependencies

## Timeouts

The following timeouts can be configured:

- `create` - (Default: 10 minutes) Used for creating the connector
- `update` - (Default: 10 minutes) Used for updating the connector
- `delete` - (Default: 10 minutes) Used for deleting the connector

## Import

The resource supports import using the ID format: `workspace_id:namespace_name:connector_name`

```bash
terraform import alicloud_flink_connector.example workspace_id:namespace_name:connector_name
```