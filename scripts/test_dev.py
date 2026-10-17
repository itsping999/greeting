"""Exercise task creation through the CLI in an isolated documentation tree."""
from datetime import date
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parent.parent


class TaskCommandTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='greeting-task-')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        (self.root / 'scripts').mkdir()
        (self.root / 'api').mkdir()
        plans = self.root / 'docs/plans'
        development = self.root / 'docs/development'
        plans.mkdir(parents=True)
        development.mkdir()
        shutil.copyfile(ROOT / 'scripts/dev.py', self.root / 'scripts/dev.py')
        shutil.copyfile(ROOT / 'Makefile', self.root / 'Makefile')
        shutil.copyfile(ROOT / 'docs/development/check_docs.py', development / 'check_docs.py')
        shutil.copyfile(ROOT / 'docs/plans/task-template.md', plans / 'task-template.md')
        self.index = plans / 'README.md'
        self.index.write_text(
            '# Plans\n\n[template](task-template.md)\n\n'
            '<!-- task-index:start -->\n'
            '| Plan | Status | Scope |\n| --- | --- | --- |\n'
            '<!-- task-index:end -->\n', encoding='utf-8')
        (development / 'README.md').write_text(
            '# Development\n[checker](check_docs.py)\n', encoding='utf-8')
        (self.root / 'docs/README.md').write_text(
            '# Documentation\n[standard](documentation-standard.md)\n'
            '[plans](plans/README.md)\n[development](development/README.md)\n',
            encoding='utf-8')
        standard = (ROOT / 'docs/documentation-standard.md').read_text(encoding='utf-8')
        policy = standard.split('<!-- docs-check-policy:start -->', 1)[1]
        policy = policy.split('<!-- docs-check-policy:end -->', 1)[0]
        rules = json.loads(policy.split('```json\n', 1)[1].rsplit('\n```', 1)[0])
        (self.root / 'docs/documentation-standard.md').write_text(
            '# Standard\n<!-- docs-check-policy:start -->\n```json\n'
            + json.dumps(rules) + '\n```\n<!-- docs-check-policy:end -->\n',
            encoding='utf-8')

    def task(self, name):
        return subprocess.run(
            [sys.executable, '-B', 'scripts/dev.py', 'task', '--name', name],
            cwd=self.root, capture_output=True, text=True)

    def snapshot(self):
        return {str(path.relative_to(self.root)): path.read_bytes()
                for path in self.root.rglob('*') if path.is_file()}

    def test_created_plan_is_indexed_and_passes_shared_gate(self):
        for name in ('order-create', 'invoice-export'):
            result = subprocess.run(
                ['make', 'task', 'NAME=' + name, 'DOCS_PYTHON=' + sys.executable],
                cwd=self.root, capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout.strip().splitlines()[-1], 'docs/plans/' + name + '.md')
            content = (self.root / 'docs/plans' / (name + '.md')).read_text(encoding='utf-8')
            self.assertIn('# ' + name, content)
            self.assertIn('状态：Draft', content)
            self.assertIn(date.today().isoformat(), content)
            self.assertNotIn('{{', content)
            self.assertEqual(self.index.read_text(encoding='utf-8').count('](' + name + '.md)'), 1)
        result = subprocess.run(
            [sys.executable, '-B', 'docs/development/check_docs.py'],
            cwd=self.root, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertFalse((self.root / 'docs/tasks').exists())
        self.assertFalse(any(self.root.rglob('__pycache__')))

    def test_existing_plan_is_not_overwritten(self):
        self.assertEqual(self.task('order-create').returncode, 0)
        before = self.snapshot()
        result = self.task('order-create')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(before, self.snapshot())

    def test_invalid_names_do_not_change_documents(self):
        for name in ('', '../outside', 'BadName', 'latest', 'draft', 'subject-copy',
                     'readme', 'task-template'):
            with self.subTest(name=name):
                before = self.snapshot()
                result = self.task(name)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(before, self.snapshot())

    def test_invalid_index_is_rejected_before_creating_a_plan(self):
        text = self.index.read_text(encoding='utf-8')
        for invalid in (
                text.replace('<!-- task-index:end -->', ''),
                text.replace('<!-- task-index:start -->', '<!-- task-index:start -->' * 2),
                text.replace('<!-- task-index:start -->', '<!-- swap -->')
                    .replace('<!-- task-index:end -->', '<!-- task-index:start -->')
                    .replace('<!-- swap -->', '<!-- task-index:end -->')):
            with self.subTest(index=invalid):
                self.index.write_text(invalid, encoding='utf-8')
                before = self.snapshot()
                result = self.task('order-create')
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(before, self.snapshot())


if __name__ == '__main__':
    unittest.main()
