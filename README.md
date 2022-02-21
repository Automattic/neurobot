# neurobot - A Matrix Workflow Builder

It is an engine in which you can define workflows to be triggered by certain events leading to execution of a set of instructions defined under that workflow.

Currently supported event triggers:

| Trigger | Variety |
| ------- | ------- |
| External webhook request with payload `?message=X` | `webhook` |

Currently supported workflow step:

| Workflow Step | Variety |
| ------------- | ------- |
| Show message on stdout | `stdout` |
| Post message to a Matrix room | `postMatrixMessage` |

## Instructions to run it

### Quick Demo

Copy `.env.sample` file as `.env` file and run the program using this command:

`go run main.go`

First run, inserts some data into the sqlite database `wfb.db` which will enable you to send it a HTTP request with payload `?message=hello` which will trigger the workflow that makes it log to `stdout`.

After running the program, you can send a HTTP request using CURL like this:

`curl localhost:8080/webhooks-listener/quickstart?message=Hello`

You should see these lines in output:

```
Request received on webhook listener! /webhooks-listener/quickstart
suffix: quickstart registered: true

Running workflow #1 payload:{Hello }

>>Hello
```

### Adding your own workflow

Currently, its too early in the experimentation phase to build a UI to add database records. So you have to add them manually in the SQLite database.

In the `workflows` table, add a new row & take note of the workflow id. In `triggers` table, add a new row and specify workflow id under `workflow_ids` column, which is meant to be a CSV. Now, under `workflow_steps` table, add a new row and specify workflow id under `workflow_id` column.

Certain triggers and certain workflow steps require additional info which are to be added in their respective meta tables: `trigger_meta` and `workflow_step_meta`.

### Matrix HomeServer

To run `neurobot` with Matrix homeserver, you can specify credentials in the `.env` file and run again by same command:

`go run main.go`

### Debug mode

Add debug flag:

`go run main.go -debug=true`

## Architecture

![neurobot's architecture](https://github.com/Automattic/neurobot/blob/master/neurobot-visual.png?raw=true)

Engine is built to react on the basis of events. Workflows are defined as an ordered list of workflow steps that are to be executed when the workflow is started. And workflows' execution start when the chosen event for its execution is triggered.

For example: An incoming webhook can trigger a workflow, which can contain workflow step(s) of like posting a message to Matrix room. Or a new item in RSS Feed triggers a workflow, which can execute steps like posting a message to Matrix room and sending an external webhook request.

Right now, there is no UI to define triggers, workflows & workflow steps but most likely it will be built as an interaction with the bot user that's meant to be used for this purpose.
