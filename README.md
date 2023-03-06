# bettor 🎲

[![CI](https://github.com/elh/bettor/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/elh/bettor/actions/workflows/ci.yaml) [![CD](https://github.com/elh/bettor/actions/workflows/cd.yaml/badge.svg?branch=main)](https://github.com/elh/bettor/actions/workflows/cd.yaml)

Bettor is a [parimutuel betting (or pool betting)](https://en.wikipedia.org/wiki/Parimutuel_betting) Discord bot and API for running your own virtual betting books. Inspired by [Twitch Predictions](https://help.twitch.tv/s/article/channel-points-predictions?language=en_US).

Make a fun game out of predictions, document friendly disagreements, and maybe figure out who's the best at seeing the future 🔮🙀! Create bets with your friends on whatever you want right there in the context of a conversation on Discord. Bettor will keep track of everyone's bets and handle pay outs.

<p align="center">
    <img width="60%" alt="Bettor Discord commands" src="https://user-images.githubusercontent.com/1035393/222991983-d5716eea-ccc9-442a-9b7a-d74870161268.png">
    <img width="60%" alt="Creating a bet" src="https://user-images.githubusercontent.com/1035393/222992460-b65a27d9-fb2d-49e8-828a-b9a09b8aa265.png">
    <img width="60%" alt="Locking a bet" src="https://user-images.githubusercontent.com/1035393/222991988-9fa5d4ac-8753-4718-bacc-e6c5caac29aa.png">
</p>

## Discord Bot

Bettor provides a bot that can be invited to Discord servers (or "guilds"). It provides commands for creating and managing betting pools like `/start-bet`, `/join-bet`, and `/bettor`. It creates a betting "book" for the Discord server which isolates that server's bets, markets, and users from other servers. Upon first interaction with the bot, users in the server are initialized with a default value of betting points.

The Discord bot must be added to servers with `applications.commands` scopes granted.

## API

Bettor provides a Buf Connect, gRPC-compatible API and server. Web clients can hit this via gRPC-Web generated clients.

Resources are partitioned by betting book.

See [docs/](https://github.com/elh/bettor/blob/main/docs/index.html) for API documentation.

> **Note**
> As of 1/23/23 usage, data is persisted using gobs in files. This should be replaced with a proper database eventually.

> **Warning**
> As of 1/23/23 usage, authn/z and perms are not implemented because access to the API is only possible by the Discord bot which piggybacks on Discord's user identity to restrict requests.

## Using Bettor

At the moment, I am not running a public Bettor Discord bot. I built this in a quick sprint as a fun toy for my own Discords and to play with some Go tools.

If you want to run your own Bettor instance, you can adapt or run this project with your own Discord credentials.
* To deploy to Fly.io:
    * Create your own app with Fly.io and update `app` in `fly.toml`.
    * Create a mounted volume for persistence.
    * Create a Discord Bot and provide `discordToken` as an environment variable using Fly.io secrets.
* Deploying elsewhere should be easy. Just make sure that `discordToken` is provided as a flag or environment variable.

I do not expect to work on this anymore or host a public version, but who knows. I'd love to hear from you if you found this cool or have feedback or requests. 😇

## Development

APIs are defined using Protocol buffers built with Buf. Server is implemented using Buf Connect.

Integration with Discord API is provided by `bwmarrin/discordgo`. Tracing and logging are provided using OpenTelemetry and go-kit/log.

My Bettor instance is deployed to Fly.io via Docker. CI/CD is provided by Github Actions.

Dependencies are managed using Nix. Optionally if `direnv` is installed, a Nix shell will be started automatically when entering the root directory.

### Running Locally

Install dependencies using Nix. `default.nix` defines all development dependencies for the project.

If you want to run the Discord bot, you will need to enable it and provide a Discord bot token as flags in the server main file. For dev testing, you will want to use a different Discord bot than the one used in production, joined to your test Discord servers.

See flags: `-runDiscord`, `-discordToken`, `-cleanUpDiscordCommands`

Go flags can be provided via environment variables and `.env` files using `joho/godotenv`.

```
$ make run-local-server
$ make run-local-bot
```

### Workflow

See development workflow commands in `Makefile`.
```
$ make gen
$ make test
$ make lint
$ make docker
```
