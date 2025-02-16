"""
Copy all files from the `downloads` directory to the `archive` directory. Following
the naming convention of the Go application. This requires manually correcting the
`Date` field in the `_found_files.json` file.

The final list of found files will be out put to `found_files.json` from where it
can be copied to the archive `metadata.json` file's `FoundFiles` field.
"""

import json
from pathlib import Path

if __name__ == "__main__":
    download_dir = Path("downloads")
    archive_dir = Path("archive")

    with (download_dir / "_found_files.json").open() as f:
        found_files = json.load(f)

    for file in found_files:
        file_path = Path(file["File"])
        new_file_path = (
            archive_dir / file["Format"] / f"{file['Date']}.{file['Format']}"
        )
        new_file_path.parent.mkdir(parents=True, exist_ok=True)

        if new_file_path.exists():
            print(f"File {new_file_path} already exists. Skipping")
            continue

        file_path.replace(new_file_path)
        file["File"] = f"{file['Format']}/{file['Date']}.{file['Format']}"

    with (download_dir / "found_files.json").open("w") as f:
        json.dump(found_files, f, indent=2)
