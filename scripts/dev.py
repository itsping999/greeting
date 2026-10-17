#!/usr/bin/env python3
"""Repository-local development helpers; standard library only."""
import argparse
from datetime import date
import hashlib
import importlib.util
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parent.parent
GENERATORS = ('protoc-gen-go', 'protoc-gen-go-grpc', 'protoc-gen-go-http',
              'protoc-gen-openapi', 'protoc-gen-validate',
              'wire', 'kratos')


def generated(path):
    return (path.name.endswith(('.pb.go', '.pb.validate.go'))
            or path.name == 'wire_gen.go' or path.name == 'openapi.yaml')


def snapshots(root):
    files = [p for directory in ('api', 'internal', 'cmd', 'tests')
             for p in (root / directory).rglob('*.go') if generated(p)]
    if (root / 'openapi.yaml').exists():
        files.append(root / 'openapi.yaml')
    return {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in files}


def format_sources(write):
    paths = sorted(str(p.relative_to(ROOT)) for directory in ('api', 'internal', 'cmd', 'tests')
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


def scaffold_smoke():
    with tempfile.TemporaryDirectory(prefix='greeting-scaffold-') as tmp:
        target = Path(tmp) / 'source'
        def ignore(directory, names):
            omitted = set(names) & {'.git', '.tools', 'bin', '__pycache__', '.idea', '.vscode'}
            omitted.update(n for n in names if n.endswith(('.log', '.key', '.cert')))
            if Path(directory) == ROOT / 'configs':
                omitted.add('config.yaml')
            return omitted
        shutil.copytree(ROOT, target, ignore=ignore)
        (target / '.tools').symlink_to(ROOT / '.tools', target_is_directory=True)
        shutil.copyfile(ROOT / 'configs/container.yaml', target / 'configs/config.yaml')
        for proto, crud in [('invoice/v1/invoice.proto', '0'), ('order/v1/order.proto', '1')]:
            subprocess.run(['make', 'scaffold-module', 'PROTO=' + proto, 'CRUD=' + crud],
                           cwd=target, check=True)
        subprocess.run(['go', 'test', './...'], cwd=target, check=True)
        subprocess.run(['make', 'check-generated'], cwd=target, check=True)
    print('普通和 CRUD 模块的生成、注册、Wire、编译及生成一致性通过（当前源码未修改）')
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
    spec = importlib.util.spec_from_file_location(
        'documentation_gate', ROOT / 'docs/development/check_docs.py')
    checker = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(checker)
    if name == 'readme' or not checker.valid_name(name + '.md'):
        raise ValueError('NAME 必须是稳定主题名称，不能使用临时、最新或副本名称')

    directory = ROOT / 'docs/plans'
    destination = directory / (name + '.md')
    index = directory / 'README.md'
    index_text = index.read_text(encoding='utf-8')
    start, end = '<!-- task-index:start -->', '<!-- task-index:end -->'
    if index_text.count(start) != 1 or index_text.count(end) != 1:
        raise ValueError('计划索引必须包含唯一活动计划标记')
    if index_text.index(start) >= index_text.index(end):
        raise ValueError('计划索引的活动计划标记顺序错误')
    content = (directory / 'task-template.md').read_text(encoding='utf-8')
    content = content.replace('{{name}}', name).replace('{{date}}', date.today().isoformat())
    entry = f'| [{name}]({name}.md) | Draft | 待填写范围 |\n'
    updated_index = index_text.replace(end, entry + end, 1)
    with destination.open('x', encoding='utf-8') as out:
        out.write(content)
    try:
        index.write_text(updated_index, encoding='utf-8')
    except OSError:
        destination.unlink()
        raise
    print(destination.relative_to(ROOT))
    return 0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=('doctor', 'fmt', 'fmt-check', 'check-generated', 'scaffold-smoke', 'task'))
    parser.add_argument('--name', default='')
    args = parser.parse_args()
    try:
        return {'doctor': doctor, 'fmt': lambda: format_sources(True),
                'fmt-check': lambda: format_sources(False), 'check-generated': check_generated,
                'scaffold-smoke': scaffold_smoke, 'task': lambda: task(args.name)}[args.command]()
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        print(str(error), file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
