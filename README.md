# Matrix Workflow Builder

It is an engine in which you can define workflows to be triggered by certain events leading to execution of a set of instructions defined under that workflow.

Currently supported event triggers:

- External webhook request with payload `?message=X`

Currently supported workflow step:

- Show message on `stdout`
- Post message to a Matrix room

## Instructions to run it

### Quick Demo

For a quick demonstration, just run the program using this command:

`go run main.go -debug=true`

First run, inserts some data into the sqlite database `wfb.db` which will enable you to send it a HTTP request with payload `?message=hello` which will trigger the workflow that makes it log to `stdout`.

After running the program, you can send a HTTP request using CURL like this:

`curl localhost:8080/webhooks-listener/quick?message=Hello`

### Matrix HomeServer

To run the workflow builder with Matrix homeserver, you can run it using this command, and have to supply Matrix credentials for the bot account you intend to use for posting messages in the `.env` file:

`go run main.go -homeserver="http://localhost:8008" -username="morpheus" -password="redpill" -debug=true`

## Architecture

![matrix workflow builder architecture](https://github.com/Automattic/matrix-workflow-builder/blob/master/matrix-workflow-builder-visual.png?raw=true)

Engine is built to react on the basis of events. Workflows are defined as an ordered list of workflow steps that are to be executed when the workflow is started. And workflows' execution start when the chosen event for its execution is triggered.

For example: An incoming webhook can trigger a workflow, which can contain workflow step(s) of like posting a message to Matrix room. Or a new item in RSS Feed triggers a workflow, which can execute steps like posting a message to Matrix room and sending an external webhook request.

Right now, there is no UI to define triggers, workflows & workflow steps but most likely it will be built as an interaction with the bot user that's meant to be used for this purpose.
