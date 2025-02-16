"""
Download comprehensive rules from the Wayback Machine. Even with the retries the
script may fail to download some files, so it's recommended to run it multiple times
until all files are downloaded.

The script will save the files in the `downloads` directory and create a JSON file
`_found_files.json` with the metadata of the found files. The `Date` field is guessed
from the URL and needs to manually corrected to the right format. After that is done
the files can be imported to the archive with the `import_files.py` script.
"""

import json
import re
import time
from pathlib import Path
from typing import TypedDict

import requests

base_url = "https://web.archive.org/cdx/search/xd"

base_params = {
    "collapse": "digest",
    "filter": "statuscode:200",
    "output": "json",
    "url": "http://wizards.com/magic/comprules/*",
}


class Snapshot(TypedDict):
    urlkey: str
    timestamp: str
    original: str
    mimetype: str
    statuscode: str
    digest: str
    length: str


def get_snapshots() -> list[Snapshot]:
    response = requests.get(base_url, params=base_params)
    snapshots = response.json()

    headers = snapshots[0]
    snapshots_dict = [dict(zip(headers, snapshot)) for snapshot in snapshots[1:]]
    return snapshots_dict


def archive_url(snapshot: Snapshot) -> str:
    return f"https://web.archive.org/web/{snapshot['timestamp']}/{snapshot['original']}"


def download_url(snapshot: Snapshot) -> str:
    return (
        f"https://web.archive.org/web/{snapshot['timestamp']}id_/{snapshot['original']}"
    )


def download_file(snapshot: Snapshot, download_dir: Path) -> tuple[Path, bool]:
    download_dir.mkdir(exist_ok=True)
    file_path = download_dir / Path(snapshot["original"]).name
    file_path = file_path.with_name(
        f"{file_path.stem}_{snapshot['digest']}{file_path.suffix}"
    )

    if file_path.exists():
        print(f"File {file_path} already exists, skipping download")
        return file_path, False

    url = download_url(snapshot)

    print(f"Requesting {url}")
    response = requests.get(url, stream=True, headers={"Accept-Encoding": "identity"})
    response.raise_for_status()

    print(f"Downloading to {file_path}")
    with file_path.open("wb") as file:
        for chunk in response.iter_content(chunk_size=None):
            file.write(chunk)

    return file_path, True


def date_guess(snapshot: Snapshot) -> str:
    match = re.findall(r"\d+", snapshot["original"])
    if match:
        return max(match, key=len)
    return snapshot["timestamp"]


def get_format(snapshot: Snapshot) -> str:
    match snapshot["mimetype"]:
        case "text/plain":
            return "txt"
        case "application/pdf":
            return "pdf"
        case "application/msword":
            return "docx"
        case "application/rtf":
            return "rtf"

    return snapshot["mimetype"]


class FoundFile(TypedDict):
    Date: str
    Format: str
    File: str
    OriginalURL: str
    Source: str
    Comment: str


if __name__ == "__main__":
    download_dir = Path("downloads")

    found_files: list[FoundFile] = []

    print("Searching for snapshots")
    snapshots = get_snapshots()

    print(f"Found {len(snapshots)} snapshots")

    print("Removing duplicates")
    seen_digests = set()
    unique_snapshots = []

    for snapshot in snapshots:
        if snapshot["digest"] not in seen_digests:
            seen_digests.add(snapshot["digest"])
            unique_snapshots.append(snapshot)

    print(f"Found {len(unique_snapshots)} unique snapshots")
    snapshots = unique_snapshots

    for snapshot in snapshots:
        if snapshot["mimetype"] == "text/html":
            continue

        max_attempts = 5
        for attempt in range(max_attempts):
            try:
                path_saved, file_downloaded = download_file(snapshot, download_dir)
                break
            except requests.RequestException as e:
                print(f"Attempt {attempt + 1} failed: {e}")
                if attempt < max_attempts - 1:
                    sleep_time = min(5 * (2**attempt), 30)
                    print(f"Retrying in {sleep_time} seconds...")
                    time.sleep(sleep_time)
                else:
                    print(f"Failed to download after {max_attempts} attempts.")
                    path_saved, file_downloaded = None, False

        found_files.append(
            {
                "Date": date_guess(snapshot),
                "Format": get_format(snapshot),
                "File": str(path_saved),
                "OriginalURL": snapshot["original"],
                "Source": archive_url(snapshot),
                "Comment": "Found from the Wayback Machine.",
            }
        )

        found_files.sort(key=lambda x: x["File"])
        with (download_dir / "_found_files.json").open("w") as file:
            json.dump(found_files, file, indent=2)

        if file_downloaded:
            print("Sleeping for 5 seconds to avoid rate limiting...")
            time.sleep(5)
