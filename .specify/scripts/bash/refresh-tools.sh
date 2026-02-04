#!/usr/bin/env bash

set -e

# Load common helpers for Unicode support and shared functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/common.sh" ]; then
    # shellcheck source=/dev/null
    source "$SCRIPT_DIR/common.sh"
    # Ensure UTF-8 locale for better Unicode handling
    ensure_utf8_locale || true
else 
    echo 
fi

# Set root dir
if git rev-parse --show-toplevel >/dev/null 2>&1; then
    ROOT_DIR=$(git rev-parse --show-toplevel)
else
    # Fallback: check if we are in .specify/scripts/bash (depth 3) or scripts/bash (depth 2)
    case "$SCRIPT_DIR" in
        */.specify/scripts/bash)
            ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
            ;;
        *)
            ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
            ;;
    esac
fi

SKILLS_DIR="$ROOT_DIR/.github/skills"
JSON_MODE=false

# Query mode flags
QUERY_MCP=false
QUERY_SYSTEM=false
QUERY_SHELL=false
QUERY_PROJECT=false
OUTPUT_FORMAT="json"
IS_QUERY_MODE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --mcp)
      QUERY_MCP=true
      IS_QUERY_MODE=true
      shift
      ;;
    --system)
      QUERY_SYSTEM=true
      IS_QUERY_MODE=true
      shift
      ;;
    --shell)
      QUERY_SHELL=true
      IS_QUERY_MODE=true
      shift
      ;;
    --project)
      QUERY_PROJECT=true
      IS_QUERY_MODE=true
      shift
      ;;
    --format)
      OUTPUT_FORMAT="$2"
      shift 2
      ;;
    --json)
      JSON_MODE=true
      OUTPUT_FORMAT="json"
      shift
      ;;
    *)
      shift
      ;;
  esac
done

# Shared paths
TOOLS_DOC="$ROOT_DIR/.specify/memory/tools.md"

json_escape() {
    local value="$1"
    value="${value//\\/\\\\}"
    value="${value//\"/\\\"}"
    value="${value//$'\n'/\\n}"
    value="${value//$'\r'/\\r}"
    value="${value//$'\t'/\\t}"
    printf '%s' "$value"
}

