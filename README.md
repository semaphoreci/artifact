# Artifact

## Table of contents

- [Use-cases](#use-cases)
- [Concepts](#concepts)
- [CLI](#cli)
  - [push](#push)
  - [pull](#pull)
  - [yank](#yank)

## Use-cases

1. Arhiving artifacts - storing final deliverables
2. Promoting artifacts though pipeline and workflow
3. Debugging - Store and inspect logs, screenshots and other debug data through UI or CLI

## Concepts

### Big picture

Artifacts can be stored and accessed on three different layers. This concept makes it easy to use artifacts for different purposes with very simple commands.

Artifact stores:

- **Project** - static, one per project
- **Workflow** - dynamic, nasted under current workflow
- **Job** - dynamic, nasted under current job


### Project level

**Use-case** - Storing final deliverables

This is static level. It's an artifact store that is per project and it's possible to simply upload/download files and directories from this level from every job from any workflow and pipeline. Here is an example how you can interect with project level store from any job in your pipeline or from your local development environemnt.

From any jobs running on Semaphore:

```sh
artifact push project myapp-v3.tar
artifact pull project myapp-v3.tar
```

From your development environment:

```sh
sem artifact push project --project payment-api myapp-v3.tar
sem artifact pull project -p payment-api myapp-v3.tar
```

You can also view and download your artifacts from project page in the UI.


### Workflow level

**Use-case** - Promoting build artifacts accross pipelines and block levels. e.g. promoting from _Build and test_ pipeline into _Production deployment pipeline_.

This is dynamic level. New store is created for each new workflow. Here are examples for interacting with this store from any pipeline and any job within workflow.

From any jobs running on Semaphore:

```sh
artifact push workflow myapp-v3.tar
artifact pull workflow myapp-v3.tar
```

From your development environment:

```sh
sem artifact push workflow --workflow <WORKFLOW_ID> myapp-v3.tar
sem artifact pull workflow -w <WORKFLOW_ID> myapp-v3.tar
```

You can also view and download artifacts from workflow page in the UI.

### Job level

**Use-case** - Debugging jobs with easy access to artifacts that job created. e.g. Storing logs, screenshots, core dumps and inspecting them them on the job page.

This is dynamic level. New store is created for each new job. Here are examples for interacting with this store from job.

From job running on Semaphore:

```sh
artifact push job logs/build.log
artifact pull job logs/build.log
```

From your development environment:

```sh
sem artifact push job --job <JOB_ID> logs/build.log
sem artifact pull job -p <JOB_ID> logs/build.log
```

You can also view and download artifacts on the job page in the UI.

## CLI

## push

#### `artifact push job x.zip`

##### Description

Uploads file or path into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip`.

###### Example 1: Uploading nested file.

`artifact push job logs/webserver/access.log` pushs file into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/log/webserver/access.log`

###### Example 2: Uploading directory

`artifact push job logs/webserver` pushs directory with all sub directories and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/log/webserver`

##### Alternative forms and flags

1. `--destination` or `-d` sets destination directory or file path

`artifact push job x.zip -d y.zip` pushs file into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/y.zip`.

`artifact push job logs/webserver --destination debuglogs` pushs all sub-dirs and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/debuglogs`.

2. `--job <job-id>` or `-j <job-id>`

By default command is looking for `SEMAPHORE_JOB_ID` env var. If it's not available it fails. If flag `--job` is specified it takes precedence over `SEMAPHORE_JOB_ID`.

3. `--expire-in 10d` or `-e 10d`

Expires - deletes the files or directories after amount of time specified.

Supported options are:
- just integer (number of seconds)
- `Nh` for N hours
- `Nd` for N days
- `Nw` for N weeks
- `Nm` for N months
- `Ny` for N years

If expires flag is not set artifact never expires.

##### Output

TODO

##### Requirements
- SEMAPHORE_JOB_ID (not required if `--job` flag is specified)
- Linux, macOS: `~/.artifact/credentials`
- Windows: `dir "%UserProfile%\.artifact\credentials"`

### Putting artifacts into artifact store on different levels

Other supported levels include `workflow` and `project` level. These are variations of the command depending on the level:

#### `artifact push workflow x.zip`

File is stored into `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip`

#### `artifact push project x.zip`

File is stored into `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip`

## pull

#### `artifact pull job x.zip`

##### Description

Artifact stored at `/artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip` will be push at current directory as `x.zip`.

##### Alternative forms and flags

. `--job <job-id>` or `-j <job-id>`

By default command is looking for `SEMAPHORE_JOB_ID` env var. If it's not available it fails. If flag `--job` is specified it takes precedence over `SEMAPHORE_JOB_ID`.

##### Requirements
- SEMAPHORE_JOB_ID (not required if `--job` flag is specified)
- Linux, macOS: `~/.artifact/credentials`
- Windows: `dir "%UserProfile%\.artifact\credentials"`

### Putting artifacts into artifact store on different levels

Other supported levels include `workflow` and `project` level. These are variations of the command depending on the level:

#### `artifact pull workflow x.zip`

File is stored into `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip` would be restored at current directory as `x.zip`.

#### `artifact pull projects x.zip`

File is stored into `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip` would be restored at current directory as `x.zip`.

## yank

#### `artifact yank`

##### Description

Deletes artifact.

`artifact yank job x.zip` deletes `/artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip`

`artifact yank workflow x.zip` deletes `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip`

`artifact yank project x.zip` deletes `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip`

## list

#### `artifact list`

##### Description

`artifact list job` lists root of the job directory `/artifacts/jobs/<SEMAPHORE_JOB_ID>/`

`artifact list workflow` lists root of the job directory `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/`

`artifact list project` lists root of the job directory `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/`
