# bettor 🎲

[![CI](https://github.com/elh/bettor/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/elh/bettor/actions/workflows/ci.yaml) [![CD](https://github.com/elh/bettor/actions/workflows/cd.yaml/badge.svg?branch=main)](https://github.com/elh/bettor/actions/workflows/cd.yaml)

Bettor is a parimutuel (or "pool") betting Discord Bot and API for running your own virtual betting books. Inspired by [Twitch Predictions](https://help.twitch.tv/s/article/channel-points-predictions?language=en_US).

## Discord Bot

Bettor runs a Discord bot that can be invited to servers. It provides commands for creating and managing betting pools like `/start-bet`, `/join-bet`, and `bettor`. It creates a betting "book" for the Discord server (or "guild") which isolates that server's bets, markets, and users from other servers. Upon first interaction with the bot, users in the server are initialized with some default value of points.

Discord bot must be added with `applications.commands` scopes.

## API

Bettor offers a Buf Connect, gRPC-compatible API. Web clients can hit this via gRPC-Web generated clients.

Resources are partitioned by betting "book".

NOTE: As of 1/23/23, Data is persisted using gobs. This should be replaced with a proper database eventually.

NOTE: As of 1/23/23, Authn/z and perms are not implemented because access to the API is only possible by the Discord bot which piggybacks on Discord's user identity to correctly restrict requests.

See [docs/](https://github.com/elh/bettor/blob/main/docs/index.html) for API documentation.

## Build

Dependencies are managed using Nix. Optionally if `direnv` is installed, a Nix shell will be started automatically when entering the root directory.

APIs are defined using Protocol buffers built with Buf. Server is implemented using Buf Connect.

Integration with Discord API is provided by `bwmarrin/discordgo`.

Bettor is deployed to Fly.io via Docker.

CI/CD is provided by Github Actions.

Tracing and logging are provided using OpenTelemetry and go-kit/log.

### Running Locally

Install dependencies using Nix. `default.nix` defines all development dependencies for the project.

If you want to run the Discord bot, you will need to enable it and provide a Discord bot token as flags in the server main file. For dev testing, you will want to use a different Discord bot than the one used in production, joined to your test Discord servers.

Go flags can be provided via environment variables and `.env` files using `joho/godotenv`.

```
$ make run-local-server
$ make run-local-bot
```

### Development

See development workflow commands in `Makefile`.
```
$ make gen
$ make test
$ make lint
$ make docker
```
