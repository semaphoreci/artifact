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

- **Project** - one per project
- **Workflow** - nasted under current workflow
- **Job** - nasted under current job


### Project level

**Use-case** - Storing final deliverables

It's an artifact store that is per project and it's possible to simply upload/download files and directories from this level from every job from any workflow and pipeline. Here is an example how you can interect with project level store from any job in your pipeline or from your local development environemnt.

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

New store is created for each new workflow. Here are examples for interacting with this store from any pipeline and any job within workflow.

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

New store is created for each new job. Here are examples for interacting with this store from job.

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

`artifact push job myapp/v1/logs/webserver` pushs directory with all sub directories and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/webserver`

`artifact push job /var/semaphore/webserver/logs/testxyz` pushs directory with all sub directories and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/testxyz`

__We are relying on semantics of UNIX__

`cp -r /var/semaphore/webserver/logs/testxyz /tmp`

`cp /var/semaphore/webserver/logs/testxyz /tmp`

`cp -r webserver/logs/testxyz /tmp`

End result for both examples above: `/tmp/testxyz`

##### Alternative forms and flags

1. `--destination` or `-d` sets destination directory or file path

`artifact push job x.zip -d y.zip` pushs file into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/y.zip`.

Example for directory: `artifact push job logs/webserver --destination debuglogs` pushs all sub-dirs and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/debuglogs`.

Example for deeply nested directory as destination: `artifact push job logs/webserver --destination path/to/debuglogs` pushs all sub-dirs and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/path/to/debuglogs`.

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

4. `--force` or `-f`

`artifact push job x.zip` if `x.zip` exists in the bucket this command should fail. To overwrite file or directory user would need to specify "force" flag.

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

Artifact stored at `/artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip` will be downloaded in current directory as `x.zip`.

Example with directory: `artifact pull job logs`, if logs is directory `logs` it will be created locally in current directory and whole content of `logs` from bucket will be downloaded into `logs` directory locally.

##### Alternative forms and flags

1. `--destination` or `-d` sets destination directory or file path

`artifact pull job x.zip -d z.zip` pushs file into `z.zip`.

Example for directory: `artifact pull job logs --destination debuglogs` pulls all sub-dirs and files into `debuglogs` locally.

Example for deeply nested directory as destination: `artifact pull job logs --destination path/to/debuglogs` pulls all sub-dirs and files into `path/to/debuglogs` in current local directory.

2. `--job <job-id>` or `-j <job-id>`

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

Example for directory: `artifact yank job logs` deletes `/artifacts/jobs/<SEMAPHORE_JOB_ID>/logs` directory and all recursively all the content that is in the `logs` directory in the bucket.

`artifact yank workflow x.zip` deletes `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip`

`artifact yank project x.zip` deletes `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip`

## list

#### `artifact list`

##### Description

`artifact list job` lists root of the job directory `/artifacts/jobs/<SEMAPHORE_JOB_ID>/`

`artifact list workflow` lists root of the job directory `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/`

`artifact list project` lists root of the job directory `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/`
