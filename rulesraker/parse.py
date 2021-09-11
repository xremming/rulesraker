import datetime
import re
import string
from dataclasses import dataclass
from enum import Enum
from os import PathLike
from typing import Iterable, List, Literal, Optional, Tuple, Union


class ParseError(Exception): pass


@dataclass
class Section:
    id: str
    text: str
    type: Literal["section"] = "section"


@dataclass
class Chapter:
    id: str
    text: str
    type: Literal["chapter"] = "chapter"


@dataclass
class Rule:
    id: str
    subrule: bool
    text: List[str]
    examples: List[str]
    type: Literal["rule"] = "rule"


@dataclass
class GlossaryItem:
    id: str
    keys: List[str]
    key_text: str
    text: List[str]


Part = Union[Section, Chapter, Rule]


def try_decode(bs: bytes) -> str:
    encodings = ["utf-8-sig", "cp1252"]

    for encoding in encodings:
        try:
            return bs.decode(encoding)
        except UnicodeDecodeError:
            continue

    raise UnicodeDecodeError(
        "|".join(encodings),
        bs, 0, len(bs),
        f"could not decode input with any of the following encodings: {', '.join(encodings)}"
    )


def fix_inconsistencies(s: str) -> str:
    """There are are a handful of inconsistencies in one file (at the moment) these problems are fixed here."""
    return (s
        .replace("Ability1.", "Ability\n1.")
        .replace("Ante1.", "Ante\n1.")
        .replace("Archenemy1.", "Archenemy\n1.")
        .replace("Draw1.", "Draw\n1.")
        .replace("Exile1.", "Exile\n1.")
        .replace("Face Down1.", "Face Down\n1.")
        .replace("Face Up1.", "Face Up\n1.")
        .replace("Graveyard1.", "Graveyard\n1.")
        .replace("Hand1.", "Hand\n1.")
        .replace("Library1.", "Library\n1.")
        .replace("Loyalty1.", "Loyalty\n1.")
    )


split_re = re.compile(r"\n{2,}")


def split_parts(v: str) -> List[str]:
    return split_re.split("\n".join(v.splitlines()))


class State(Enum):
    start = 0
    rules = 1
    glossary = 2
    end = 3


def separate_parts_and_glossary(vs: List[str]) -> Tuple[List[str], List[str]]:
    parts = []
    glossary = []

    state = State.start
    for part in vs:
        part = part.strip()
        if state == State.start and part == "Credits":
            state = State.rules
            continue
        if state == State.rules and part == "Glossary":
            state = State.glossary
            continue
        if state == State.glossary and part == "Credits":
            state = State.end
            continue
        if state == State.start or state == State.end:
            continue

        if part == "":
            continue

        if state == State.rules:
            parts.append(part)
        if state == State.glossary:
            glossary.append(part)

    return parts, glossary


section_re = re.compile(r"(?P<id>\d\.) (?P<text>.+)")
chapter_re = re.compile(r"(?P<id>\d{3,}\.) (?P<text>.+)")
rule_re = re.compile(r"(?P<id>\d{3,}\.\d+\.?|\d{3,}\.\d+[a-z]\.?) (?P<text>.+)")


def match(rule: str) -> Part:
    if m := rule_re.match(rule):
        return Rule(m["id"], not m["id"].endswith("."), [m["text"]], [])
    elif m := chapter_re.match(rule):
        return Chapter(m["id"], m["text"])
    elif m := section_re.match(rule):
        return Section(m["id"], m["text"])
    else:
        raise ParseError(f"could not parse rule: {rule}")


def parse_rules(rules: List[str]) -> Iterable[Part]:
    for rule in rules:
        lines = list(map(lambda v: v.strip(), rule.splitlines()))

        # the simple case, only one paragraph of rule and no examples
        if len(lines) == 1:
            yield match(rule)
            continue

        text = []
        examples = []
        for line in lines:
            if not line.startswith("Example: "):
                text.append(line)
            else:
                examples.append(line)

        if len(text) == 0:
            raise ParseError("could not parse rule: {rule}")

        if m := rule_re.match(text[0]):
            text[0] = m["text"]
            yield Rule(m["id"], not m["id"].endswith("."), text, examples)
        else:
            raise ParseError(f"could not parse rule: {rule}")


def make_id(s: str) -> str:
    out = []
    for c in s.lower():
        if c in string.ascii_lowercase:
            out.append(c)
        if c == " ":
            out.append("-")
    return "".join(out)


def split_key(v: str) -> List[str]:
    if v == "Active Player, Nonactive Player Order":
        return [v]
    return list(map(lambda s: s.strip(), v.replace("(Obsolete)", "").split(", ")))


def parse_glossary(glossary: List[str]) -> Iterable[GlossaryItem]:
    for item in glossary:
        splitted = item.split("\n", 1)
        if len(splitted) != 2:
            raise ParseError(f"invalid glossary item: {item}")
        key, text = splitted
        yield GlossaryItem(make_id(key), split_key(key), key, text.splitlines())


@dataclass
class Rules:
    parts: List[Part]
    glossary: List[GlossaryItem]
    effective: Optional[datetime.date]


effective_date_re = re.compile("effective as of (\w+ \d+, \d+)\.")


def parse_file(path: PathLike) -> Rules:
    with open(path, "rb") as fp:
        content = try_decode(fp.read())
        fixed = fix_inconsistencies(content)
        splitted = split_parts(fixed)
        rules, glossary = separate_parts_and_glossary(splitted)

        parsed_rules = list(parse_rules(rules))
        parsed_glossary = list(parse_glossary(glossary))

        effective_date = None
        if m := effective_date_re.search(fixed):
            d = m[1]
            dt = datetime.datetime.strptime(m[1], "%B %d, %Y")
            effective_date = datetime.date(dt.year, dt.month, dt.day)
        else:
            raise ParseError("could not parse effective date from file")

        return Rules(parsed_rules, parsed_glossary, effective_date)
