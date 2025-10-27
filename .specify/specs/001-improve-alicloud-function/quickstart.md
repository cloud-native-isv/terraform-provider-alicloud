# Quick Start: Improve AliCloud Function Compute Support

## Overview

This feature improves the AliCloud Function Compute support in the Terraform provider by completing and standardizing the implementation of all FC-related resources. After implementation, users will have reliable and consistent support for managing all aspects of Function Compute through Terraform.

## Prerequisites

- Terraform 0.12+
- AliCloud account with Function Compute permissions
- AliCloud credentials configured (via environment variables, credentials file, or IAM roles)

## Getting Started

1. **Update the provider**: Ensure you're using a version of the Terraform provider that includes these improvements.

2. **Configure the provider**:
```hcl
provider "alicloud" {
  region = "cn-hangzhou"
  # Credentials will be loaded from environment variables or credentials file
}
```

3. **Create an FC function**:
```hcl
resource "alicloud_fc_function" "my_function" {
  function_name = "my-function"
  description   = "My Function Compute function"
  runtime       = "python3.9"
  handler       = "index.handler"
  timeout       = 60
  memory_size   = 512
  
  environment_variables = {
    ENV_VAR1 = "value1"
    ENV_VAR2 = "value2"
  }
}
```

4. **Create an FC layer**:
```hcl
resource "alicloud_fc_layer_version" "my_layer" {
  layer_name         = "my-layer"
  description        = "My Function Compute layer"
  compatible_runtime  = ["python3.9", "python3.8"]
  
  code {
    zip_file = "path/to/layer.zip"
  }
}
```

5. **Create an FC trigger**:
```hcl
resource "alicloud_fc_trigger" "my_trigger" {
  function_name      = alicloud_fc_function.my_function.function_name
  trigger_name       = "my-trigger"
  trigger_type       = "timer"
  qualifier          = "LATEST"
  trigger_config     = jsonencode({
    cronExpression = "@hourly"
    enable         = true
  })
}
```

6. **Apply the configuration**:
```bash
terraform init
terraform plan
terraform apply
```

## Common Patterns

### Managing Function Versions
```hcl
resource "alicloud_fc_function_version" "my_function_version" {
  function_name = alicloud_fc_function.my_function.function_name
  description   = "Version 1.0"
}
```

### Creating Function Aliases
```hcl
resource "alicloud_fc_alias" "my_alias" {
  function_name = alicloud_fc_function.my_function.function_name
  alias_name    = "prod"
  version       = alicloud_fc_function_version.my_function_version.version
}
```

### Configuring Async Invocation
```hcl
resource "alicloud_fc_async_invoke_config" "my_async_config" {
  function_name = alicloud_fc_function.my_function.function_name
  qualifier     = alicloud_fc_alias.my_alias.alias_name
  
  destination_config {
    on_success {
      destination = "acs:fc:::functions/my-success-handler"
    }
    on_failure {
      destination = "acs:fc:::functions/my-failure-handler"
    }
  }
  
  max_async_event_age_in_seconds = 3600
  max_async_retry_attempts       = 3
}
```

## Best Practices

1. **Use meaningful names**: Choose descriptive names for your FC resources to make them easy to identify.

2. **Set appropriate timeouts**: Configure function timeouts based on your function's expected execution time.

3. **Manage environment variables carefully**: Use sensitive environment variables for secrets and avoid hardcoding them in configuration files.

4. **Version your functions**: Use function versions and aliases to manage different stages of your application (dev, test, prod).

5. **Configure error handling**: Use async invoke configurations to handle function execution failures gracefully.

6. **Monitor resource usage**: Set appropriate memory sizes and concurrency limits based on your function's resource requirements.

## Troubleshooting

### Common Issues

1. **Permission errors**: Ensure your AliCloud credentials have the necessary Function Compute permissions.

2. **Invalid configuration**: Check that all required fields are provided and have valid values.

3. **Resource limits**: Be aware of AliCloud Function Compute service limits and quotas.

### Getting Help

- Check the [AliCloud Function Compute documentation](https://www.alibabacloud.com/help/en/function-compute)
- Review Terraform provider documentation
- File issues on the Terraform provider GitHub repository