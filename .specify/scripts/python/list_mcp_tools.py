#!/usr/bin/env python3
"""
List MCP Tools Script
Reference implementation based on Plan B from original design docs.
Extracts 'mcpServers' configuration from VS Code mcp.json files.
"""

import asyncio
import json
import os
import platform
import re
import sys
import traceback
import uuid
from pathlib import Path
from typing import Any, Dict, List

# Ensure mcp[cli] is installed
try:
    import httpx
    from mcp import ClientSession
    from mcp.client.sse import sse_client
except ImportError:
    print(
        "Error: mcp library not found. Please install it with: pip3 install 'mcp[cli]'",
        file=sys.stderr,
    )
    sys.exit(1)


async def fetch_tools_with_mcp(
    url: str, headers: Dict[str, str], type: str = "sse"
) -> List[Dict[str, Any]]:
    """
    Fetch tools using the MCP SDK's SSE client for SSE, or generic HTTP POST for http.
    """
    if not url.startswith(("http://", "https://")):
        return []

    try:
        if type == "sse":
            # Connect via SSE
            async with sse_client(url, headers=headers) as (read, write):
                async with ClientSession(read, write) as session:
                    await session.initialize()
                    result = await session.list_tools()
                    tools_list = []
                    for tool in result.tools:
                        tools_list.append(
                            {
                                "name": tool.name,
                                "description": tool.description,
                                "inputSchema": tool.inputSchema,
                            }
                        )
                    return tools_list

        elif type == "http":
            # Stateless JSON-RPC over HTTP
            # For servers that don't support SSE (returning 405 on GET), we must implement
            # the MCP handshake manually over POST: initialize -> initialized -> tools/list

            # Ensure session ID is present for tracking
            # But DO NOT send it for initialize if we want the server to create one (hypothesis)
            session_id_header = headers.get("Mcp-Session-Id")
            if not session_id_header:
                session_id_header = str(uuid.uuid4())
                headers["Mcp-Session-Id"] = session_id_header

            async with httpx.AsyncClient(timeout=30.0) as client:
                # 1. Initialize
                # Try WITHOUT Mcp-Session-Id first for initialize.
                init_headers = headers.copy()
                if "Mcp-Session-Id" in init_headers:
                    del init_headers["Mcp-Session-Id"]

                init_payload = {
                    "jsonrpc": "2.0",
                    "method": "initialize",
                    "id": 1,
                    "params": {
                        "protocolVersion": "2024-11-05",
                        "capabilities": {},
                        "clientInfo": {"name": "spec-kit", "version": "1.0"},
                    },
                }
                resp_init = await client.post(
                    url, headers=init_headers, json=init_payload
                )

                # Check 400 or 404 here specially
                if resp_init.status_code == 400 and "Mcp-Session-Id" in resp_init.text:
                    # Fallback: The server requires header even for init.
                    # But we know random uuid fails.
                    # Only option: use the random one and hope it was a fluke?
                    # No, 404 confirmed validation.
                    # We reprint error and exit if this happens.
                    pass

                resp_init.raise_for_status()

                # Capture session ID if returned in headers
                if "Mcp-Session-Id" in resp_init.headers:
                    headers["Mcp-Session-Id"] = resp_init.headers["Mcp-Session-Id"]

                # Check for protocol errors
                init_data = resp_init.json()
                if "error" in init_data:
                    print(
                        f"    [RPC Error during init] {init_data['error']}",
                        file=sys.stderr,
                    )
                    # If "session not found" on init, it's fatal.
                    return []

                # 2. Initialized Notification
                # Standard MCP requires this notification after successful initialize response
                notify_payload = {
                    "jsonrpc": "2.0",
                    "method": "notifications/initialized",
                    "params": {},
                }
                await client.post(url, headers=headers, json=notify_payload)

                # 3. List Tools
                list_payload = {
                    "jsonrpc": "2.0",
                    "method": "tools/list",
                    "id": 2,
                    "params": {},
                }
                resp_tools = await client.post(url, headers=headers, json=list_payload)
                resp_tools.raise_for_status()

                data = resp_tools.json()
                if "error" in data:
                    print(f"    [RPC Error] {data['error']}", file=sys.stderr)
                    return []

                if "result" in data and "tools" in data["result"]:
                    tools = data["result"]["tools"]
                    return tools
                return []

    except Exception as e:
        print(f"    [Error] Failed to fetch tools: {e}", file=sys.stderr)
        # traceback.print_exc() # Reduce noise
        return []

    except Exception as e:
        print(f"    [Error] Failed to fetch tools: {e}", file=sys.stderr)
        traceback.print_exc()
        return []


