#!/usr/bin/env python3
"""Read-only documentation gate. Policy lives in documentation-standard.md."""
import argparse
import hashlib
import json
import re
import sys
from pathlib import Path, PurePosixPath
from urllib.parse import unquote, urlsplit

POLICY_START = "<!-- docs-check-policy:start -->"
POLICY_END = "<!-- docs-check-policy:end -->"
ATTACHMENTS = {".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".pdf",
               ".json", ".yaml", ".yml", ".csv", ".tsv", ".txt"}
GENERIC_NAMES = {"tmp", "temp", "draft", "latest", "backup", "copy", "new",
                 "untitled", "临时", "草稿", "最新", "备份", "副本", "新建文档", "未命名"}
LINK = re.compile(r'!?\[[^\]\n]*\]\(\s*(?:<([^>\n]+)>|([^\s)]+))(?:\s+["\'][^\n]*["\'])?\s*\)')
DEFINITION = re.compile(r'^\s{0,3}\[([^\]]+)\]:\s*(?:<([^>]+)>|(\S+))')
REFERENCE = re.compile(r'!?\[([^\]\n]+)\]\[([^\]\n]*)\]')


def issue(errors, code, path, message, line=None):
    location = str(path) + (":" + str(line) if line else "")
    errors.append("{} {}: {}".format(code, location, message))


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate policy key")
        result[key] = value
    return result


def load_policy(docs):
    path = docs / "documentation-standard.md"
    if path.is_symlink() or not path.is_file():
        raise ValueError("policy document is missing or is a symbolic link")
    text = path.read_text(encoding="utf-8")
    if text.count(POLICY_START) != 1 or text.count(POLICY_END) != 1:
        raise ValueError("policy markers must occur exactly once")
    section = text.split(POLICY_START, 1)[1].split(POLICY_END, 1)[0]
    match = re.fullmatch(r'\s*```json\s*\n(.*?)\n```\s*', section, re.S)
    if not match:
        raise ValueError("policy must be one JSON code block")
    policy = json.loads(match.group(1), object_pairs_hook=unique_object)
    keys = {"version", "allowed_directories", "allowed_subdirectories",
            "allowed_root_files", "allowed_tool_files", "required_files"}
    if not isinstance(policy, dict) or set(policy) != keys:
        raise ValueError("policy fields do not match the supported schema")
    if type(policy["version"]) is not int or policy["version"] != 2:
        raise ValueError("unsupported policy version")
    for key in keys - {"version"}:
        values = policy[key]
        if (not isinstance(values, list) or not all(isinstance(v, str) and v for v in values)
                or len(values) != len(set(values))):
            raise ValueError("policy lists must contain unique relative paths")
        for value in values:
            path = PurePosixPath(value)
            if (path.is_absolute() or str(path) != value or "\\" in value
                    or any(p in {".", ".."} for p in path.parts)):
                raise ValueError("policy paths must be normalized and relative")
    directories = set(policy["allowed_directories"])
    if not directories or any(not re.fullmatch(r'[a-z][a-z0-9-]*', v) for v in directories):
        raise ValueError("top-level directory names are invalid")
    all_dirs = directories | set(policy["allowed_subdirectories"])
    for name in policy["allowed_subdirectories"]:
        path = PurePosixPath(name)
        if (len(path.parts) != 2 or path.parts[0] not in directories
                or not re.fullmatch(r'[a-z][a-z0-9-]*', path.name)):
            raise ValueError("subdirectories must be named topics immediately below an allowed category")
    root_files = set(policy["allowed_root_files"])
    required = set(policy["required_files"])
    if not {"README.md", "documentation-standard.md"} <= root_files & required:
        raise ValueError("root entry and policy documents must be allowed and required")
    if any("/" in v or not v.endswith(".md") for v in root_files):
        raise ValueError("root files must be Markdown documents")
    tools = set(policy["allowed_tool_files"])
    for name in tools:
        if (str(PurePosixPath(name).parent) not in all_dirs
                or PurePosixPath(name).parts[0] != "development" or not name.endswith(".py")):
            raise ValueError("tools must be registered Python files in development")
    for name in required:
        path = PurePosixPath(name)
        if ((str(path.parent) == "." and name not in root_files)
                or (str(path.parent) != "." and str(path.parent) not in all_dirs)
                or (not name.endswith(".md") and name not in tools)):
            raise ValueError("required files must be inside allowed document locations or registered tools")
    return policy


