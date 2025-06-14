# SLS Service File Splitting Plan

## Current File Analysis
- File: `/cws_data/terraform-provider-alicloud/alicloud/service_alicloud_sls.go`
- Current size: ~1500+ lines
- Target: Split into multiple files, each under 1000 lines

## Splitting Strategy

### 1. service_alicloud_sls_base.go (~200 lines)
- [x] SlsService struct definition
- [x] NewSlsService and NewSlsServiceV2 constructors
- [x] Package imports and basic types

### 2. service_alicloud_sls_project.go (~300 lines)
- [ ] DescribeSlsProject
- [ ] DescribeListTagResources
- [ ] SlsProjectStateRefreshFunc
- [ ] SetResourceTags
- [ ] CreateProject, UpdateProject, DeleteProject
- [ ] ChangeResourceGroup, UpdateProjectPolicy
- [ ] SlsLogging functions (Create, Update, Delete, Get)

### 3. service_alicloud_sls_logstore.go (~300 lines)
- [ ] DescribeSlsLogStore
- [ ] DescribeGetLogStoreMeteringMode
- [ ] DescribeSlsLogStoreIndex
- [ ] SlsLogStoreStateRefreshFunc
- [ ] Related helper functions

### 4. service_alicloud_sls_job.go (~400 lines)
- [ ] DescribeSlsAlert + SlsAlertStateRefreshFunc
- [ ] DescribeSlsScheduledSQL + SlsScheduledSQLStateRefreshFunc
- [ ] DescribeSlsEtl + SlsEtlStateRefreshFunc
- [ ] DescribeSlsOssExportSink + SlsOssExportSinkStateRefreshFunc
- [ ] DescribeSlsIngestion + SlsIngestionStateRefreshFunc

### 5. service_alicloud_sls_config.go (~400 lines)
- [ ] DescribeSlsCollectionPolicy + SlsCollectionPolicyStateRefreshFunc
- [ ] DescribeSlsMachineGroup + SlsMachineGroupStateRefreshFunc
- [ ] DescribeSlsLogtailConfig + SlsLogtailConfigStateRefreshFunc
- [ ] DescribeSlsLogtailAttachment + SlsLogtailAttachmentStateRefreshFunc
- [ ] DescribeSlsLogAlertResource
- [ ] DescribeSlsDashboard + SlsDashboardStateRefreshFunc
- [ ] DescribeSlsOssShipper
- [ ] DescribeSlsResource + SlsResourceStateRefreshFunc
- [ ] DescribeSlsResourceRecord + SlsResourceRecordStateRefreshFunc

### 6. service_alicloud_sls_legacy.go (~400 lines)
- [ ] LogService struct (legacy compatibility)
- [ ] NewLogService constructor
- [ ] All legacy wrapper functions (DescribeLogProject, DescribeLogStore, etc.)
- [ ] All legacy Wait functions (WaitForLogProject, WaitForLogStore, etc.)

## Implementation Steps
1. ✅ Create TODO.md plan
2. [ ] Create service_alicloud_sls_base.go
3. [ ] Create service_alicloud_sls_project.go
4. [ ] Create service_alicloud_sls_logstore.go
5. [ ] Create service_alicloud_sls_job.go
6. [ ] Create service_alicloud_sls_config.go
7. [ ] Create service_alicloud_sls_legacy.go
8. [ ] Update original file to only include base content
9. [ ] Test compilation
10. [ ] Verify no functionality is lost

## Notes
- Maintain all existing function signatures for backward compatibility
- Ensure proper imports in each file
- Keep related functions together (e.g., Describe + StateRefresh functions)
- All files should have proper package declaration and imports