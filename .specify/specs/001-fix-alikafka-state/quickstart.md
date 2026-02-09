# Quickstart: Debugging Alikafka State

To verify the state fix, you can run an acceptance test or manually verify with Terraform.

## Prerequisites

- Terraform v0.12+ installed
- Alibaba Cloud credentials configured (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`, `ALICLOUD_REGION`)

## Steps

1. **Build the Provider**:
   ```bash
   make build
   ```

2. **Run Acceptance Test**:
   Current tests should be enhanced to check for non-empty fields.
   ```bash
   make testacc TEST=TestAccAlicloudAlikafkaInstance_basic
   ```

3. **Manual Verification**:
   Create a `main.tf`:
   ```hcl
   resource "alicloud_alikafka_instance" "default" {
     name          = "tf-test-kafka"
     partition_num = 50
     disk_type     = "1"
     disk_size     = 500
     deploy_type   = 5
   }
   
   output "endpoints" {
     value = {
       domain = alicloud_alikafka_instance.default.domain_endpoint
       ssl    = alicloud_alikafka_instance.default.ssl_domain_endpoint
     }
   }
   ```
   Run `terraform apply` and check outputs.

## Expected Outcome

The outputs should contain valid endpoint strings (e.g., `alikafka-pre-cn-...`) instead of empty strings or nulls.