def valid_name(name):
    if name == "README.md":
        return True
    stem = name[:-3]
    if stem in GENERIC_NAMES or re.search(r'(?:-copy(?:-?\d+)?|副本\d*)$', stem):
        return False
    return bool(re.fullmatch(r'[a-z0-9]+(?:\.[0-9]+)*(?:-[a-z0-9]+(?:\.[0-9]+)*)*', stem)
                or re.search(r'[\u3400-\u9fff]', stem)
                and re.fullmatch(r'[A-Za-z0-9\u3400-\u9fff-]+', stem))


def parse_page(path, root, errors):
    name = path.relative_to(root)
    text = path.read_text(encoding="utf-8")
    lines = []
    fence = None
    for number, line in enumerate(text.splitlines(), 1):
        marker = re.match(r'^\s{0,3}(`{3,}|~{3,})(.*)$', line)
        if marker:
            run, rest = marker.groups()
            if fence is None:
                fence = (run[0], len(run))
            elif run[0] == fence[0] and len(run) >= fence[1] and not rest.strip():
                fence = None
            continue
        if fence is None:
            lines.append((number, line))
    if fence is not None:
        issue(errors, "E_FENCE", name, "unclosed Markdown code block")
    headings = set()
    definitions = {}
    links = []
    for number, line in lines:
        heading = re.match(r'^\s{0,3}#{1,6}\s+(.+?)(?:\s+#+)?\s*$', line)
        if heading:
            title = re.sub(r'[`*_]', '', heading.group(1)).strip().lower()
            slug = re.sub(r'[^\w\- ]', '', title).replace(' ', '-')
            candidate = slug
            count = 0
            while candidate in headings:
                count += 1
                candidate = slug + "-" + str(count)
            headings.add(candidate)
        plain = re.sub(r'(`+).*?\1', '', line)
        definition = DEFINITION.match(plain)
        if definition:
            label = " ".join(definition.group(1).lower().split())
            definitions[label] = definition.group(2) or definition.group(3)
        else:
            for match in LINK.finditer(plain):
                links.append((number, match.group(1) or match.group(2)))
    for number, line in lines:
        plain = re.sub(r'(`+).*?\1', '', line)
        if DEFINITION.match(plain):
            continue
        for match in REFERENCE.finditer(plain):
            label = " ".join((match.group(2) or match.group(1)).lower().split())
            if label not in definitions:
                issue(errors, "E_REFERENCE", name, "undefined Markdown link reference", number)
            else:
                links.append((number, definitions[label]))
        remaining = REFERENCE.sub('', LINK.sub('', plain))
        for match in re.finditer(r'!?\[([^\]\n]+)\]', remaining):
            label = " ".join(match.group(1).lower().split())
            if label in definitions:
                links.append((number, definitions[label]))
    return {"text": text, "anchors": headings, "links": links, "targets": set()}


