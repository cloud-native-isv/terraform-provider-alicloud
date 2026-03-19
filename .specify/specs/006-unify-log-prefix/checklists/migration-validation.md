# Migration Validation Checklist

## Purpose
This checklist validates that the migration from `alicloud_sls_*` to `alicloud_log_*` naming follows the required approach and maintains functionality.

## Pre-Migration Validation

- [ ] All `alicloud_sls_*` resources identified and mapped correctly
- [ ] All `alicloud_sls_*` data sources identified and mapped correctly  
- [ ] Mapping document reflects actual provider registrations
- [ ] No hard-coded references to `alicloud_sls_*` exist in documentation
- [ ] Migration guide completed with clear instructions

## Migration Process Validation

- [ ] Provider registration updated to use `alicloud_log_*` names
- [ ] Legacy `alicloud_sls_*` registrations removed or marked deprecated
- [ ] Functions renamed appropriately (if applicable) to match new naming
- [ ] Import functionality updated to work with new resource names
- [ ] All tests pass with new naming convention
- [ ] Backwards compatibility maintained during transition period (if applicable)

## Post-Migration Validation

- [ ] New `alicloud_log_*` resources accessible and functional
- [ ] New `alicloud_log_*` data sources accessible and functional
- [ ] Legacy `alicloud_sls_*` resources/data sources return appropriate error with migration guidance
- [ ] Error messages include clear replacement instructions
- [ ] Documentation updated to reference `alicloud_log_*` names
- [ ] Examples updated to use new naming convention
- [ ] Integration tests confirm functionality equivalence

## Rollback Validation (if needed)

- [ ] Clear rollback procedures documented
- [ ] Ability to revert to previous state exists
- [ ] State compatibility considerations noted

## Final Verification

- [ ] All acceptance criteria met
- [ ] Migration complete and validated
- [ ] No broken references to old naming
- [ ] Users can successfully migrate existing configurations