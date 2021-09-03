from argparse import ArgumentTypeError
from pathlib import Path
from typing import Callable


def path_to_dir(*, must_exist=False, mkdir=True) -> Callable[[str], Path]:
    def inner(v: str) -> Path:
        out = Path(v)

        if must_exist and not out.exists():
            raise ArgumentTypeError(f"path does not exist: {v}")
        if not out.is_dir():
            raise ArgumentTypeError(f"path is not a directory: {v}")

        if mkdir:
            out.mkdir(parents=True, exist_ok=True)

        return out
    return inner


def path_to_file(*, must_exist=False) -> Callable[[str], Path]:
    def inner(v: str) -> Path:
        out = Path(v)

        if must_exist and not out.exists():
            raise ArgumentTypeError(f"path does not exist: {v}")
        if not out.is_file():
            raise ArgumentTypeError(f"path is not a file: {v}")

        return out
    return inner
