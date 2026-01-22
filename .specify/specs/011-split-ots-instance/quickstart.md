# Quickstart: Split OTS Instance Resources

## Creating a VCU Instance

```hcl
resource "alicloud_ots_instance_vcu" "default" {
  name        = "tf-test-vcu-instance"
  description = "Created via terraform vcu resource"
  elastic_vcu_upper_limit = 2.0
  
  tags = {
    Created = "Terraform"
    Type    = "VCU"
  }
}
```

## Creating a Standard Instance

```hcl
resource "alicloud_ots_instance" "default" {
  name                   = "tf-test-ssd-instance"
  description            = "Created via terraform legacy resource"
  instance_specification = "SSD" # or "HYBRID"
  
  tags = {
    Created = "Terraform"
    Type    = "Standard"
  }
}
```

## Migration Note

If you have existing "VCU" instances managed by `alicloud_ots_instance`, you **MUST** migrate them:

1. Remove the old resource from your `.tf` file.
2. Add the new `alicloud_ots_instance_vcu` resource to your `.tf` file matching the real configuration.
3. Run state move:
   ```bash
   terraform state mv alicloud_ots_instance.old_name alicloud_ots_instance_vcu.new_name
   ```
4. Run `terraform plan` to verify no changes are required (clean state match).
