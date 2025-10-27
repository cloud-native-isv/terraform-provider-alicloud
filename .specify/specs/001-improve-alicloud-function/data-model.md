# Data Model: Improve AliCloud Function Compute Support

## Entity: Function

Represents an FC function resource.

### Fields
- FunctionName (string): The name of the function
- Description (string): Description of the function
- Runtime (string): Runtime environment (e.g., python3.9, nodejs14)
- Handler (string): Function handler
- Timeout (int32): Function timeout in seconds
- MemorySize (int32): Memory size in MB
- Environment (map[string]*string): Environment variables
- State (string): Current state of the function
- CreatedTime (string): Creation timestamp
- LastModifiedTime (string): Last modification timestamp

### Relationships
- Belongs to zero or more Layers (through runtime compatibility)
- Can have multiple Versions
- Can have multiple Aliases
- Can have multiple Triggers
- Can have AsyncInvokeConfig
- Can have ConcurrencyConfig
- Can have ProvisionConfig

## Entity: Layer

Represents an FC layer resource.

### Fields
- LayerName (string): The name of the layer
- Version (int32): Layer version
- Description (string): Description of the layer
- CompatibleRuntime ([]string): Compatible runtimes
- Code (LayerCode): Layer code information
- ACL (string): Access control list
- License (string): License information
- CreateTime (string): Creation timestamp

### Relationships
- Associated with Functions through CompatibleRuntime
- Has multiple versions

## Entity: Trigger

Represents an FC trigger resource.

### Fields
- TriggerName (string): The name of the trigger
- TriggerType (string): Type of trigger (e.g., oss, timer, http)
- Description (string): Description of the trigger
- Qualifier (string): Qualifier for the function version
- SourceArn (string): Source ARN
- TargetArn (string): Target ARN
- InvocationRole (string): Invocation role
- Status (string): Current status
- TriggerConfig (string): Trigger configuration
- TriggerId (string): Trigger ID
- CreatedTime (string): Creation timestamp
- LastModifiedTime (string): Last modification timestamp
- HttpTrigger (HttpTrigger): HTTP trigger information

## Entity: CustomDomain

Represents an FC custom domain resource.

### Fields
- DomainName (string): The domain name
- Protocol (string): Protocol (HTTP/HTTPS)
- CertificateConfig (CertificateConfig): Certificate configuration
- ApiVersion (string): API version
- RouteConfig (RouteConfig): Route configuration
- CreatedTime (string): Creation timestamp
- LastModifiedTime (string): Last modification timestamp

## Entity: Alias

Represents an FC alias resource.

### Fields
- AliasName (string): The name of the alias
- VersionId (string): Version ID
- Description (string): Description of the alias
- RoutingConfig (RoutingConfig): Routing configuration
- CreatedTime (string): Creation timestamp
- LastModifiedTime (string): Last modification timestamp

## Entity: FunctionVersion

Represents a version of an FC function.

### Fields
- FunctionName (string): The name of the function
- VersionId (string): Version ID
- Description (string): Description of the version
- CreatedTime (string): Creation timestamp
- LastModifiedTime (string): Last modification timestamp

## Entity: AsyncInvokeConfig

Represents asynchronous invocation configuration for FC functions.

### Fields
- FunctionName (string): The name of the function
- Qualifier (string): Qualifier for the function version
- AsyncTTL (int32): Async TTL
- DestinationConfig (DestinationConfig): Destination configuration
- MaxAsyncEventAgeInSeconds (int32): Max async event age in seconds
- MaxAsyncRetryAttempts (int32): Max async retry attempts
- CreatedTime (string): Creation timestamp
- LastModifiedTime (string): Last modification timestamp

## Entity: ConcurrencyConfig

Represents concurrency configuration for FC functions.

### Fields
- FunctionName (string): The name of the function
- Qualifier (string): Qualifier for the function version
- ReservedConcurrency (int32): Reserved concurrency
- ProvisionedConcurrency (int32): Provisioned concurrency
- LastModifiedTime (string): Last modification timestamp

## Entity: ProvisionConfig

Represents provision configuration for FC functions.

### Fields
- FunctionName (string): The name of the function
- Qualifier (string): Qualifier for the function version
- Target (int32): Target provision count
- ScheduledActions ([]ScheduledAction): Scheduled actions
- TargetTrackingPolicies ([]TargetTrackingPolicy): Target tracking policies
- LastModifiedTime (string): Last modification timestamp

## Validation Rules

1. FunctionName must be unique within an account
2. Runtime must be a supported value
3. Timeout must be within allowed range
4. MemorySize must be within allowed range
5. Environment variable names must follow naming conventions
6. Layer version must be positive
7. Trigger configurations must be valid JSON
8. Custom domain names must be valid domain names
9. Alias names must be unique per function
10. Version descriptions must not exceed maximum length

## State Transitions

### Function States
- Creating → Active
- Updating → Active
- Deleting → Deleted

### Trigger States
- Creating → Active
- Updating → Active
- Deleting → Deleted

### CustomDomain States
- Creating → Active
- Updating → Active
- Deleting → Deleted

### Alias States
- Creating → Active
- Updating → Active
- Deleting → Deleted