#!/usr/bin/env bash

# Load common helpers for Unicode support and shared functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/common.sh" ]; then
  # shellcheck source=/dev/null
  source "$SCRIPT_DIR/common.sh"
  ensure_utf8_locale || true
else
  echo "Failed to load common.sh, spec-kit framework not installed correctly"
  exit 1
fi

set -e

# Set root dir
if git rev-parse --show-toplevel >/dev/null 2>&1; then
  ROOT_DIR=$(git rev-parse --show-toplevel)
else
  case "$SCRIPT_DIR" in
    */.specify/scripts/bash)
      ROOT_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
      ;;
    *)
      ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
      ;;
  esac
fi

SPECIFY_PY_DIR="$ROOT_DIR/.specify/scripts/python"

if [ -f "$SPECIFY_PY_DIR/tools-utils.py" ]; then
  PY_SCRIPTS_DIR="$SPECIFY_PY_DIR"
else
  PY_SCRIPTS_DIR="$ROOT_DIR/scripts/python"
fi

TOOLS_UTILS_SCRIPT="$PY_SCRIPTS_DIR/tools-utils.py"

for required in "$TOOLS_UTILS_SCRIPT"; do
  if [ ! -f "$required" ]; then
    echo "Required script not found: $required"
    exit 1
  fi
done

QUERY_MCP=false
QUERY_SYSTEM=false
QUERY_SHELL=false
QUERY_PROJECT=false

usage() {
  echo "Usage: $0 [--mcp] [--system] [--shell] [--project]"
  exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --mcp)
      QUERY_MCP=true
      shift
      ;;
    --system)
      QUERY_SYSTEM=true
      shift
      ;;
    --shell)
      QUERY_SHELL=true
      shift
      ;;
    --project)
      QUERY_PROJECT=true
      shift
      ;;
    *)
      usage
      ;;
  esac
done

if [ "$QUERY_MCP" = false ] && [ "$QUERY_SYSTEM" = false ] && [ "$QUERY_SHELL" = false ] && [ "$QUERY_PROJECT" = false ]; then
  usage
fi

get_mcp_tools_json() {
  python3 "$TOOLS_UTILS_SCRIPT" --action list --type mcp 2>/dev/null || echo "[]"
}

get_system_binaries_json() {
  python3 "$TOOLS_UTILS_SCRIPT" --action list --type system 2>/dev/null || echo '{"binaries":[]}'
}

get_shell_function_json() {
  python3 "$TOOLS_UTILS_SCRIPT" --action list --type shell --functions-only 2>/dev/null || echo "[]"
}

get_project_scripts_json() {
  python3 "$TOOLS_UTILS_SCRIPT" --action list --type project --root-dir "$ROOT_DIR" 2>/dev/null || echo "[]"
}

if [ "$QUERY_MCP" = true ]; then
  get_mcp_tools_json
fi
if [ "$QUERY_SYSTEM" = true ]; then
  get_system_binaries_json
fi
if [ "$QUERY_SHELL" = true ]; then
  get_shell_function_json
fi
if [ "$QUERY_PROJECT" = true ]; then
  get_project_scripts_json
fi
