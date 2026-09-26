#!/usr/bin/env python3
"""Repository-local development helpers; standard library only."""
import argparse
import hashlib
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent
GENERATORS = ('protoc-gen-go', 'protoc-gen-go-grpc', 'protoc-gen-go-http',
              'protoc-gen-openapi', 'protoc-gen-validate',
              'wire')


def generated(path):
    return (path.name.endswith(('.pb.go', '.pb.validate.go'))
            or path.name == 'wire_gen.go' or path.name == 'openapi.yaml')


def snapshots(root):
    files = [p for directory in ('api', 'internal', 'cmd')
             for p in (root / directory).rglob('*.go') if generated(p)]
    if (root / 'openapi.yaml').exists():
        files.append(root / 'openapi.yaml')
    return {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in files}


def format_sources(write):
    paths = sorted(str(p.relative_to(ROOT)) for directory in ('api', 'internal', 'cmd')
                   for p in (ROOT / directory).rglob('*.go') if not generated(p))
    result = subprocess.run(['gofmt', '-w' if write else '-l', *paths],
                            cwd=ROOT, check=True, capture_output=True, text=True)
    if result.stdout.strip():
        print('需要 gofmt：\n' + result.stdout, file=sys.stderr)
        return 1
    return 0


def check_generated():
    before = snapshots(ROOT)
    with tempfile.TemporaryDirectory(prefix='greeting-generated-') as tmp:
        target = Path(tmp) / 'source'

        def ignore(directory, names):
            omitted = set(names) & {'.git', '.tools', 'bin', '__pycache__', '.idea', '.vscode'}
            omitted.update(n for n in names if n.endswith(('.log', '.key', '.cert')))
            if Path(directory) == ROOT / 'configs':
                omitted.add('config.yaml')
            return omitted

        shutil.copytree(ROOT, target, ignore=ignore)
        subprocess.run(['make', 'all', 'TOOLS_DIR=' + str(ROOT / '.tools/bin')],
                       cwd=target, check=True)
        after = snapshots(target)
    changed = sorted(p for p in before.keys() | after.keys() if before.get(p) != after.get(p))
    if changed:
        print('生成物不一致，请运行 make all 并审查：\n' + '\n'.join(changed), file=sys.stderr)
        return 1
    print('生成代码与当前源码一致（当前工作区未修改）')
    return 0


def doctor():
    failed = False
    for name in ('go', 'gofmt', 'make', 'python3', 'protoc', *GENERATORS,
                 'curl', 'openssl', 'docker'):
        local = ROOT / '.tools/bin' / name
        executable = str(local) if local.is_file() else shutil.which(name)
        required = name in ('go', 'gofmt', 'make', 'python3')
        print(f'{name}: {executable or ("缺失（基础依赖）" if required else "缺失（按需）")}')
        failed |= required and not executable
    print('生成检查需要 protoc 和全部生成器；缺生成器运行 make init。smoke 需要 curl、openssl；broker smoke 另需 Docker。')
    return int(failed)


def task(name):
    if not re.fullmatch(r'[a-z][a-z0-9]*(?:-[a-z0-9]+)*', name):
        raise ValueError('NAME 必须是小写英文 slug，例如 order-create')
    destination = ROOT / 'docs/tasks' / (name + '.md')
    content = (ROOT / 'docs/tasks/TEMPLATE.md').read_text().replace('{{name}}', name)
    with destination.open('x') as out:
        out.write(content)
    print(destination.relative_to(ROOT))
    return 0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=('doctor', 'fmt', 'fmt-check', 'check-generated', 'task'))
    parser.add_argument('--name', default='')
    args = parser.parse_args()
    try:
        return {'doctor': doctor, 'fmt': lambda: format_sources(True),
                'fmt-check': lambda: format_sources(False), 'check-generated': check_generated,
                'task': lambda: task(args.name)}[args.command]()
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        print(str(error), file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
