# Artifact

## Table of contents

- [Use-cases](#use-cases)
- [Concepts](#concepts)
- [xxx](#xxx)
- [CLI](#cli)
  - [put](#put)
  - [get](#get)
  - [list](#list)
  - [delete](#delete)
  - [exists](#exists)


## Use-cases

1. Arhiving artifacts - by default artifacts are kept forever however you can specify expiration policy or delete them manually
2. Promoting artifacts though pipeline and workflow
3. Debugging - Store and inspect logs, screenshots and other debug data through UI or CLI

## Concepts

Artifacts can be stored on various hierarchical lavels of your CI/CD process.

### Job level artifacts

When uploading artifacts this is deafult level. Artifacts stored on this level can be inspected through the UI by joing to a job page in the UI.

Upload artifact in job with `artifact` CLI:

`artifact put yourfile.json`

To view artifacts uploaded in a job simple visit job page: https://ORG.semaphoreci.com/jobs/JOB_ID

To download artifact from your local development environment use:

`sem artifact get --job JOB_ID yourfile.json`

To download all files generated by this job use

`sem artifact get --all --job JOB_ID`

It is also possible to download artifacts stored on job level from another job within pipeline however it's impractocal. For this use-case of promoting artifacts through pipeline and workflow you should store artifacts of those levels. For the sake of completness command to do that would be:

`artifact get --job JOB_ID yourfile.json`

Implementation notes:
- all files and directories uploaded from the same job are stored in the same base directory which is `/artifacts/jobs/JOB_ID/`
- artifact store is backed by a S3 bucket or compatible store.
- Uploading same file or directory that already exists will fail. Use `--force` or `-f` flag to overwrite it.

### Pipeline level artifacts

Storing artifacts on the pipeline level is best used when you need to work with artifacts produced within one block in other blocks within same pipeline, but you don't need those artifacts in other pipelines that you can have within same workflow.


## CLI

## put

#### `artifact put x.zip`

##### Description

Uploads file or path into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip`.

###### Example 1: Uploading nested file.

`artifact put logs/webserver/access.log` puts file into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/log/webserver/access.log`

###### Example 2: Uploading directory

`artifact put logs/webserver` puts directory with all sub directories and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/log/webserver`

##### Alternative forms and flags

1. `artifact put job x.zip` is alias for `artifact put x.zip`

2. `--destination` or `-d` sets destination directory or file path

`artifact put job x.zip -d y.zip` puts file into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/y.zip`.

`artifact put job logs/webserver --destination debuglogs` puts all sub-dirs and files into `/artifacts/jobs/<SEMAPHORE_JOB_ID>/debuglogs`.

3. `--job <job-id>` or `-j <job-id>`

By default command is looking for `SEMAPHORE_JOB_ID` env var. If it's not available it fails. If flag `--job` is specified it takes precedence over `SEMAPHORE_JOB_ID`.

4. `--expire-in 10d` or `-e 10d`

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

Other supported levels include `pipeline`, `workflow` and `project` level. These are variations of the command depending on the level:

#### `artifact put pipeline x.zip`

File is stored into `/artifacts/pipelines/<SEMAPHORE_PIPELINE_ID>/x.zip`

#### `artifact put workflow x.zip`

File is stored into `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip`

#### `artifact put project x.zip`

File is stored into `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip`

## get

#### `artifact get x.zip`

##### Description

Artifact stored at `/artifacts/jobs/<SEMAPHORE_JOB_ID>/x.zip` will be put at current directory as `x.zip`.

##### Alternative forms and flags

1. `artifact get job x.zip` is alias for `artifact get x.zip`

2. `--job <job-id>` or `-j <job-id>`

By default command is looking for `SEMAPHORE_JOB_ID` env var. If it's not available it fails. If flag `--job` is specified it takes precedence over `SEMAPHORE_JOB_ID`.

##### Requirements
- SEMAPHORE_JOB_ID (not required if `--job` flag is specified)
- Linux, macOS: `~/.artifact/credentials`
- Windows: `dir "%UserProfile%\.artifact\credentials"`

### Putting artifacts into artifact store on different levels

Other supported levels include `pipeline`, `workflow` and `project` level. These are variations of the command depending on the level:

#### `artifact get pipeline x.zip`

File is stored at `/artifacts/pipelines/<SEMAPHORE_PIPELINE_ID>/x.zip` would be restored at current directory as `x.zip`.

#### `artifact get workflow x.zip`

File is stored into `/artifacts/workflows/<SEMAPHORE_WORKFLOW_ID>/x.zip` would be restored at current directory as `x.zip`.

#### `artifact get projects x.zip`

File is stored into `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/x.zip` would be restored at current directory as `x.zip`.

## get

#### `artifact list`

##### Description

`artifact list` lists root of the job directory `/artifacts/jobs/<SEMAPHORE_JOB_ID>/`

`artifact list job` lists root of the job directory `/artifacts/jobs/<SEMAPHORE_JOB_ID>/`

`artifact list pipeline` lists root of the job directory `/artifacts/pipelines/<SEMAPHORE_PIPELINE_ID>/`

`artifact list workflow` lists root of the job directory `/artifacts/workflows/<SEMAPHORE_PIPELINE_ID>/`

`artifact list project` lists root of the job directory `/artifacts/projects/<SEMAPHORE_PROJECT_ID>/`
