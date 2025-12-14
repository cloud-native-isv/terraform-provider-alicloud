# Quickstart: SelectDB Resources

This guide shows how to use the updated SelectDB resources.

## Prerequisites

- Terraform v0.13+
- Alibaba Cloud Provider

## Example Usage

### Create a SelectDB Instance

```hcl
resource "alicloud_selectdb_instance" "default" {
  instance_name       = "tf-test-selectdb"
  instance_class      = "selectdb.xlarge"
  cache_size          = 200
  charge_type         = "PostPaid"
  zone_id             = "cn-hangzhou-i"
  vpc_id              = "vpc-bp1..."
  vswitch_id          = "vsw-bp1..."
  engine              = "selectdb"
  engine_version      = "3.0"
  deploy_scheme       = "single_az"
  
  username            = "admin_test"
  password            = "YourPassword123!"
  
  security_ip_groups {
    group_name       = "default"
    security_ip_list = ["127.0.0.1"]
  }
  
  tags = {
    Created = "Terraform"
    Env     = "Test"
  }
}
```

### Create a SelectDB Cluster

```hcl
resource "alicloud_selectdb_cluster" "default" {
  instance_id    = alicloud_selectdb_instance.default.id
  cluster_name   = "tf-test-cluster"
  description    = "Terraform test cluster"
  cluster_class  = "selectdb.2xlarge"
  cache_size     = 400
  charge_type    = "PostPaid"
  zone_id        = "cn-hangzhou-i"
  vpc_id         = "vpc-bp1..."
  vswitch_id     = "vsw-bp1..."
  engine         = "selectdb"
  engine_version = "3.0"
  
  params {
    name  = "doris_scanner_thread_pool_thread_num"
    value = "48"
  }
}
```

### Query Instances and Clusters

```hcl
data "alicloud_selectdb_instances" "ids" {
  ids = [alicloud_selectdb_instance.default.id]
}

output "selectdb_instance_id" {
  value = data.alicloud_selectdb_instances.ids.instances.0.id
}

data "alicloud_selectdb_clusters" "ids" {
  ids = [alicloud_selectdb_cluster.default.id]
}

output "selectdb_cluster_id" {
  value = data.alicloud_selectdb_clusters.ids.clusters.0.id
}
```
