import argparse
from pathlib import Path
from typing import Callable


def path_to_dir() -> Callable[[str], Path]:
    def inner(v: str) -> Path:
        out = Path(v)
        if not out.is_dir() and out.exists():
            raise argparse.ArgumentTypeError(f"path exists and is not a directory: {v}")
        out.mkdir(parents=True, exist_ok=True)
        return out
    return inner
