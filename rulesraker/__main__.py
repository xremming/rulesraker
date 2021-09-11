import argparse
import json
import sys
from dataclasses import asdict
from pathlib import Path
from typing import List, Optional

from rulesraker.parse import parse_file
from rulesraker.render import render

from .download import find_rules_files
from .utils import path_to_dir, path_to_file


def do_find_files(args: argparse.Namespace) -> Optional[int]:
    for rule in find_rules_files(args.start_year):
        print(rule.url)


def do_download(args: argparse.Namespace) -> Optional[int]:
    d = Path(args.directory)
    d.mkdir(parents=True, exist_ok=True)
    for rule in find_rules_files(args.start_year):
        with open(d / rule.filename, "wb") as fp:
            fp.write(rule.get())


def do_parse(args: argparse.Namespace) -> Optional[int]:
    for path in args.file:
        rules = parse_file(path)
        print(json.dumps(asdict(rules), default=str))


def do_render(args: argparse.Namespace) -> Optional[int]:
    for path in args.file:
        rules = parse_file(path)
        print(render(rules))


def main(args: List[str]) -> Optional[int]:
    parser = argparse.ArgumentParser(args[0])
    parser.set_defaults(_do=None)

    subparser = parser.add_subparsers()

    find_files = subparser.add_parser("find-files")
    find_files.set_defaults(_do=do_find_files)
    find_files.add_argument("--start-year", "-y", type=int, nargs="?")

    download = subparser.add_parser("download")
    download.set_defaults(_do=do_download)
    download.add_argument("--directory", "-d", type=path_to_dir(), default="./rules/")
    download.add_argument("--start-year", "-y", type=int, nargs="?")

    parse = subparser.add_parser("parse")
    parse.set_defaults(_do=do_parse)
    parse.add_argument("file", type=path_to_file(must_exist=True), nargs="+")

    render = subparser.add_parser("render")
    render.set_defaults(_do=do_render)
    render.add_argument("file", type=path_to_file(must_exist=True), nargs="+")

    args = parser.parse_args(args[1:])

    if args._do is None:
        parser.print_help()
        return 2

    return args._do(args)


if __name__ == "__main__":
    code = main(sys.argv)
    if code is not None:
        sys.exit(code)
