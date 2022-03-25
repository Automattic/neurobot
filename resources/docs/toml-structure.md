TOML Structure
==============

We utilise [TOML config file](https://toml.io/en/) to define workflows. Highly recommended to get yourself familiar with it.

To define a workflow, you need to define 3 parts:
- Workflow details
- Trigger
- Workflow steps

Workflows are defined in an array `[[workflow]]` with some basic details and a special `identifier` field. Next `[workflow.trigger]` are defined with some basic details and meta fields, which varies with each variety of the trigger. Last comes `[[workflow.step]]` which is itself an array within a single workflow, again defined with some basic details and meta fields, which varies with each variety of the workflow step.

Refer to `workflows-sample.toml` to see a few examples.

## `identifier` field

This special field is what helps in identifing a particular workflow. If this stays same, and all fields are updated, the engine would figure out what workflow to update in the database on subsequent runs. So, this field should never be modified. Internally, its saved as a workflow meta field.

## Disable a workflow

Simply change the value of `active` to `false`

## Meta fields

### Triggers

#### `webhook` trigger

##### `urlSuffix`

This would be the url suffix in webhooks listening endpoint for your trigger. `https://example.com/{urlSuffix}`

### Workflow Steps

#### `postMatrixMessage` workflow step

##### `messagePrefix`

This would be added as a prefix to every message that is to be posted.

##### `matrixRoom`

Default matrix room to post the message to, when not specified in payload.

##### `asBot`

What bot user to use to post the message as. `neurobot` bot user is used when not specified.
