import concurrent.futures
import datetime
from dataclasses import dataclass
from typing import Iterable

import requests


@dataclass
class RulesFile:
    year: int
    month: int
    day: int
    url: str
    filename: str

    def get(self, sess=None):
        if sess is None:
            sess = requests
        resp = sess.get(self.url)
        resp.raise_for_status()
        return resp.text


def check(sess, year, month, day):
    url = f"https://media.wizards.com/{year}/downloads/MagicCompRules%20{year}{month:02}{day:02}.txt"
    resp = sess.head(url)
    if resp.status_code == 200:
        return RulesFile(year, month, day, url, f"{year}-{month:02}-{day:02}.txt")
    return None


def find_rules_files(start_year: int = None) -> Iterable[RulesFile]:
    if start_year is None:
        start_year = datetime.date.today().year
    if start_year <= 2016:
        raise ValueError("start_year must be >= 2017: no rules files exist before 2017")

    sess = requests.Session()

    with concurrent.futures.ThreadPoolExecutor() as executor:
        results = []

        for year in range(start_year, 2016, -1):
            for month in range(1, 13):
                for day in range(1, 32):
                    # print(f"{year}-{month:02}-{day:02}")
                    results.append(executor.submit(check, sess, year, month, day))

        for result in concurrent.futures.as_completed(results):
            res = result.result()
            if res is not None:
                yield res
