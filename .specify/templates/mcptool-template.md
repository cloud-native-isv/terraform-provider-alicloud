# MCP Tool Record: [MCP TOOL NAME]

**Tool Name**: [MCP TOOL NAME]  
**MCP Server**: [MCP SERVER NAME]  
**Status**: [Draft | Verified | Deprecated]  
**Last Updated**: [YYYY-MM-DD]

## Description

[Short, user-friendly description of what this MCP tool does and when to use it]

## Parameters

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| [param] | [type] | [yes/no] | [purpose] | [default or empty] |

## Returns

| Name | Type | Description |
|------|------|-------------|
| [field] | [type] | [meaning of return field] |

## Usage Notes

- [Any constraints, preconditions, or special handling]
- [Rate limits, permission requirements, or error behavior]

## Examples

### Example Input

```json
{
  "tool": "[MCP TOOL NAME]",
  "arguments": {
    "[param]": "value"
  }
}
```

### Example Output

```json
{
  "result": {
    "[field]": "value"
  }
}
```

## Discovery Metadata

- **Discovery Method**: [auto-discovery | manual-entry | imported]
- **Discovery Source**: [server config path, tool registry, or reference]
- **Verification Status**: [unverified | verified]
- **Notes**: [Any additional context]

使用[MCP TOOL PARAMETERS]调用[MCP TOOL NAME]工具
