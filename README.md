# nerdkord

nerdkord is a bot Discord for nerds.
It can be installed on server *and* user-side.
Of course, nerdkord is easily self-hostable!

## Features

- $\LaTeX$ render
- Math calculation
- Converting math input into $\LaTeX$ code

## Install

You can build it and deploy it with Docker.
Download the `Dockerfile`, the `docker-compose.yml` and the `.env.example`.
Rename `.env.example` into `.env` and add your token inside.
Now, you can start the bot.

It is also possible to compile it with Go 1.24+.
You must have texlive (packages `texlive texlive-binextra texlive-dvi xdvik texmf-dist-full` for Alpine Linux) installed
to run the bot.
Currently, you must pass your token in the argument `-token`, e.g.
```bash
$ ./bot -token your_token
```
Later we will support environment variable to load the token.

## Configuration

The config file is located at `config/config.toml`.
The default file:
```toml
debug = false
author = 'nyttikord'
use_postgres_instead_of_sqlite = false

[sqlite]
path = 'nerdkord.db'

[postgres]
host = 'localhost'
user = 'nerdkord'
password = 'password'
db_name = 'nerdkord'
port = 5432
time_zone = 'Europe/Paris'

```
- `debug` is true if the bot is in debug mode (:warning: does not support user-side installs if this is true!)
- `author` is the host of the bot
- `use_postgres_instead_of_sqlite` is true if the bot must use PostreSQL instead of SQLite3
- `[sqlite].path` is a path to the SQLite3 file
- `[postgres].host` is the host of PostgreSQL
- `[postgres].user` is the user to use
- `[postgres].password` is the user's password
- `[postgres].db_name` is the name of the DB to use
- `[postgres].port` is the port of PostreSQL
- `[postgres].time_zone` is the timezone to use with PostgreSQL

## Technologies

- Go
- [anhgelus/gokord](https://github.com/anhgelus/gokord) for interacting with Discord API and the database
- [nyttikord/gomath](https://github.com/nyttikord/gomath) for parsing and evaluating math expression