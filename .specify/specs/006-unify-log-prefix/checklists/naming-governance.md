# Naming Governance Checklist: 统一日志服务资源前缀

**Purpose**: Validate requirement quality for versioning policy and naming-scope governance before implementation
**Created**: 2026-03-19
**Feature (Requirement Scope)**: [.specify/specs/006-unify-log-prefix](..)
**Requirements (What)**: [requirements.md](../requirements.md)
**Specifications (How, context only)**: [plan.md](../plan.md), [tasks.md](../tasks.md)

**Note**: This checklist validates the quality of Requirements (What). Specification references are used only for traceability and gap detection.

## Requirement Completeness

- [ ] CHK001 Are deprecation timing requirements complete enough to cover release type boundaries (major/minor/patch), not only “recent release”? [Completeness, Req §FR-010, Req §Clarifications]
- [ ] CHK002 Are requirements explicitly documented for how users learn canonical replacements when every `alicloud_sls_*` item is in scope? [Completeness, Req §FR-002, Req §FR-003, Req §FR-009]
- [ ] CHK003 Are requirement statements complete on what constitutes an “unmappable” legacy entry and required fallback wording? [Completeness, Req §FR-002, Req §Edge Cases]
- [ ] CHK004 Are governance-scope requirements complete for inclusion, exclusion, and pending states without leaving implicit exceptions? [Completeness, Req §FR-006, Req §FR-009]

## Requirement Clarity

- [ ] CHK005 Is “按最近发布版本生效” defined with unambiguous release semantics to avoid interpretation drift across teams? [Clarity, Ambiguity, Req §FR-010]
- [ ] CHK006 Is “明确错误与替代项提示” specified with minimum required message fields (legacy name, canonical name, action hint)? [Clarity, Req §FR-007]
- [ ] CHK007 Is “一一对应或明确聚合关系” constrained with explicit criteria for when many-to-one mappings are acceptable? [Clarity, Req §FR-002]
- [ ] CHK008 Is “统一命名文档可定位目标能力” clarified with objective discoverability expectations (indexing/search path/lookup flow)? [Clarity, Req §FR-003, Gap]

## Requirement Consistency

- [ ] CHK009 Do compatibility requirements remain internally consistent between “强制统一立即停用” and migration-risk guidance expectations? [Consistency, Conflict, Req §Clarifications, Req §FR-004]
- [ ] CHK010 Do requirement-level versioning statements align with project governance constraints on release versioning policy? [Consistency, Conflict, Req §FR-010, Spec constitution.md §V]
- [ ] CHK011 Are assumptions about “不提供 state 迁移” consistent with all migration acceptance scenarios and edge-case narratives? [Consistency, Req §FR-008, Req §User Story 2, Req §Edge Cases]
- [ ] CHK012 Are boundary requirements consistent with success criteria so that every in-scope legacy name has either replacement mapping or retirement note? [Consistency, Req §FR-009, Req §SC-005]

## Acceptance Criteria Quality

- [ ] CHK013 Is SC-003 measurable with explicit sample definition, cohort size, and timing window rather than only percentage/time targets? [Acceptance Criteria, Measurability, Req §SC-003, Req §Measurement Sources]
- [ ] CHK014 Is SC-004 measurable with explicit baseline period and label taxonomy to make “咨询下降 50%” objectively auditable? [Acceptance Criteria, Measurability, Req §SC-004, Req §Measurement Sources]
- [ ] CHK015 Are success criteria traceably mapped to specific requirement IDs to avoid orphan metrics? [Traceability, Req §FR-001~FR-010, Req §SC-001~SC-006]

## Scenario & Edge Coverage

- [ ] CHK016 Are alternate scenarios defined for cases where a legacy name maps to multiple canonical capabilities? [Coverage, Alternate Flow, Req §FR-002, Req §Edge Cases]
- [ ] CHK017 Are exception scenarios specified for users upgrading with residual `alicloud_sls_*` usage in shared automation modules? [Coverage, Exception Flow, Req §FR-007, Req §Edge Cases]
- [ ] CHK018 Are recovery requirements defined for failed destroy-recreate migration steps, including rollback decision criteria? [Coverage, Recovery Flow, Gap, Req §FR-004, Req §FR-008]
- [ ] CHK019 Are non-functional requirements specified for migration communication quality (timeliness, accuracy, audience coverage)? [Non-Functional, Gap, Req §FR-005, Req §SC-006]

## Dependencies & Assumptions

- [ ] CHK020 Are external dependency assumptions (release-note process, issue tagging, support workflow) explicitly documented as requirements-level constraints? [Dependencies, Assumption, Req §Measurement Sources, Gap]
- [ ] CHK021 When tasks prescribe specific artifacts, do the requirements explicitly justify those artifacts or mark them as implementation choices? [Traceability, Spec tasks.md §Phase 1/2/6, Gap]
- [ ] CHK022 Do plan-level policy choices have explicit requirement references, especially around immediate deprecation behavior? [Traceability, Spec plan.md §Constraints, Req §FR-010, Conflict]

## Ambiguities & Conflicts

- [ ] CHK023 Is terminology stabilized between “停用”“下线”“停止入口”“统一入口” to prevent multi-meaning interpretation? [Ambiguity, Req §FR-004, Req §FR-007]
- [ ] CHK024 Are any requirement statements implicitly contradicting constitutional governance constraints explicitly resolved in requirements text? [Conflict, Req §FR-010, Spec constitution.md §V]

## Notes

- Check items off as completed: `[x]`
- Record findings inline under each item when failing
- Use this checklist as requirement-quality gate before `/speckit.implement`
