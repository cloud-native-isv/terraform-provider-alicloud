# Quickstart: ADBPG Terraform Resources

**Spec**: [requirements.md](requirements.md) | **Plan**: [plan.md](plan.md)

## Scenario 1: Create an ADBPG Instance (User Story 1 - P1)

```hcl
resource "alicloud_adbpg_instance" "example" {
  engine_version    = "6.0"
  db_instance_class = "gpdb.group.segsdx1"
  db_instance_mode  = "StorageElastic"
  seg_node_num      = 4
  storage_size      = 50
  storage_type      = "cloud_essd"

  zone_id              = "cn-hangzhou-h"
  vpc_id               = alicloud_vpc.default.id
  vswitch_id           = alicloud_vswitch.default.id
  instance_network_type = "VPC"
  pay_type             = "PayAsYouGo"
  description          = "terraform-adbpg-example"

  tags = {
    Environment = "dev"
    Team        = "data-platform"
  }
}
```

## Scenario 2: Create an Account (User Story 2 - P1)

```hcl
resource "alicloud_adbpg_account" "admin" {
  db_instance_id      = alicloud_adbpg_instance.example.id
  account_name        = "admin_user"
  account_password    = "YourSecureP@ssw0rd"
  account_type        = "Super"
  account_description = "Admin account for ADBPG"
}

resource "alicloud_adbpg_account" "app" {
  db_instance_id      = alicloud_adbpg_instance.example.id
  account_name        = "app_user"
  account_password    = "AppP@ssw0rd123"
  account_type        = "Normal"
  account_description = "Application read-write account"
}
```

## Scenario 3: Create a Database (User Story 3 - P2)

```hcl
resource "alicloud_adbpg_database" "analytics" {
  db_instance_id = alicloud_adbpg_instance.example.id
  db_name        = "analytics_db"
  db_description = "Main analytics database"
  character_name = "UTF8"
}
```

## Scenario 4: Allocate Public Connection (User Story 4 - P2)

```hcl
resource "alicloud_adbpg_connection" "public" {
  db_instance_id           = alicloud_adbpg_instance.example.id
  connection_string_prefix = "adbpg-public-example"
}

output "connection_string" {
  value = alicloud_adbpg_connection.public.connection_string
}

output "port" {
  value = alicloud_adbpg_connection.public.port
}
```

## Scenario 5: Configure Security IP Whitelist (User Story 5 - P2)

```hcl
resource "alicloud_adbpg_security_ip_array" "office" {
  db_instance_id             = alicloud_adbpg_instance.example.id
  db_instance_ip_array_name  = "office_ips"
  security_ip_list           = "10.0.0.0/8,172.16.0.0/12"
}
```

## Scenario 6: Query Existing Instances (User Story 6 - P2)

```hcl
data "alicloud_adbpg_instances" "running" {
  status = "Running"

  tags = {
    Environment = "prod"
  }
}

output "instance_ids" {
  value = data.alicloud_adbpg_instances.running.instances[*].id
}
```

## Scenario 7: Query Accounts (User Story 7 - P3)

```hcl
data "alicloud_adbpg_accounts" "all" {
  db_instance_id = alicloud_adbpg_instance.example.id
}

output "account_names" {
  value = data.alicloud_adbpg_accounts.all.accounts[*].account_name
}
```

## Scenario 8: Configure Backup Policy (User Story 8 - P3)

```hcl
resource "alicloud_adbpg_backup_policy" "daily" {
  db_instance_id        = alicloud_adbpg_instance.example.id
  backup_retention_period = 7
  preferred_backup_period = "Monday,Wednesday,Friday"
  preferred_backup_time   = "02:00Z-03:00Z"
  enable_recovery_point   = true
}
```

## Scenario 9: Enable SSL

```hcl
resource "alicloud_adbpg_ssl" "enabled" {
  db_instance_id = alicloud_adbpg_instance.example.id
  ssl_enabled    = true
}
```

## Scenario 10: Query Available Zones

```hcl
data "alicloud_adbpg_zones" "available" {}

output "zones" {
  value = data.alicloud_adbpg_zones.available.zones[*].id
}
```

## Complete Example

```hcl
# VPC infrastructure
resource "alicloud_vpc" "default" {
  vpc_name   = "adbpg-vpc"
  cidr_block = "172.16.0.0/16"
}

resource "alicloud_vswitch" "default" {
  vpc_id     = alicloud_vpc.default.id
  cidr_block = "172.16.0.0/24"
  zone_id    = "cn-hangzhou-h"
}

# ADBPG Instance
resource "alicloud_adbpg_instance" "main" {
  engine_version        = "6.0"
  db_instance_class     = "gpdb.group.segsdx1"
  db_instance_mode      = "StorageElastic"
  seg_node_num          = 4
  storage_size          = 50
  storage_type          = "cloud_essd"
  zone_id               = "cn-hangzhou-h"
  vpc_id                = alicloud_vpc.default.id
  vswitch_id            = alicloud_vswitch.default.id
  instance_network_type = "VPC"
  pay_type              = "PayAsYouGo"
  description           = "production-analytics"
}

# Account
resource "alicloud_adbpg_account" "admin" {
  db_instance_id   = alicloud_adbpg_instance.main.id
  account_name     = "admin"
  account_password = var.admin_password
  account_type     = "Super"
}

# Database
resource "alicloud_adbpg_database" "app_db" {
  db_instance_id = alicloud_adbpg_instance.main.id
  db_name        = "app_analytics"
  character_name = "UTF8"
}

# Backup Policy
resource "alicloud_adbpg_backup_policy" "daily" {
  db_instance_id          = alicloud_adbpg_instance.main.id
  backup_retention_period = 7
  preferred_backup_period = "Monday,Wednesday,Friday"
  preferred_backup_time   = "02:00Z-03:00Z"
  enable_recovery_point   = true
}

# SSL
resource "alicloud_adbpg_ssl" "enabled" {
  db_instance_id = alicloud_adbpg_instance.main.id
  ssl_enabled    = true
}

# Security IP Whitelist
resource "alicloud_adbpg_security_ip_array" "app_servers" {
  db_instance_id            = alicloud_adbpg_instance.main.id
  db_instance_ip_array_name = "app_servers"
  security_ip_list          = "10.0.1.0/24,10.0.2.0/24"
}
```
