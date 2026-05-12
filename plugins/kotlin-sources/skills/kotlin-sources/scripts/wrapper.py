#!/usr/bin/env python3
"""
Kotlin Sources Skill Wrapper

Provides utilities for exploring the Kotlin compiler source code.
Works with local clones or fetches from GitHub.
"""

import argparse
import subprocess
import sys
import os
from pathlib import Path
from typing import Optional

GITHUB_REPO = "JetBrains/kotlin"
GITHUB_URL = f"https://github.com/{GITHUB_REPO}"
DEFAULT_BRANCH = "master"
DEFAULT_REPO_DIR = Path(
    os.environ.get("KOTLIN_SOURCES_DIR", Path.home() / ".kotlin-sources" / "kotlin")
)

# Common search locations for different query types
SEARCH_HINTS = {
    "diagnostic": ["compiler/fir/checkers", "compiler/frontend"],
    "ir": ["compiler/ir"],
    "fir": ["compiler/fir"],
    "backend": ["compiler/backend.jvm", "compiler/backend.js", "compiler/backend.wasm"],
    "stdlib": ["libraries/stdlib"],
    "plugin": ["plugins"],
    "native": ["kotlin-native"],
    "analysis": ["analysis"],
}


def find_kotlin_repo():
    """Find local Kotlin repository."""
    # Check current directory
    cwd = Path.cwd()
    if (cwd / "compiler" / "fir").exists() and (cwd / "libraries" / "stdlib").exists():
        return cwd

    # Check common locations
    common_paths = [
        DEFAULT_REPO_DIR,
        Path.home() / "projects" / "kotlin",
        Path.home() / "dev" / "kotlin",
        Path.home() / "kotlin",
    ]

    for p in common_paths:
        if p.exists() and (p / "compiler").exists():
            return p

    return None


def run_command(cmd: list[str], cwd: Optional[Path] = None) -> None:
    subprocess.run(cmd, cwd=cwd, check=True)


def ensure_kotlin_repo(ref: str = DEFAULT_BRANCH, update: bool = False) -> Path:
    """Clone or update the Kotlin repository."""
    repo_path = DEFAULT_REPO_DIR
    repo_path.parent.mkdir(parents=True, exist_ok=True)

    if (repo_path / ".git").exists():
        if update:
            print(f"Updating Kotlin repository at {repo_path}...")
            run_command(["git", "fetch", "--tags", "--prune", "origin"], cwd=repo_path)
        else:
            print(f"Kotlin repository found at {repo_path}")
    else:
        print(f"Cloning Kotlin repository to {repo_path}...")
        run_command(["git", "clone", "--depth", "100", f"{GITHUB_URL}.git", str(repo_path)])

    if ref != DEFAULT_BRANCH:
        run_command(["git", "fetch", "--depth", "1", "origin", ref], cwd=repo_path)
        run_command(["git", "checkout", ref], cwd=repo_path)
    elif update:
        run_command(["git", "checkout", DEFAULT_BRANCH], cwd=repo_path)
        run_command(["git", "pull", "--ff-only", "origin", DEFAULT_BRANCH], cwd=repo_path)

    current = subprocess.run(
        ["git", "rev-parse", "--short", "HEAD"],
        cwd=repo_path,
        capture_output=True,
        text=True,
        check=True,
    ).stdout.strip()
    print(f"Kotlin sources ready at: {repo_path}")
    print(f"Current commit: {current}")
    return repo_path


