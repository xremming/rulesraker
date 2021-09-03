import argparse
import sys
from pathlib import Path
from typing import List, Optional

from .download import find_rules_files
from .utils import path_to_dir


def do_find_files(args: argparse.Namespace) -> Optional[int]:
    for rule in find_rules_files(args.start_year):
        print(rule.url)


def do_download(args: argparse.Namespace) -> Optional[int]:
    d = Path(args.directory)
    d.mkdir(parents=True, exist_ok=True)
    for rule in find_rules_files(args.start_year):
        with open(d / rule.filename, "wb") as fp:
            fp.write(rule.get())


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

    args = parser.parse_args(args[1:])

    if args._do is None:
        parser.print_help()
        return 2

    return args._do(args)


if __name__ == "__main__":
    code = main(sys.argv)
    if code is not None:
        sys.exit(code)
