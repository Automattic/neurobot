Architecture
============

neurobot's architecture is built to react on the basis of events, not just events within Matrix but outside of Matrix as well. This provides a great foundation to describe any kind of integration that we can possibly think of.

[💥 Trigger] --------> [ 🚀 Workflow = [📡 WorkflowStep] + [📤 WorkflowStep] ]

## Components

The files you need to concern yourself with are:
- Compiled program (binary)
- `.env` - used for configuration
- `workflows.toml` - used for defining workflows using [TOML syntax](https://toml.io/en/)
- `wfb.db` - SQLite database file

You can compile the program by `make build` and binary file gets saved in `bin` directory. Then just start the program, with the possiblity of specifying what `.env` file to load. By default it looks for it in the current directory. All configuration sits inside of `.env` file. When starting up, for the first time, a SQLite database would be created and with every run, workflows defined in TOML file are imported, overwriting previous imported data of the defined workflows. TOML file will eventually be replaced by a UI, but that's not on the short-term roadmap. Refer to [TOML file structure](toml-structure.md) to make sense of it.

## How does it work?

You need to create a bot user (a user that's meant to be programmatically controlled is a bot, there is no other difference between a regular user and bot user) on your Matrix homeserver and supply its access token in the `.env` file. You don't have to name it `neurobot` but for documentation, that's the name we will assume, you have chosen. If your workflows would require matrix actions that require admin priveleges, you can promote `neurobot` to be an admin on the server.

If you need to post message as a different bot, meaning a different name and picture, you would have to create a new bot user, and supply its credentials (currently only possible to do by directly entering into the `bots` database table). This is an intentional design choice, so that anyone with hosted homeservers can also setup workflows/integrations by just adding more bot users. You would get to choose which bot user to use, in the relevant workflow step. Make sure that bot has been invited to the room, in which its supposed to post a message.

Upon startup, engine would login as all bots individually and maintain a pool of matrix client instances and starts the `sync` process with the homeserver, giving each bot the chance of reacting to events as they come in. It also loads the triggers, workflows and workflow steps that are defined in the database. Do note that TOML file is only parsed once & imported at startup and then everything happens based on the data inside the database. Its only when the program starts again, that TOML file is reimported. In future, we would implement signalling the program to reload TOML file without requiring a reload of the main program itself.

When triggers are loaded, it starts the monitoring process of defined triggers. For `webhook` variety of triggers, we start a single webhooks listener server, which handles all incoming HTTP requests from outside services. All endpoints share a common prefix `webhooks-listener`. For `poller` variety of triggers, it invokes setup mechanism of these triggers, based on which they can keep polling. This isn't well-built yet, just the skeleton of the mechanism exist.

More variety of triggers are planned such as:
- Matrix based events (commands invoked, emoji reactions etc)
- Schedule based (Cron)

When workflow steps are loaded, they are just queued up in their specified order within a particular workflow and await start of the workflow. When a workflow starts, it may or may not have a payload to pass to the first workflow step. Every workflow step would accept the payload from the previous workflow step and passes it forward, with any modification it chooses to make to it.

Each trigger and workflow step carries additional meta information based on their variety.

## What other varieties of workflow steps are planned?

Hard to put an exhaustive list, as we would build what we need first. Some are:

- Ping an external endpoint with payload data
- Query API to add more data to payload data
- DM a certain user
- Ask questions to a group of users in a DM and aggregate those answers & post in a matrix room. `Stand up meetings`
- Send email

## What else is planned for the future?

### Keeping tabs on who's online

We have a polyglots command in Automattic, which when invoked can help you find someone who speaks a certain language and is online right now. Supporting such a command would require knowing who is online and this is probably best done by `neurobot` by maintaining a "online users" list, which can simply be utilised by a workflow step.
