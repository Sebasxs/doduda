<p align="center">
  <h3 align="center">DODUDA</h3>
  <p align="center">Dofus 3 unpacker via terminal</p>
</p>

> [!NOTE]
> This is a modified version of the original **doduda** tool. All the core functionality belongs to [dofusdude](https://github.com/dofusdude)<br>
> Modifications include downloading more data and simplified flags.

## Installation

```bash
git clone https://github.com/sebasxs/doduda
cd doduda
go build
./doduda # Run the program
```

## Features

See `doduda --help` for all parameters.

This version of `doduda` simplifies the download process into three main categories:

-  **Data:** Includes core game data like items, quests, monsters, etc.
-  **Images:** All game pictos including items, monsters (low-res), ui, etc.
   -  Images with multiple resolutions are downloaded at the highest resolution by default.
   -  Duplicate images resulting from sprite-texture2D parity during unpacking are correctly filtered and organized into appropriate folders.
-  **Languages:** i18n files for different languages.

> [!NOTE]
> Now all the images and core data are downloaded and not only items, mounts and quests.

### Watchdog

The `doduda listen` command enables a watchdog that will listen for new Dofus versions and react to their release.
This allows you to automate tasks when a new version is released.

<img src="https://vhs.charm.sh/vhs-g7BGgJ5f4iUhuzRhoYzzR.gif" alt="watchdog example" width="600">

You can use that for getting anything that supports webhooks to react to Dofus version updates. Some ideas are:

-  [Discord Channels](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
-  [ntfy.sh](https://ntfy.sh) (Push notifications to your phone)

## Usage

The core functionality of `doduda` remains the same.

### New flags

-  `--ignore data`: Skips downloading core game data.
-  `--ignore images`: Skips downloading all game pictos.

### Removed flags

-  `--ignore items`
-  `--ignore quests`
-  `--ignore mounts`
-  `--ignore itemsimages`
-  `--ignore mountsimages`
-  `--mount-image-workers`

## The dofusdude auto-update cycle

> [!NOTE]
> This section describes the original update pipeline by dofusdude. It is left here for informational purposes as it does not directly relate to changes done in this fork.

This tool is the first step in a pipeline that updates the data on [GitHub](https://github.com/dofusdude/dofus2-main) when a new Dofus version is released.

1. Two watchdogs (`doduda listen`) listen for new Dofus versions. One for main and one for beta. When something releases, the watchdog calls the GitHub API to start a workflow that uses `doduda` to download and parse the update to check for new elements and item_types. They hold global IDs for the API, so they must be consistent with each update.
2. At the end of the `doduda` workflow, it triggers the corresponding data repository to do a release, which then downloads the latest `doduda` binary (because it is a different workflow) and runs it to produce the final dataset. The data repository opens a release and uploads the files.
3. After a release, `doduapi` needs to know that a new update is available. The data repository workflow calls the update endpoint. The API then fetches the latest version from GitHub, indexes, starts the image upscaler (if necessary) and does a atomic switch of the database when ready.

## Known Problems

-  Run `doduda` with `--headless` in a server environment or automations to avoid "no tty" errors. This error occurs because the program is trying to access a terminal that doesn't exist.
-  If you get an error regarding a missing Docker socket when running `doduda render`, find out where your `docker.sock` is and link it to the missing path or export your path as `DOCKER_HOST` environment variable `export DOCKER_HOST=unix://<your docker.sock path>`. Docker requires this socket for communication, so if the program can't find it, this error occurs.

## Credit

This fork was created solely for personal needs in my own project, [cori](https://github.com/sebasxs/cori). All core functionality and the original project belong to [stelzo](https://github.com/stelzo).

This tool is for experimental purposes. All game assets and intellectual property belong to Ankama Games.
