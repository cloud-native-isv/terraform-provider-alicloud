terraform {
  required_version = ">= 1.0.0"
}

provider "alicloud" {
  region = "cn-hangzhou"
}

data "alicloud_cms_service" "this" {
  enable = "On"
}

output "cms_service_status" {
  value = data.alicloud_cms_service.this.status
}