def validate(root):
    errors = []
    docs = root / "docs"
    if docs.is_symlink() or not docs.is_dir():
        return ["E_POLICY docs: missing directory or symbolic link"], 0, 0
    try:
        policy = load_policy(docs)
    except (ValueError, OSError, UnicodeError):
        return ["E_POLICY docs/documentation-standard.md: invalid or unreadable policy"], 0, 0
    allowed_dirs = set(policy["allowed_directories"]) | set(policy["allowed_subdirectories"])
    root_files = set(policy["allowed_root_files"])
    tools = set(policy["allowed_tool_files"])
    files = set()
    found_dirs = set()
    pending = [docs]
    while pending:
        directory = pending.pop()
        for path in sorted(directory.iterdir()):
            relative = path.relative_to(docs).as_posix()
            display = path.relative_to(root)
            if path.is_symlink():
                issue(errors, "E_SYMLINK", display, "symbolic links are not permitted in docs")
            elif path.is_dir():
                if relative not in allowed_dirs:
                    issue(errors, "E_DIRECTORY", display, "directory is not in the documentation policy")
                else:
                    found_dirs.add(relative)
                    pending.append(path)
            elif path.is_file():
                if path.name == ".DS_Store":
                    continue
                files.add(path)
                if path.parent == docs and relative not in root_files:
                    issue(errors, "E_ROOT_FILE", display, "extra file in documentation root")
                if path.suffix == ".md":
                    if not valid_name(path.name):
                        issue(errors, "E_NAME", display, "invalid or temporary document name")
                elif relative not in tools and path.suffix not in ATTACHMENTS:
                    issue(errors, "E_ARTIFACT", display, "unsupported file or unregistered tool")
            else:
                issue(errors, "E_ARTIFACT", display, "unsupported filesystem entry")
    required_indexes = {docs / "README.md"} | {docs / name / "README.md" for name in found_dirs}
    exact_names = {p.relative_to(docs).as_posix() for p in files}
    for path in sorted(required_indexes | {docs / name for name in policy["required_files"]}):
        if path.relative_to(docs).as_posix() not in exact_names:
            issue(errors, "E_REQUIRED", path.relative_to(root), "required file is missing or has wrong case")
    pages = {}
    for path in sorted(p for p in files if p.suffix == ".md"):
        try:
            pages[path] = parse_page(path, root, errors)
        except (OSError, UnicodeError):
            issue(errors, "E_READ", path.relative_to(root), "cannot read UTF-8 Markdown")
    agent = root / "AGENTS.md"
    if agent.is_file() and not agent.is_symlink():
        pages[agent] = parse_page(agent, root, errors)
    link_count = 0
    for path, page in pages.items():
        for number, target in page["links"]:
            link_count += 1
            try:
                url = urlsplit(target)
                if url.scheme or url.netloc:
                    continue
                local = unquote(url.path)
                local = re.sub(r':\d+(?::\d+)?$', '', local)
                destination = (path.parent / local).resolve() if local else path
                destination.relative_to(root)
            except (ValueError, OSError):
                issue(errors, "E_LINK", path.relative_to(root), "invalid or outside-repository local link", number)
                continue
            if not destination.exists():
                issue(errors, "E_LINK", path.relative_to(root), "local link target does not exist", number)
                continue
            page["targets"].add(destination)
            if url.fragment and destination.suffix == ".md" and destination in pages:
                if unquote(url.fragment) not in pages[destination]["anchors"]:
                    issue(errors, "E_ANCHOR", path.relative_to(root), "Markdown heading anchor does not exist", number)
    for path in sorted(p for p in pages if p.is_relative_to(docs)):
        if path.name != "README.md":
            index = pages.get(path.parent / "README.md")
            if index is None or path not in index["targets"]:
                issue(errors, "E_INDEX", path.relative_to(root), "document is not linked from its directory README")
    for path in sorted(required_indexes - {docs / "README.md"}):
        parent_index = pages.get(path.parent.parent / "README.md")
        if parent_index is None or path not in parent_index["targets"]:
            issue(errors, "E_INDEX", path.relative_to(root), "directory index is not linked from its parent index")
    referenced = set().union(*(p["targets"] for p in pages.values())) if pages else set()
    for path in sorted(p for p in files if p.suffix != ".md"):
        if path not in referenced:
            issue(errors, "E_ORPHAN", path.relative_to(root), "tool or attachment is not referenced by Markdown")
    seen = {}
    for path, page in pages.items():
        digest = hashlib.sha256(page["text"].strip().encode()).digest()
        if digest in seen:
            issue(errors, "E_DUPLICATE", path.relative_to(root), "identical document body already exists")
        else:
            seen[digest] = path
    return errors, len(pages), link_count


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[2],
                        help="repository root (also used for isolated validation fixtures)")
    args = parser.parse_args()
    try:
        errors, count, links = validate(args.root.resolve())
    except (OSError, UnicodeError):
        print("E_READ docs: cannot inspect the documentation tree", file=sys.stderr)
        return 1
    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        print("Documentation gate FAILED: {} issue(s)".format(len(errors)), file=sys.stderr)
        return 1
    print("Documentation gate PASSED: {} Markdown files, {} links".format(count, links))
    return 0


if __name__ == "__main__":
    sys.exit(main())
