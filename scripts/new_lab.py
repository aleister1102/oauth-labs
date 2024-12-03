import sys
import re
from dataclasses import dataclass
import uuid
import secrets
import string
import subprocess
import os
import shutil
import argparse


LAB_DIR_RE = re.compile(r"^lab(\d{2})$")
BASE_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
CADDYFILE = os.path.join(BASE_DIR, "docker", "caddy", "Caddyfile")
COMPOSE_FILE = os.path.join(BASE_DIR, "docker-compose.yaml")


@dataclass
class Config:
    lab_number: str
    registration_secret: str
    database_host: str
    client_id: str
    client_secret: str
    client_cookie_secret: str
    client_database_password: str
    server_cookie_secret: str
    server_encryption_key: str
    server_database_password: str
    private_key: str


def append_to_caddy(lab_number: str):
    block = """
server-§LAB_NUMBER§.oauth.labs {
    import common
    header {
        server: server-§LAB_NUMBER§
    }
    reverse_proxy server-§LAB_NUMBER§:3000
}

client-§LAB_NUMBER§.oauth.labs {
    import common
    header {
        server: client-§LAB_NUMBER§
    }
    reverse_proxy client-§LAB_NUMBER§:3000
}\
""".replace("§LAB_NUMBER§", lab_number)
    with open(CADDYFILE, "a") as fout:
        fout.write(block)


def append_to_compose(lab_number: str):
    compose_block = f"""\
  server-{lab_number}:
    image: docker.io/library/server-{lab_number}
    build:
      context: .
      dockerfile: ./docker/Dockerfile.baselab
      args:
        - 'LAB_NUMBER={lab_number}'
        - 'COMPONENT=server'
    networks:
      - oauth-labs
    depends_on:
      - caddy
      - db
      - valkey
    volumes:
      - ./docker/lab{lab_number}/server.config.yaml:/app/config.yaml
    cpus: 1
    mem_limit: 1g
  
  client-{lab_number}:
    image: docker.io/library/client-{lab_number}
    build:
      context: .
      dockerfile: ./docker/Dockerfile.baselab
      args:
        - 'LAB_NUMBER={lab_number}'
        - 'COMPONENT=client'
    networks:
      - oauth-labs
    depends_on:
      - caddy
      - db
      - valkey
      - server-{lab_number}
    volumes:
      - ./docker/lab{lab_number}/client.config.yaml:/app/config.yaml
    cpus: 1
    mem_limit: 1g

"""
    with open(COMPOSE_FILE) as fin:
        compose_head = ""
        compose_tail = ""
        for line in fin:
            if line.startswith("secrets:"):
                compose_tail += line
                break
            compose_head += line
        for line in fin:
            compose_tail += line

    compose = compose_head + compose_block + compose_tail

    with open(COMPOSE_FILE, "w") as fout:
        fout.write(compose)

def _write_sql_init(outfile: str, config: Config):
    block = f"""
-- lab{config.lab_number}
CREATE USER 'server{config.lab_number}'@'%' IDENTIFIED BY '{config.server_database_password}';
CREATE DATABASE server{config.lab_number};
GRANT ALL PRIVILEGES ON server{config.lab_number}.* TO 'server{config.lab_number}'@'%';
CREATE USER 'client{config.lab_number}'@'%' IDENTIFIED BY '{config.client_database_password}';
CREATE DATABASE client{config.lab_number};
GRANT ALL PRIVILEGES ON client{config.lab_number}.* TO 'client{config.lab_number}'@'%';
"""

    with open(outfile, "a") as fout:
        fout.write(block)


def append_to_sql_init(config: Config, dev_config: Config):
    _write_sql_init(os.path.join(BASE_DIR, "docker", "db", "init.prod.sql"), config)
    _write_sql_init(os.path.join(BASE_DIR, "docker", "db", "init.dev.sql"), dev_config)

def new_password(length: int) -> str:
    charset = string.ascii_letters + string.digits
    return "".join([secrets.choice(charset) for _ in range(0, length + 1)])


def new_rsa_private_key() -> str:
    cmd = ["openssl", "genrsa", "-out", "-", "2048"]
    proc = subprocess.run(cmd, capture_output=True)
    raw_key = proc.stdout.decode("utf8")
    formatted_key = ""
    for i, line in enumerate(raw_key.splitlines(keepends=False)):
        if i != 0:
            line = " " * 4 + line
        formatted_key += line + "\n"
    return formatted_key


