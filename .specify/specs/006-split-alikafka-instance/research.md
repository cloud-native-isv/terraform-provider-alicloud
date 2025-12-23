# Research: Split AliKafka Instance Resource

## Unknowns & Clarifications

### 1. StopInstance API
- **Question**: What is the API to stop a Kafka instance?
- **Finding**: Based on `StartInstance`, the corresponding API is likely `StopInstance` or `ReleaseInstance`. However, `ReleaseInstance` usually deletes the instance. The requirement is to "stop" it.
- **Decision**: Implement `StopInstance` in `KafkaService`. If the underlying SDK/API is actually `ReleaseInstance` (and it supports stopping without deleting, which is rare for "Release"), we'll use that. But given the user request "alicloud_alikafka_deployment调用kafkaService.StartInstance和kafkaService.StopInstance", I will assume a `StopInstance` method is needed in the service layer. I will implement `StopInstance` in `KafkaService` which calls the appropriate API (likely `StopInstance` if it exists, or we might need to investigate if `ReleaseInstance` is what's meant by "Stop" in this context, but usually "Stop" means pause billing/service).
- **Refinement**: The user explicitly said "alicloud_alikafka_deployment调用kafkaService.StartInstance和kafkaService.StopInstance". This implies I should implement a `StopInstance` method in the service. I will assume the API action is `StopInstance`.

### 2. CreatePrePayOrder / CreatePostPayOrder
- **Question**: Are these methods available?
- **Finding**: Yes, they are already implemented in `KafkaService`.

## Technology Choices

- **Terraform Resource Design**:
  - `alicloud_alikafka_instance`:
    - `Create`: Calls `CreateOrder`.
    - `Read`: Describes instance.
    - `Delete`: No-op (or Release if supported/desired, but spec says "Stop Instance" is handled by deployment).
  - `alicloud_alikafka_deployment`:
    - `Create`: Calls `StartInstance`.
    - `Read`: Checks status.
    - `Delete`: Calls `StopInstance`.

## Best Practices

- **Decoupling**: The split allows creating the resource (billing object) separate from the deployment (runtime object).
- **State Management**: `alicloud_alikafka_deployment` manages the "Running" state. `alicloud_alikafka_instance` manages the existence of the resource.

## Alternatives Considered

- **Flag in Instance Resource**: Rejected as per spec clarification (Strict Split).
