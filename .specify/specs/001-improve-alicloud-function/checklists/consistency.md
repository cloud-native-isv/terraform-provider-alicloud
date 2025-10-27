# FC Service Implementation Consistency Checklist: Improve AliCloud Function Compute Support

**Purpose**: Validate consistency of FC service implementations across all resource types
**Created**: October 27, 2025
**Feature**: [/cws_data/terraform-provider-alicloud/.specify/specs/001-improve-alicloud-function/spec.md](file:///cws_data/terraform-provider-alicloud/.specify/specs/001-improve-alicloud-function/spec.md)

## ID Encoding/Decoding Consistency

- [ ] CHK001 Are all Encode*Id functions consistently named with resource-specific prefixes? [Consistency, Spec §User Story 2]
- [ ] CHK002 Are all Decode*Id functions consistently named with resource-specific prefixes? [Consistency, Spec §User Story 2]
- [ ] CHK003 Do all Encode*Id functions follow the same parameter patterns? [Consistency, Gap]
- [ ] CHK004 Do all Decode*Id functions follow the same parameter patterns and return types? [Consistency, Gap]
- [ ] CHK005 Are ID encoding/decoding functions properly documented with format specifications? [Clarity, Gap]

## Describe Method Consistency

- [ ] CHK006 Are all Describe* methods consistently named with resource-specific prefixes? [Consistency, Spec §User Story 2]
- [ ] CHK007 Do all Describe* methods follow the same parameter validation patterns? [Consistency, Gap]
- [ ] CHK008 Do all Describe* methods return the same error handling patterns? [Consistency, Gap]
- [ ] CHK009 Are Describe* method signatures consistent across all FC service implementations? [Consistency, Gap]
- [ ] CHK010 Do all Describe* methods properly handle empty/nil parameter cases? [Completeness, Gap]

## CRUD Operation Consistency

- [ ] CHK011 Are Create* method signatures consistent across all FC service implementations? [Consistency, Spec §User Story 2]
- [ ] CHK012 Are Update* method signatures consistent across all FC service implementations? [Consistency, Spec §User Story 2]
- [ ] CHK013 Are Delete* method signatures consistent across all FC service implementations? [Consistency, Spec §User Story 2]
- [ ] CHK014 Do all Create* methods follow the same parameter validation patterns? [Consistency, Gap]
- [ ] CHK015 Do all Update* methods follow the same parameter validation patterns? [Consistency, Gap]
- [ ] CHK016 Do all Delete* methods follow the same parameter validation patterns? [Consistency, Gap]
- [ ] CHK017 Are error handling patterns consistent across all CRUD operations? [Consistency, Gap]

## State Refresh Function Consistency

- [ ] CHK018 Are all *StateRefreshFunc functions consistently named with resource-specific prefixes? [Consistency, Spec §User Story 2]
- [ ] CHK019 Do all *StateRefreshFunc functions follow the same pattern for status checking? [Consistency, Gap]
- [ ] CHK020 Are fail state handling patterns consistent across all *StateRefreshFunc implementations? [Consistency, Gap]
- [ ] CHK021 Do all *StateRefreshFunc functions properly handle IsNotFoundError cases? [Consistency, Gap]
- [ ] CHK022 Are the return value patterns consistent across all *StateRefreshFunc implementations? [Consistency, Gap]

## Wait Function Consistency

- [ ] CHK023 Are all WaitFor*Creating functions consistently named and structured? [Consistency, Spec §User Story 2]
- [ ] CHK024 Are all WaitFor*Updating functions consistently named and structured? [Consistency, Spec §User Story 2]
- [ ] CHK025 Are all WaitFor*Deleting functions consistently named and structured? [Consistency, Spec §User Story 2]
- [ ] CHK026 Do all WaitFor* functions use the same timeout and delay parameters? [Consistency, Gap]
- [ ] CHK027 Are the state transition patterns consistent across all WaitFor* functions? [Consistency, Gap]

## Schema Building Consistency

- [ ] CHK028 Are Build*InputFromSchema functions consistently named across all FC service implementations? [Consistency, Spec §User Story 2]
- [ ] CHK029 Are Build*UpdateInputFromSchema functions consistently named across all FC service implementations? [Consistency, Spec §User Story 2]
- [ ] CHK030 Do all Build*InputFromSchema functions follow the same pattern for handling schema data? [Consistency, Gap]
- [ ] CHK031 Do all Build*UpdateInputFromSchema functions follow the same pattern for handling schema changes? [Consistency, Gap]
- [ ] CHK032 Are the error handling patterns consistent in all Build*InputFromSchema functions? [Consistency, Gap]

## Schema Setting Consistency

- [ ] CHK033 Are SetSchemaFrom* functions consistently named across all FC service implementations? [Consistency, Spec §User Story 2]
- [ ] CHK034 Do all SetSchemaFrom* functions follow the same pattern for setting schema data? [Consistency, Gap]
- [ ] CHK035 Are the error handling patterns consistent in all SetSchemaFrom* functions? [Consistency, Gap]
- [ ] CHK036 Do all SetSchemaFrom* functions properly handle nil parameter cases? [Completeness, Gap]
- [ ] CHK037 Are computed field setting patterns consistent across all SetSchemaFrom* functions? [Consistency, Gap]

## Error Handling Consistency

- [ ] CHK038 Are error messages consistently formatted across all FC service implementations? [Consistency, Spec §FR-009]
- [ ] CHK039 Do all FC service methods use the same error wrapping patterns (WrapError, WrapErrorf)? [Consistency, Spec §FR-009]
- [ ] CHK040 Are NotFoundError handling patterns consistent across all FC service implementations? [Consistency, Spec §FR-009]
- [ ] CHK041 Are retry logic patterns consistent across all FC service implementations? [Consistency, Spec §FR-010]
- [ ] CHK042 Are the same error types used consistently for similar error conditions? [Consistency, Gap]

## API Integration Consistency

- [ ] CHK043 Do all FC service implementations consistently use the GetAPI() method for API access? [Consistency, Plan §Constitution Check]
- [ ] CHK044 Are cws-lib-go API method calls consistently wrapped across all FC service implementations? [Consistency, Plan §Constitution Check]
- [ ] CHK045 Are parameter conversion patterns consistent when calling cws-lib-go API methods? [Consistency, Gap]
- [ ] CHK046 Are response handling patterns consistent across all FC service implementations? [Consistency, Gap]
- [ ] CHK047 Are the same logging patterns used consistently across all FC service implementations? [Consistency, Gap]