def expand_path_variables(value: str) -> str:
    """
    Expand VSCode-style variables like ${env:VAR} and ${userHome} in strings.
    """
    if not isinstance(value, str):
        return str(value)

    # 1. Handle ${userHome}
    if "${userHome}" in value:
        value = value.replace("${userHome}", str(Path.home()))

    # 2. Handle ${env:VARIABLE}
    # Pattern to match ${env:VAR_NAME}
    env_pattern = re.compile(r"\$\{env:([a-zA-Z_][a-zA-Z0-9_]*)\}")

    def replace_env(match):
        var_name = match.group(1)
        # Return env var value, or empty string if not set
        val = os.environ.get(var_name)
        if val is None:
            # Warn locally? Or just return empty?
            # For header auth, empty usually breaks it, but explicit warning is better.
            print(
                f"    [Warning] Environment variable '{var_name}' not found.",
                file=sys.stderr,
            )
            return ""
        return val

    return env_pattern.sub(replace_env, value)


def get_vscode_settings_paths() -> List[Path]:
    """
    Get potential paths for VS Code MCP configuration files.
    Priority:
    1. Workspace Config (.vscode/mcp.json)
    2. Remote Config (VS Code Server data)
    3. User Config (OS-specific standard User paths)
    """
    paths = []
    home = Path.home()

    # 1. Workspace Config (Highest Priority)
    paths.append(Path.cwd() / ".vscode" / "mcp.json")

    # 2. Remote Config (VS Code Server)
    # Check both standard and insiders servers
    # These are typically found on the remote machine when using VS Code Remote
    paths.append(home / ".vscode-server" / "data" / "User" / "mcp.json")
    paths.append(home / ".vscode-server-insiders" / "data" / "User" / "mcp.json")

    # 3. User Config (Local machine via standard paths)
    # For a purely remote script, these might not exist or might refer to the remote user's desktop config.
    system = platform.system()
    if system == "Windows":
        appdata = os.environ.get("APPDATA")
        if appdata:
            base = Path(appdata)
            paths.append(base / "Code" / "User" / "mcp.json")
            paths.append(base / "Code - Insiders" / "User" / "mcp.json")
    elif system == "Darwin":  # macOS
        base = home / "Library" / "Application Support"
        paths.append(base / "Code" / "User" / "mcp.json")
        paths.append(base / "Code - Insiders" / "User" / "mcp.json")
    else:  # Linux
        base = home / ".config"
        paths.append(base / "Code" / "User" / "mcp.json")
        paths.append(base / "Code - Insiders" / "User" / "mcp.json")
        paths.append(base / "Code - OSS" / "User" / "mcp.json")

    return paths