def write_dev_config_files(config: Config):
    lab_dir = os.path.join(BASE_DIR, f"lab{config.lab_number}")
    server_config = f"""\
server:
  host: '127.0.0.1'
  port: 3000

database:
  host: '127.0.0.1'
  port: 3306
  name: 'server{config.lab_number}'
  username: 'server{config.lab_number}'
  password: '{config.server_database_password}'

cookie:
  secret: '{config.server_cookie_secret}'
  path: '/'
  secure: false
  domain: '127.0.0.1'

redis:
  host: '127.0.0.1'
  port: 6379
  database: 0

oauth:
  issuer: 'http://127.0.0.1:3000'
  registration_secret: '{config.registration_secret}'
  allowed_clients:
    - '{config.client_id}'
  encryption_key: '{config.server_encryption_key}'
  private_key: |
    {config.private_key}
"""
    server_config_file = os.path.join(lab_dir, "server", "config.yaml")
    with open(server_config_file, "w") as fout:
        fout.write(server_config)

    client_config = f"""\
server:
  host: '127.0.0.1'
  port: 3001

database:
  host: '127.0.0.1'
  port: 3306
  name: 'client{config.lab_number}'
  username: 'client{config.lab_number}'
  password: '{config.client_database_password}'

client:
  id: '{config.client_id}'
  name: 'client-{config.lab_number}'
  secret: '{config.client_secret}'
  scopes:
    - 'read:profile'
  uri: 'http://127.0.0.1:3001'
  logo_uri: 'http://127.0.0.1:3001/static/img/logo.png'
  redirect_uri: 'http://127.0.0.1:3001/callback'

authorization_server:
  issuer: 'http://127.0.0.1:3000'
  authorize_uri: 'http://127.0.0.1:3000/oauth/authorize'
  token_uri: 'http://127.0.0.1:3000/oauth/token'
  jwk_uri: 'http://127.0.0.1:3000/.well-known/jwks.json'
  revocation_uri: 'http://127.0.0.1:3000/oauth/revoke'
  register_uri: 'http://127.0.0.1:3000/oauth/register'
  registration_secret: '{config.registration_secret}'

resource_server:
  base_url: 'http://127.0.0.1:3000'

cookie:
  secret: '{config.client_cookie_secret}'
  secure: false
  path: '/'
  domain: '127.0.0.1'

redis:
  host: '127.0.0.1'
  port: 6379
  database: 0
"""
    client_config_file = os.path.join(lab_dir, "client", "config.yaml")
    with open(client_config_file, "w") as fout:
        fout.write(client_config)


def copy_lab_dir(base: str, number: str):
    base_number = LAB_DIR_RE.search(base).group(1)
    lab_dir = os.path.join(BASE_DIR, f"lab{number}")
    shutil.copytree(base, lab_dir)

    for component in ("server", "client"):
        component_dir = os.path.join(lab_dir, component)
        constants_file = os.path.join(
            component_dir, "internal", "constants", "constants.go"
        )
        with open(constants_file) as fout:
            raw_data = fout.read()
            raw_data = raw_data.replace(f'"{base_number}"', f'"{number}"')
        with open(constants_file, "w") as fout:
            fout.write(raw_data)

        for root, _, files in os.walk(component_dir):
            for f in files:
                source_file = os.path.join(root, f)
                if f == "go.mod" or f == "config.yaml":
                    pass
                elif f.endswith(".go") is False:
                    continue

                with open(source_file) as fin:
                    source = fin.read()

                source = source.replace(f"client-{base_number}", f"client-{number}")
                source = source.replace(f"client{base_number}", f"client{number}")
                source = source.replace(f"server-{base_number}", f"server-{number}")
                source = source.replace(f"server{base_number}", f"server{number}")
                source = source.replace(f"lab{base_number}", f"lab{number}")
                with open(source_file, "w") as fout:
                    fout.write(source)


def main(args):
    number = str(args.number).zfill(2)
    lab_dir = os.path.join(BASE_DIR, f"lab{number}")
    if os.path.isdir(lab_dir):
        print(f"Error: lab {number} already exists.")
        exit(1)

    config = Config(
        lab_number=number,
        registration_secret=secrets.token_hex(32),
        database_host="db",
        client_id=str(uuid.uuid4()),
        client_secret=new_password(32),
        client_cookie_secret=secrets.token_hex(32),
        client_database_password=new_password(32),
        server_cookie_secret=secrets.token_hex(32),
        server_encryption_key=secrets.token_hex(32),
        server_database_password=new_password(32),
        private_key=new_rsa_private_key(),
    )
    dev_config = Config(
        lab_number=number,
        registration_secret=secrets.token_hex(32),
        database_host="127.0.0.1",
        client_id=config.client_id,
        client_secret=new_password(32),
        client_cookie_secret=secrets.token_hex(32),
        client_database_password=new_password(32),
        server_cookie_secret=secrets.token_hex(32),
        server_encryption_key=secrets.token_hex(32),
        server_database_password=new_password(32),
        private_key=new_rsa_private_key(),
    )
    copy_lab_dir(args.base, number)
    append_to_compose(number)
    append_to_caddy(number)
    append_to_sql_init(config, dev_config)
    write_dev_config_files(dev_config)


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("-n", "--number", type=int, required=True)
    ap.add_argument("-b", "--base", type=str, default="lab00")
    args = ap.parse_args()
    if not LAB_DIR_RE.search(args.base):
        print(f"Error: invalid base lab")
        exit(1)
    if not os.path.isdir(os.path.join(BASE_DIR, args.base)):
        print(f"Error: unable to use {args.base} as base lab, directory not found.")
        exit(1)

    main(args)