get_mcp_tools_json() {
    if [ -f "$TOOLS_DOC" ]; then
        local json
        json=$(awk '
            BEGIN { in_section=0; in_block=0 }
            /^## MCP Tools/ { in_section=1 }
            in_section && /^```json/ { in_block=1; next }
            in_block && /^```/ { exit }
            in_block { print }
        ' "$TOOLS_DOC")
        if [ -n "$json" ]; then
            printf '%s' "$json"
            return
        fi
    fi
    printf '[]'
}

get_system_binaries_json() {
    local os_release=""
    if [ -f /etc/os-release ]; then
        os_release=$(cat /etc/os-release)
    fi
    local kernel
    kernel=$(uname -a)

    local binaries=(git docker kubectl python3 python pip node npm hatch gh jq curl wget make)
    
    printf '{"os_release":"%s","kernel":"%s","binaries":[' "$(json_escape "$os_release")" "$(json_escape "$kernel")"

    local first=true
    for b in "${binaries[@]}"; do
        local path
        path=$(command -v "$b" 2>/dev/null || true)
        if [ -n "$path" ]; then
            if [ "$first" = true ]; then
                first=false
            else
                printf ','
            fi
            printf '{"name":"%s","path":"%s"}' "$(json_escape "$b")" "$(json_escape "$path")"
        fi
    done
    printf ']}'
}

get_shell_function_json() {
    local first=true
    printf '['
    # Use compgen to list all functions
    # define mapping to skip internal functions of this script
    local skip_funcs="json_escape|get_mcp_tools_json|get_system_binaries_json|get_shell_function_json|get_project_scripts_json|print_mcp_tools_markdown|print_system_binaries_markdown|print_shell_function_markdown|print_project_scripts_markdown|ensure_utf8_locale|report_error|report_success|validate_skill_name|create_skill_structure"
    
    while IFS= read -r func; do
        if [[ "$func" =~ ^($skip_funcs)$ ]]; then
            continue
        fi
        
        # Get function definition (first line only, effectively the signature)
        local def
        def=$(declare -f "$func" | head -n 10)
        
        if [ "$first" = true ]; then
            first=false
        else
            printf ','
        fi
        printf '{"name":"%s","definition":"%s"}' "$(json_escape "$func")" "$(json_escape "$def")"
    done < <(compgen -A function | sort)
    printf ']'
}

get_project_scripts_json() {
    local scripts_dir=""
    if [ -d "$ROOT_DIR/.specify/scripts" ]; then
        scripts_dir="$ROOT_DIR/.specify/scripts"
    elif [ -d "$ROOT_DIR/scripts" ]; then
        scripts_dir="$ROOT_DIR/scripts"
    else
        printf '[]'
        return
    fi

    local first=true
    printf '['
    while IFS= read -r file; do
        local rel_path="${file#$ROOT_DIR/}"
        local name
        name=$(basename "$file")
        local type="bash"
        local description=""

        if [[ "$file" == *.py ]]; then
            type="python"
            if command -v python3 >/dev/null 2>&1; then
                description=$(python3 -c "import ast; print((ast.get_docstring(ast.parse(open('$file').read())) or '').split('\n')[0])" 2>/dev/null || true)
            fi
        else
            # Get first non-shebang comment
            description=$(grep '^#' "$file" | grep -v '^#!' | head -n 1 | sed 's/^#[[:space:]]*//')
        fi
        
        # Default description if empty
        if [ -z "$description" ]; then
            description="No description available"
        fi

        if [ "$first" = true ]; then
            first=false
        else
            printf ','
        fi
        printf '{"name":"%s","path":"%s","type":"%s","description":"%s"}' \
            "$(json_escape "$name")" "$(json_escape "$rel_path")" "$(json_escape "$type")" "$(json_escape "$description")"
    done < <(find "$scripts_dir" -type f \( -name "*.sh" -o -name "*.py" \) | sort)
    printf ']'
}

print_mcp_tools_markdown() {
    local json
    json=$(get_mcp_tools_json)
    echo "## MCP Tools"
    if command -v jq >/dev/null 2>&1; then
        echo "$json" | jq -r '
            def format_args:
                if (.inputSchema and .inputSchema.properties and (.inputSchema.properties | length > 0)) then
                    "\n  - Arguments:\n" +
                    (.inputSchema.properties | to_entries | map("    - `" + .key + "` (" + (.value.type // "any") + ")" + (if .value.description then ": " + .value.description else "" end)) | join("\n"))
                else
                    ""
                end;

            if type=="object" and .servers then
                .servers[] | (
                    "### " + (.name // "Unknown Server") + "\n",
                    (.tools[]? | 
                        "- **" + (.name // "Unknown") + "**: " + (.description // "No description") + format_args
                    )
                )
            else
                .[] | 
                "- **" + (.name // "Unknown") + "**: " + (.description // "No description") + format_args
            end'
    else
        echo '```json'
        echo "$json"
        echo '```'
    fi
    echo ""
}

print_system_binaries_markdown() {
    local json
    json=$(get_system_binaries_json)
    
    echo "## System Information"
    if command -v jq >/dev/null 2>&1; then
        echo "$json" | jq -r '
            if .os_release != "" then
                "### OS Release\n```\n" + .os_release + "\n```\n"
            else "" end,
            "### Kernel\n`" + .kernel + "`\n",
            "### System Binaries\n| Binary | Path |\n|---|---|",
            (.binaries[] | "| " + .name + " | " + .path + " |")
        '
    else
        # Fallback to json dump
        echo '```json'
        echo "$json"
        echo '```'
    fi
    echo ""
}

print_shell_function_markdown() {
    local json
    json=$(get_shell_function_json)
    echo "## Shell Functions"
    if command -v jq >/dev/null 2>&1; then
        echo "$json" | jq -r '
            if length > 0 then
                .[] | ("### " + .name + "\n```bash\n" + .definition + "\n```\n")
            else
                "No custom shell functions found."
            end'
    else
        echo '```json'
        echo "$json"
        echo '```'
    fi
    echo ""
}

print_project_scripts_markdown() {
    local json
    json=$(get_project_scripts_json)
    if [ "$json" = "[]" ]; then
        return
    fi
    
    echo "## Project Scripts"
    if command -v jq >/dev/null 2>&1; then
        echo "$json" | jq -r '
            if length > 0 then
                "| Script | Description | Path | Type |\n|---|---|---|---|",
                (.[] | "| " + .name + " | " + .description + " | " + .path + " | " + .type + " |")
            else
                "No project scripts found."
            end'
    else
        echo '```json'
        echo "$json"
        echo '```'
    fi
    echo ""
}

# If in query mode, handle directly in shell and exit
if [ "$IS_QUERY_MODE" = true ]; then
    if [ "$OUTPUT_FORMAT" = "markdown" ]; then
        if [ "$QUERY_MCP" = true ]; then
            print_mcp_tools_markdown
        fi
        if [ "$QUERY_SYSTEM" = true ]; then
            print_system_binaries_markdown
        fi
        if [ "$QUERY_SHELL" = true ]; then
            print_shell_function_markdown
        fi
        if [ "$QUERY_PROJECT" = true ]; then
            print_project_scripts_markdown
        fi
    else
        first=true
        printf '{'
        if [ "$QUERY_MCP" = true ]; then
            if [ "$first" = true ]; then
                first=false
            else
                printf ','
            fi
            printf '"mcp_tools":%s' "$(get_mcp_tools_json)"
        fi
        if [ "$QUERY_SYSTEM" = true ]; then
            if [ "$first" = true ]; then
                first=false
            else
                printf ','
            fi
            printf '"system_binaries":%s' "$(get_system_binaries_json)"
        fi
        if [ "$QUERY_SHELL" = true ]; then
            if [ "$first" = true ]; then
                first=false
            else
                printf ','
            fi
            printf '"shell_functions":%s' "$(get_shell_function_json)"
        fi
        if [ "$QUERY_PROJECT" = true ]; then
            if [ "$first" = true ]; then
                first=false
            else
                printf ','
            fi
            printf '"project_scripts":%s' "$(get_project_scripts_json)"
        fi
        printf '}'
    fi
    exit 0
fi

# Generate tools list
MCP_SCRIPT="$ROOT_DIR/.specify/scripts/python/list_mcp_tools.py"
# Fallback to local script if not found in .specify
if [ ! -f "$MCP_SCRIPT" ]; then
    MCP_SCRIPT="$ROOT_DIR/scripts/python/list_mcp_tools.py"
fi

# Try to get MCP tools list, default to empty list on failure
if [ -f "$MCP_SCRIPT" ]; then
    # Use the python from the virtualenv if active, or just python3
    MCP_LIST=$(python3 "$MCP_SCRIPT" 2>/dev/null || echo "[]")
else

    MCP_LIST="[]"
fi

if [ -z "$MCP_LIST" ]; then MCP_LIST="[]"; fi

{
    echo "# Helper Tools Index"
    echo "This document indexes available tools for the agent."
    echo ""
    echo "## MCP Tools"
    echo '```json'
    # Format with jq if available, otherwise raw
    if command -v jq >/dev/null 2>&1; then
        echo "$MCP_LIST" | jq '.'
    else
        echo "$MCP_LIST"
    fi
    echo '```'
    echo ""
    echo "## System Binaries"
    echo "Standard executables in PATH (checked via 'command -v' or 'which')."
    echo ""
    echo "## Shell Functions"
    echo "Custom shell functions defined in the active session."
    echo ""
    echo "## Project Scripts"
    echo "Automation scripts located in the project 'scripts/' directory."
}
