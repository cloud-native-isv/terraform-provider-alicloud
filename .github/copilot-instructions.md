# GitHub Copilot Instructions for Terraform Provider Alicloud

## Project Overview
This repository contains a custom implementation of Terraform Provider for Alibaba Cloud (Alicloud). It extends the official provider with custom modifications and enhancements to better serve specific use cases.

### Key Differences from Official Provider
1. **Custom SDK Implementations**: Local SDK versions in the `sdk/` directory replace some official Alibaba Cloud SDKs to implement custom modifications and bug fixes
2. **Additional Resources & Data Sources**: Implementation of terraform resources and data sources not available in the official provider
3. **Custom Resource Behavior**: Modified behavior for certain resources and data sources to better suit specific requirements

## Development Guidelines

### SDK Customization
- When working with SDK code, note that local implementations in `sdk/` directory take precedence over official ones
- Current custom SDK implementations include:
  - `alikafka-20190916`: Custom Message Queue for Apache Kafka SDK
  - `foasconsole-20211028`: Custom FOAS Console SDK
  - `ons-20190214`: Custom ONS SDK
  - `ververica-20220718`: Custom Ververica SDK

### Resource Customization
When modifying or creating resources:
1. Ensure compatibility with the existing provider structure
2. Follow the established naming conventions for resources and data sources
3. Update both resource implementation and corresponding documentation

## Common Tasks

### Adding a New Custom SDK
1. Add the SDK code to the `sdk/` directory
2. Update import paths in provider code to use the local SDK instead of the official one
3. Test thoroughly to ensure compatibility

### Modifying an Existing Resource
1. Identify the resource implementation in `alicloud/resource_alicloud_*.go`
2. Make necessary changes following the established patterns
3. Update tests and documentation
4. Verify with acceptance tests

### Adding a New Resource or Data Source
1. Create implementation files following naming conventions
2. Add resource/data source registration in `alicloud/provider.go`
3. Create corresponding documentation in `website/docs/`
4. Add acceptance tests

## Code Organization
- `alicloud/`: Main provider implementation, including resources and data sources
- `sdk/`: Custom SDK implementations
- `website/`: Documentation
- `examples/`: Example configurations for using the provider

## Best Practices
1. Avoid breaking changes unless absolutely necessary
2. Maintain backward compatibility with existing configurations
3. Keep modifications focused on specific functionality needs
4. Document all changes and custom behaviors
5. Test thoroughly before submitting changes