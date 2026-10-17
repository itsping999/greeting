"""CLI regression fixtures for the documentation completion gate."""
import hashlib
import json
import shutil
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[2]
CHECKER = ROOT / "docs/development/check_docs.py"


def policy():
    text = (ROOT / "docs/documentation-standard.md").read_text(encoding="utf-8")
    block = text.split("<!-- docs-check-policy:start -->", 1)[1].split("<!-- docs-check-policy:end -->", 1)[0]
    return json.loads(block.split("```json\n", 1)[1].rsplit("\n```", 1)[0])


def fixture(base):
    docs = base / "docs"
    docs.mkdir()
    rules = policy()
    directories = rules["allowed_directories"] + rules["allowed_subdirectories"]
    for name in directories:
        (docs / name).mkdir(parents=True, exist_ok=True)
    for name in rules["allowed_tool_files"]:
        (docs / name).write_bytes((ROOT / "docs" / name).read_bytes())
    for name in directories:
        links = []
        for child in directories:
            if str(Path(child).parent) == name:
                links.append("[child]({}/README.md)".format(Path(child).name))
        for tool in rules["allowed_tool_files"]:
            if str(Path(tool).parent) == name:
                links.append("[tool]({})".format(Path(tool).name))
        (docs / name / "README.md").write_text("# index " + name + "\n\n" + "\n".join(links), encoding="utf-8")
    for name in rules["allowed_root_files"]:
        if name != "README.md":
            (docs / name).write_text("# root " + name + "\n", encoding="utf-8")
    (docs / "documentation-standard.md").write_text(
        "# standard\n\n<!-- docs-check-policy:start -->\n```json\n" + json.dumps(rules)
        + "\n```\n<!-- docs-check-policy:end -->\n", encoding="utf-8")
    links = ["[entry]({})".format(name) for name in rules["allowed_root_files"] if name != "README.md"]
    links += ["[category]({}/README.md)".format(name) for name in rules["allowed_directories"]]
    (docs / "README.md").write_text("# documentation\n\n" + "\n".join(links), encoding="utf-8")
    (base / "AGENTS.md").write_text("# fixture agent\n", encoding="utf-8")
    return docs


def page(docs, name, body="# subject\n", indexed=True):
    path = docs / name
    path.write_text(body, encoding="utf-8")
    if indexed:
        index = path.parent / "README.md"
        index.write_text(index.read_text(encoding="utf-8") + "\n[subject]({})\n".format(path.name), encoding="utf-8")


def replace(path, before, after):
    text = path.read_text(encoding="utf-8")
    assert before in text
    path.write_text(text.replace(before, after), encoding="utf-8")


def positive_names(docs):
    page(docs, "interfaces/webscada-v1.9.8-compatibility-audit.md", "# versioned API\n")
    page(docs, "interfaces/通用请求约定.md", "# 中文 标题\n## 中文 标题\n## 中文-标题-1\n## 中文 标题\n")
    index = docs / "interfaces/README.md"
    index.write_text(index.read_text(encoding="utf-8") + "\n[encoded](" + quote("通用请求约定.md") + "#中文-标题-2)\n", encoding="utf-8")


def reference_index(docs):
    page(docs, "interfaces/references.md", "# reference topic\n", indexed=False)
    index = docs / "interfaces/README.md"
    index.write_text(index.read_text(encoding="utf-8") + "\n[topic][ref]\n[ref]: references.md\n[ref]\n", encoding="utf-8")


def attachment(docs):
    (docs / "interfaces/example.json").write_text("{}")
    page(docs, "interfaces/example.md", "# example\n[attachment](example.json)\n")


def duplicate(docs):
    page(docs, "interfaces/first.md", "# identical\n")
    page(docs, "interfaces/second.md", "# identical\n")


def change_policy(docs, **updates):
    path = docs / "documentation-standard.md"
    text = path.read_text(encoding="utf-8")
    before, rest = text.split("```json\n", 1)
    block, after = rest.split("\n```", 1)
    rules = json.loads(block)
    rules.update(updates)
    path.write_text(before + "```json\n" + json.dumps(rules) + "\n```" + after, encoding="utf-8")


def topic_subdirectory(docs):
    rules = policy()
    change_policy(docs, allowed_subdirectories=rules["allowed_subdirectories"] + ["interfaces/http"])
    (docs / "interfaces/http").mkdir()
    (docs / "interfaces/http/README.md").write_text("# HTTP contracts\n", encoding="utf-8")
    index = docs / "interfaces/README.md"
    index.write_text(index.read_text(encoding="utf-8") + "\n[HTTP](http/README.md)\n", encoding="utf-8")
    page(docs, "interfaces/http/lookup.md", "# HTTP lookup\nAn indexed topic contract.\n")