def load_mcp_servers() -> List[Dict[str, Any]]:
    """
    Scan all settings paths and aggregate unique MCP servers defined in 'copilot.mcpServers'.
    """
    mcp_servers = []
    seen_names = set()

    search_paths = get_vscode_settings_paths()
    print(
        f"Scanning for VS Code settings in {len(search_paths)} locations...",
        file=sys.stderr,
    )

    for p in search_paths:
        if p.exists():
            try:
                print(f"  Reading: {p}", file=sys.stderr)
                with open(p, "r", encoding="utf-8") as f:
                    # Allow comments in json (standard for vscode settings)
                    # Simple hack: remove lines starting with keys that might be comments?
                    # Actually standard json.load might fail on comments.
                    # For a robust script, we might need a comment-stripping parser.
                    # For now, try standard json load.
                    content = f.read()

                # Basic comment stripping (C-style //)
                # This is fragile but better than nothing for settings.json
                clean_content = "\n".join(
                    [
                        line
                        for line in content.splitlines()
                        if not line.strip().startswith("//")
                    ]
                )

                settings = json.loads(clean_content)

                current_file_servers = []

                # Handle mcp.json format
                # Supports both "mcpServers" (standard) and "servers" (legacy/variant) keys
                mcp_servers_dict = settings.get("mcpServers", {})
                if not mcp_servers_dict:
                    mcp_servers_dict = settings.get("servers", {})

                if isinstance(mcp_servers_dict, dict):
                    for name, config in mcp_servers_dict.items():
                        server_entry = config.copy()
                        server_entry["name"] = name
                        current_file_servers.append(server_entry)

                if not isinstance(current_file_servers, list):
                    print(
                        f"  [Info] No valid server list found in {p}", file=sys.stderr
                    )
                    continue

                for server in current_file_servers:
                    name = server.get("name")
                    if name and name not in seen_names:
                        seen_names.add(name)
                        server["_source"] = str(p)  # metadata
                        mcp_servers.append(server)

            except json.JSONDecodeError as e:
                print(f"  [Warning] Failed to parse JSON in {p}: {e}", file=sys.stderr)
            except Exception as e:
                print(f"  [Warning] Error reading {p}: {e}", file=sys.stderr)

    return mcp_servers


async def main():
    # Check for mandatory MCP_AUTH
    mcp_auth = os.environ.get("MCP_AUTH")
    if not mcp_auth:
        print("[Error] MCP_AUTH environment variable is not set.", file=sys.stderr)
        sys.exit(1)

    servers = load_mcp_servers()

    # Enrich servers with tools
    print("\nFetching tools from available servers...", file=sys.stderr)
    for server in servers:
        # Check if server matches criteria for extraction (http/sse)
        srv_type = server.get("type", "unknown")
        srv_url = server.get("url")

        if srv_url and srv_type in (
            "http",
            "sse",
        ):  # Assuming 'http' type uses JSON-RPC over HTTP
            print(
                f"  Probing {server.get('name', 'unnamed')} ({srv_url})...",
                file=sys.stderr,
            )
            # Some entries might have 'headers' config
            # Extract extra headers if present in config (e.g. auth)
            headers = {}
            config_headers = server.get("headers")
            if isinstance(config_headers, dict):
                for k, v in config_headers.items():
                    # Expand variables in values (e.g. ${env:MCP_TOKEN})
                    headers[k] = expand_path_variables(str(v))

            # Enforce Authorization header from env var
            headers["Authorization"] = f"Bearer {mcp_auth}"
            # Some servers require Mcp-Session-Id even for stateless calls
            if "Mcp-Session-Id" not in headers:
                headers["Mcp-Session-Id"] = str(uuid.uuid4())

            tools = await fetch_tools_with_mcp(srv_url, headers, type=srv_type)
            if tools:
                server["tools"] = tools
                server["tools_count"] = len(tools)
                print(f"    -> Found {len(tools)} tools", file=sys.stderr)
            else:
                print("    -> No tools retrieved", file=sys.stderr)

        elif srv_type == "stdio":
            print(
                f"  Skipping stdio server {server.get('name', 'unnamed')}: Cannot list tools via script",
                file=sys.stderr,
            )

    output = {
        "timestamp": __import__("datetime").datetime.now().isoformat(),
        "count": len(servers),
        "servers": servers,
        "note": "This list represents configured MCP servers. 'tools' field populated for HTTP servers if reachable.",
    }

    # Print formatted JSON to stdout
    print(json.dumps(output, indent=2, ensure_ascii=False))

    if not servers:
        print(
            "\nNo MCP servers found in standard VS Code configuration paths.",
            file=sys.stderr,
        )
        print(
            "Ensure you have configured 'copilot.mcpServers' in your settings.json.",
            file=sys.stderr,
        )


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        pass