def run_gh_command(args: list[str]) -> str:
    """Run a GitHub CLI command."""
    try:
        result = subprocess.run(
            ["gh"] + args,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error running gh command: {e.stderr}", file=sys.stderr)
        sys.exit(1)


def search_local(repo_path: Path, pattern: str, path: str = None, file_type: str = None):
    """Search in local repository using ripgrep."""
    cmd = ["rg", "--line-number", "--color=never"]

    if file_type:
        cmd.extend(["--type", file_type])

    cmd.append(pattern)

    search_path = repo_path / path if path else repo_path
    cmd.append(str(search_path))

    try:
        result = subprocess.run(cmd, capture_output=True, text=True)
        return result.stdout
    except FileNotFoundError:
        # Fallback to grep if rg not available
        cmd = ["grep", "-rn", pattern, str(search_path)]
        result = subprocess.run(cmd, capture_output=True, text=True)
        return result.stdout


def search_github(pattern: str, path: str = None, file_type: str = None):
    """Search on GitHub using gh CLI."""
    query = f'"{pattern}" repo:{GITHUB_REPO}'

    if path:
        query += f" path:{path}"

    if file_type:
        query += f" language:{file_type}"

    return run_gh_command(["search", "code", query, "--limit", "30"])


def browse_local(repo_path: Path, path: str):
    """List directory contents in local repo."""
    target = repo_path / path
    if not target.exists():
        print(f"Path not found: {path}", file=sys.stderr)
        sys.exit(1)

    if target.is_file():
        return target.read_text()

    items = sorted(target.iterdir())
    result = []
    for item in items:
        prefix = "d " if item.is_dir() else "f "
        result.append(f"{prefix}{item.name}")
    return "\n".join(result)


def browse_github(path: str, ref: str = DEFAULT_BRANCH):
    """Browse GitHub repository contents."""
    try:
        return run_gh_command([
            "api", f"repos/{GITHUB_REPO}/contents/{path}",
            "-q", '.[] | "\(.type[0]) \(.name)"'
        ])
    except:
        # Try as file
        return run_gh_command([
            "api", f"repos/{GITHUB_REPO}/contents/{path}",
            "-q", '.content', "--jq", "@base64d"
        ])


def read_file(repo_path: Path, file_path: str):
    """Read a file from local repo or GitHub."""
    if repo_path:
        target = repo_path / file_path
        if target.exists():
            return target.read_text()

    # Fallback to GitHub
    content = run_gh_command([
        "api", f"repos/{GITHUB_REPO}/contents/{file_path}",
        "--jq", ".content"
    ])
    import base64
    return base64.b64decode(content).decode("utf-8")


def find_diagnostics(repo_path: Path, diagnostic_name: str):
    """Find where a diagnostic/error is defined and used."""
    results = []

    # Search in FIR checkers (K2)
    fir_result = search_local(
        repo_path, diagnostic_name,
        path="compiler/fir/checkers"
    ) if repo_path else search_github(diagnostic_name, "compiler/fir/checkers")

    if fir_result:
        results.append("=== FIR Checkers (K2) ===")
        results.append(fir_result)

    # Search in frontend diagnostics (K1)
    k1_result = search_local(
        repo_path, diagnostic_name,
        path="compiler/frontend"
    ) if repo_path else search_github(diagnostic_name, "compiler/frontend")

    if k1_result:
        results.append("\n=== Frontend (K1) ===")
        results.append(k1_result)

    return "\n".join(results)


def compare_versions(repo_path: Path, v1: str, v2: str, path: str = None):
    """Compare two versions/tags."""
    if not repo_path:
        print("Version comparison requires local repository", file=sys.stderr)
        sys.exit(1)

    cmd = ["git", "-C", str(repo_path), "diff", v1, v2]
    if path:
        cmd.extend(["--", path])

    result = subprocess.run(cmd, capture_output=True, text=True)
    return result.stdout


def main():
    parser = argparse.ArgumentParser(description="Kotlin Sources Explorer")
    subparsers = parser.add_subparsers(dest="command", required=True)

    # Init command
    init_parser = subparsers.add_parser("init", help="Clone Kotlin sources locally")
    init_parser.add_argument("--ref", default=DEFAULT_BRANCH, help="Branch, tag, or commit")

    # Update command
    update_parser = subparsers.add_parser("update", help="Update local Kotlin sources")
    update_parser.add_argument("--ref", default=DEFAULT_BRANCH, help="Branch, tag, or commit")

    # Search command
    search_parser = subparsers.add_parser("search", help="Search source code")
    search_parser.add_argument("pattern", help="Search pattern")
    search_parser.add_argument("--path", "-p", help="Limit search to path")
    search_parser.add_argument("--type", "-t", help="File type (kt, java, etc.)")

    # Browse command
    browse_parser = subparsers.add_parser("browse", help="Browse directory")
    browse_parser.add_argument("path", help="Path to browse")

    # Read command
    read_parser = subparsers.add_parser("read", help="Read file contents")
    read_parser.add_argument("file", help="File path to read")

    # Diagnostics command
    diag_parser = subparsers.add_parser("diagnostics", help="Find diagnostic definition")
    diag_parser.add_argument("name", help="Diagnostic name")

    # Compare command
    compare_parser = subparsers.add_parser("compare", help="Compare versions")
    compare_parser.add_argument("v1", help="First version/tag")
    compare_parser.add_argument("v2", help="Second version/tag")
    compare_parser.add_argument("--path", "-p", help="Limit to path")

    args = parser.parse_args()

    if args.command == "init":
        ensure_kotlin_repo(args.ref, update=False)
        return

    if args.command == "update":
        ensure_kotlin_repo(args.ref, update=True)
        return

    repo_path = find_kotlin_repo()

    if args.command == "search":
        if repo_path:
            result = search_local(repo_path, args.pattern, args.path, args.type)
        else:
            result = search_github(args.pattern, args.path, args.type)
        print(result)

    elif args.command == "browse":
        if repo_path:
            result = browse_local(repo_path, args.path)
        else:
            result = browse_github(args.path)
        print(result)

    elif args.command == "read":
        result = read_file(repo_path, args.file)
        print(result)

    elif args.command == "diagnostics":
        result = find_diagnostics(repo_path, args.name)
        print(result)

    elif args.command == "compare":
        result = compare_versions(repo_path, args.v1, args.v2, args.path)
        print(result)


if __name__ == "__main__":
    main()