def cases():
    return [
        ("valid", lambda d: None, None),
        ("version-and-Chinese", positive_names, None),
        ("reference-index", reference_index, None),
        ("referenced-attachment", attachment, None),
        ("code-links-ignored", lambda d: page(d, "interfaces/code.md", "# code\n```md\n[ignored](missing.md)\n```\n`[ignored](missing.md)`\n"), None),
        ("Mac-metadata", lambda d: (d / ".DS_Store").write_bytes(b"metadata"), None),
        ("declared-topic-subdirectory", topic_subdirectory, None),
        ("extra-root", lambda d: (d / "random.md").write_text("# random\n"), "E_ROOT_FILE"),
        ("extra-directory", lambda d: (d / "design").mkdir(), "E_DIRECTORY"),
        ("extra-level", lambda d: (d / "interfaces/module").mkdir(), "E_DIRECTORY"),
        ("missing-index", lambda d: (d / "interfaces/README.md").unlink(), "E_REQUIRED"),
        ("missing-parent-link", lambda d: replace(d / "README.md", "[category](interfaces/README.md)", "interfaces"), "E_INDEX"),
        ("unindexed", lambda d: page(d, "interfaces/unindexed.md", indexed=False), "E_INDEX"),
        ("broken-target", lambda d: page(d, "interfaces/broken.md", "# subject\n[missing](missing.md)\n"), "E_LINK"),
        ("broken-anchor", lambda d: page(d, "interfaces/broken.md", "# subject\n[missing](README.md#not-present)\n"), "E_ANCHOR"),
        ("temporary-name", lambda d: page(d, "interfaces/latest.md"), "E_NAME"),
        ("copy-name", lambda d: page(d, "interfaces/subject-copy.md"), "E_NAME"),
        ("symlink", lambda d: (d / "interfaces/alias.md").symlink_to(d / "interfaces/README.md"), "E_SYMLINK"),
        ("runtime-artifact", lambda d: (d / "interfaces/server.log").write_text("not documentation"), "E_ARTIFACT"),
        ("orphan-attachment", lambda d: (d / "interfaces/example.json").write_text("{}"), "E_ORPHAN"),
        ("identical-body", duplicate, "E_DUPLICATE"),
        ("wrong-index-case", lambda d: (d / "interfaces/README.md").rename(d / "interfaces/readme.md"), "E_REQUIRED"),
        ("cache-directory", lambda d: (d / "development/__pycache__").mkdir(), "E_DIRECTORY"),
        ("policy-type", lambda d: replace(d / "documentation-standard.md", '"version": 2', '"version": true'), "E_POLICY"),
        ("duplicate-policy-key", lambda d: replace(d / "documentation-standard.md", '"version": 2', '"version": 2, "version": 2'), "E_POLICY"),
        ("code-fence", lambda d: page(d, "interfaces/fenced.md", "# subject\n```json\n{}\n"), "E_FENCE"),
        ("undefined-reference", lambda d: page(d, "interfaces/refs.md", "# subject\n[missing][never-defined]\n"), "E_REFERENCE"),
        ("outside-link", lambda d: page(d, "interfaces/escape.md", "# subject\n[escape](../../../outside.md)\n"), "E_LINK"),
        ("outside-policy-path", lambda d: replace(d / "documentation-standard.md", '"allowed_subdirectories": ["architecture/decisions"]', '"allowed_subdirectories": ["../outside"]'), "E_POLICY"),
        ("explicit-required-document", lambda d: change_policy(d, required_files=policy()["required_files"] + ["interfaces/contract.md"]), "E_REQUIRED"),
        ("required-outside-category", lambda d: change_policy(d, required_files=policy()["required_files"] + ["other/contract.md"]), "E_POLICY"),
        ("missing-required-root", lambda d: change_policy(d, required_files=["README.md"]), "E_POLICY"),
        ("third-level-policy", lambda d: change_policy(d, allowed_subdirectories=["interfaces/http", "interfaces/http/v2"]), "E_POLICY"),
        ("third-level-directory", lambda d: (d / "architecture/decisions/topic").mkdir(), "E_DIRECTORY"),
        ("missing-topic-index", lambda d: (topic_subdirectory(d), (d / "interfaces/http/README.md").unlink()), "E_REQUIRED"),
    ]


class DocumentationGateTest(unittest.TestCase):
    def test_allowed_categories_need_not_exist(self):
        for category in (None, "interfaces", "architecture"):
            with self.subTest(category=category), tempfile.TemporaryDirectory(prefix="software-doc-gate-") as directory:
                base = Path(directory)
                docs = fixture(base)
                for child in docs.iterdir():
                    if child.is_dir() and child.name != category:
                        shutil.rmtree(child)
                links = "# minimal project\n[standard](documentation-standard.md)\n"
                if category:
                    for child in (docs / category).iterdir():
                        if child.is_dir():
                            shutil.rmtree(child)
                    (docs / category / "README.md").write_text("# category " + category + "\n", encoding="utf-8")
                    page(docs, category + "/boundary.md", "# One real topic\nA small project reference.\n")
                    links += "[category]({}/README.md)\n".format(category)
                (docs / "README.md").write_text(links, encoding="utf-8")
                result = subprocess.run([sys.executable, "-B", str(CHECKER), "--root", str(base)],
                                        capture_output=True, text=True)
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_cli_accepts_valid_and_rejects_invalid_fixtures(self):
        for name, mutate, expected in cases():
            with self.subTest(name=name), tempfile.TemporaryDirectory(prefix="software-doc-gate-") as directory:
                base = Path(directory)
                docs = fixture(base)
                mutate(docs)
                result = subprocess.run([sys.executable, "-B", str(CHECKER), "--root", str(base)],
                                        capture_output=True, text=True)
                output = result.stdout + result.stderr
                self.assertEqual(result.returncode, 0 if expected is None else 1, output)
                if expected:
                    self.assertIn(expected, output)

    def test_real_tree_is_read_only(self):
        def snapshot():
            return {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).digest()
                    for p in (ROOT / "docs").rglob("*") if p.is_file()}
        before = snapshot()
        result = subprocess.run([sys.executable, "-B", str(CHECKER)], cwd=ROOT,
                                capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(before, snapshot())
        self.assertFalse(any(p.name == "__pycache__" for p in (ROOT / "docs").rglob("*")))


if __name__ == "__main__":
    unittest.main()
