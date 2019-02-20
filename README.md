# Artifact

## Levels

### Job

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
- `Nd` for N days
- `Nw` for N weeks
- `Nm` for N months
- `Ny` for N years

If epires flag is not set artifact never expires.

##### Output

##### Requirements
- SEMAPHORE_JOB_ID
- Linux, macOS: `~/.artifact/credentials`
- Windows: `dir "%UserProfile%\.artifact\credentials"`

