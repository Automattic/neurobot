# neurobot - A Matrix Workflow Builder

| | |
|----|----|
| <img src="https://github.com/Automattic/neurobot/blob/master/resources/icon.svg?raw=true" alt="neurobot brain emoji" width="70" /> | It is an engine in which you can define workflows to be triggered by certain events leading to execution of a set of instructions defined under that workflow. |

![neurobot's architecture](https://github.com/Automattic/neurobot/blob/master/resources/visual.png?raw=true)

> [Explanation of architecture](resources/docs/architecture.md)

Currently supported event triggers:

| Trigger | Variety |
| ------- | ------- |
| External webhook request with payload `?message=X` | `webhook` |

Currently supported workflow step:

| Workflow Step | Variety |
| ------------- | ------- |
| Show message on stdout | `stdout` |
| Post message to a Matrix room | `postMatrixMessage` |

## How to run neurobot?

### Components

List of concerned files:
- Compiled program (binary)
- `.env` - used for configuration
- `neurobot.db` - SQLite database file
- `resources/workflows.toml` - used for defining workflows using [TOML syntax](https://toml.io/en/)

You can compile the program by `make build`, which will generate the `neurobot` binary in the project root. Then just start the program, by specifying what `.env` file to load. By default it looks for it in the current directory. A sample `.env.sample` file is also provided for use. All configuration sits inside of `.env` file. When starting up, for the first time, a SQLite database would be created and with every run, workflows defined in TOML file are imported, overwriting previous imported data of the defined workflows. TOML file will eventually be replaced by a UI, but that's not on the short-term roadmap. Refer to [TOML file structure](toml-structure.md) to make sense of it.

### Matrix bot

You would need to create a bot user (a user that's meant to be programmatically controlled is a bot, there is no other difference between a regular user and bot user) on your Matrix homeserver and supply its access token in the `.env` file. You don't have to name it `neurobot` but for documentation, that's the name we will assume, you have chosen. If your workflows would require matrix actions that require admin priveleges, you can promote `neurobot` to be an admin on the server as well. For a deep understanding, we suggest reading more on [neurobot's Architecture](resources/docs/architecture.md).

### Adding your own workflow

Add workflows in your `workflows.toml` file. [Understand TOML file structure](resources/docs/toml-structure.md)

## Credits

Thanks to [OpenMoji](https://openmoji.org) for open source emojis!
