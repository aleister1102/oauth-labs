import re
import os
from typing import Set

# This script helps me double-check if I messed up with cfg.Get... calls. It
# walks over all the lab dirs and checks if the strings within calls to cfg.Get
# exist within the corresponding config.go file.


def get_config_strings(config_file: str) -> Set[str]:
    string_re = re.compile(r'^\s*cfg.SetDefault\("(.+?)",\s.*')
    strings = set()
    with open(config_file) as fin:
        for line in fin:
            line = line.strip()
            match = string_re.search(line)
            if not match:
                continue
            strings.add(match.group(1))
    return strings


def get_config_retrievals(go_file: str) -> Set[str]:
    string_re = re.compile(r'cfg.Get\w+\("(.+?)"\)')
    with open(go_file) as fin:
        data = fin.read()
    strings = string_re.findall(data)
    return set(strings)


def process_dir(lab_dir: str) -> bool:
    has_errors = False
    config_file = os.path.join(lab_dir, "internal", "config", "config.go")
    config_strings = get_config_strings(config_file)
    for root, _, files in os.walk(lab_dir):
        for f in files:
            if f.endswith(".go") is False:
                continue
            go_file = os.path.join(root, f)
            if go_file == config_file:
                continue
            retrievals = get_config_retrievals(go_file)
            if not retrievals:
                continue

            for retr in retrievals:
                if retr not in config_strings:
                    print(f"Error: {go_file!r}: config option {retr!r} not found.")
                    has_errors = True

    return has_errors


def main():
    lab_dir_re = re.compile(r"^lab\d{2}")
    base_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    lab_dirs = [ld for ld in os.listdir(base_dir) if lab_dir_re.match(ld)]
    has_errors = False
    for lab_dir in lab_dirs:
        lab_dir = os.path.join(base_dir, lab_dir)
        for component in ("server", "client"):
            comp_dir = os.path.join(lab_dir, component)
            errs = process_dir(comp_dir)
            if errs:
                has_errors = True

    if has_errors:
        return 1
    return 0


if __name__ == "__main__":
    exit(main())
