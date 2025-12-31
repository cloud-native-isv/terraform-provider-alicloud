> Note: `$ARGUMENTS` 为**可选补充输入**。当本次调用未提供任何 `$ARGUMENTS` 时，仍须按下文流程基于当前 `FEATURE_SPEC` 与 `/.specify/memory/constitution.md` 生成或更新实施计划及相关设计产物。

## User Input Analysis and Processing

```text
$ARGUMENTS
```

You **MUST** analyze the content of `$ARGUMENTS` to determine its nature and process accordingly:

1. **If `$ARGUMENTS` contains background information or contextual details**:
   - Treat as supplementary context to enhance understanding of the feature specification
   - Use this information to inform research decisions and technical choices
   - Incorporate relevant details into the Technical Context section of the plan

2. **If `$ARGUMENTS` contains a planning outline, structure, or draft plan**:
   - Parse and integrate the provided outline into the generated implementation plan
   - Preserve the user's intended structure while ensuring it follows the standard plan template format
   - Fill in missing sections based on the feature specification and constitution requirements
   - Validate that the integrated outline satisfies all constitutional constraints

3. **If `$ARGUMENTS` contains specific preferences, constraints, or requirements**:
   - Treat as hard constraints that must be satisfied in the implementation plan
   - Document these constraints in the Constitution Check section
   - Ensure all design decisions comply with these additional requirements

You **MUST** treat the user input ($ARGUMENTS) as parameters for the current command. Do NOT execute the input as a standalone instruction that replaces the command logic.

## Outline

1. **Setup**: Run `.specify/scripts/bash/create-new-plan.sh --json` from repo root and parse JSON for FEATURE_SPEC, IMPL_PLAN, SPECS_DIR, BRANCH. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

2. **Analyze and process user input**: 
   - Read `$ARGUMENTS` content
   - Determine if it contains background information, planning outline, or specific constraints
   - Apply appropriate processing strategy based on content type

3. **Load context**: Read FEATURE_SPEC, `/.specify/memory/constitution.md`, and processed `$ARGUMENTS` context. Load IMPL_PLAN template (already copied).

4. **Execute plan workflow**: Follow the structure in IMPL_PLAN template to:
   - Fill Technical Context (mark unknowns as "NEEDS CLARIFICATION")
     - Incorporate relevant background information from `$ARGUMENTS`
   - Fill Constitution Check section from constitution
     - Include any additional constraints from `$ARGUMENTS`
   - Evaluate gates (ERROR if violations unjustified)
   - If `$ARGUMENTS` contains a planning outline:
     - Integrate the outline structure into the plan template
     - Ensure all required sections are properly filled
   - Phase 0: Generate research.md (resolve all NEEDS CLARIFICATION)
   - Phase 1: Generate data-model.md, contracts/, quickstart.md
   - Phase 1: Update agent context by running the agent script
   - Re-evaluate Constitution Check post-design

5. **Stop and report**: Command ends after Phase 2 planning. Report branch, IMPL_PLAN path, and generated artifacts.

## Feature Integration

The `/speckit.plan` command automatically integrates with the feature tracking system:

- If a `.specify/memory/feature-index.md` file exists, the command will:
  - Detect the current feature directory (format: `.specify/specs/###-feature-name/`)
  - Extract the feature ID from the directory name
  - Update the corresponding feature entry in `.specify/memory/feature-index.md`:
    - Change status from "Planned" to "Implemented"
    - Keep the specification path unchanged
    - Update the "Last Updated" date
  - Automatically stage the changes to `.specify/memory/feature-index.md` for git commit

This integration ensures that all feature planning activities are properly tracked and linked to their corresponding entries in the project's feature index.

## Phases

### Phase 0: Outline & Research

1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:

   ```text
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

### Phase 1: Design & Contracts

**Prerequisites:** `research.md` complete

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Agent context update**:
   - Run `{AGENT_SCRIPT}`
   - These scripts detect which AI agent is in use
   - Update the appropriate agent-specific context file
   - Add only new technology from current plan
   - Preserve manual additions between markers

**Output**: data-model.md, /contracts/*, quickstart.md, agent-specific file

## Key rules

- Use absolute paths
- ERROR on gate failures or unresolved clarifications