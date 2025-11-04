# MCServerConservationalist

**MCServerConservationalist** is a lightweight utility to manage a Minecraft server with automated sleep and wake functionality. It allows your server to stay offline until a valid player attempts to join, conserving resources, and automatically shuts down when idle.

---

## Tutorial Video

Click the embed below

[![Watch the video](https://img.youtube.com/vi/jJmL9wqSiWA/hqdefault.jpg)](https://www.youtube.com/embed/jJmL9wqSiWA)

## Availability

**MCServerConservationalist** has only been tested for the following versions:

- Vanilla 1.21.10
- Fabric 1.21.9
- Vanilla 1.20.4
- Paper 1.20.4

## Features

- **Automatic Server Wake**: The server starts when a player on the whitelist attempts to log in.
- **Idle Shutdown**: Automatically stops the server after a configurable period of inactivity.
- **Custom MOTD & Icon**: Show a message and optional icon when the server is asleep.
- **Whitelist Control**: Restrict who can wake the server using either `whitelist.json` or a custom list.

---

## Pre-requisites

### ⚠️ IMPORTANT ⚠️

Remember to set `enable-rcon` to `true` & `rcon-port` to `25575` (recommended) in your `server.properties`. This is because MCServerConservationalist uses **RCON commands** to stop the server.

## Installation

Follow these steps to set up MCServerConservationalist:

1. **Download the files**

   - `MConservationalist.exe` → The compiled server manager, select the correct platform from [builds](builds) folder.
   - [`MCServerConservationalist.yaml`](MCServerConservationalist.yaml) → Configuration file.
   - (Optional) [`sleeping.png`](sleeping.png) → 64x64 server icon displayed when the server is asleep.

2. **Create a server folder**

   - Make a new folder for your Minecraft server (if you don’t already have one).
   - Example: `C:\MinecraftServer\` on Windows, `/home/user/minecraft/` on Linux.

3. **Copy files into the folder**

   - Move `MConservationalist.exe` and `MCServerConservationalist.yaml` into this server folder.
   - Add `sleeping.png` if you want a custom sleeping icon.

4. **Configure the [YAML](MCServerConservationalist.yaml)**

   - Open `MCServerConservationalist.yaml` in a text editor.
   - Update settings such as `motd`, `idleTimeoutSeconds`, `serverJarPath`, and whitelist options.
   - Ensure the `serverJarPath` matches the JAR in the folder (default: `server.jar`).

5. **Run the server manager**

   - On Windows: double-click `MConservationalist.exe` or run it in a terminal.
   - On Linux (if using Wine or similar): run `MConservationalist.exe` with the proper environment.
   - The program will start listening for connections and manage server wake/sleep automatically.

6. **Verify**
   - Connect to the server while it is “asleep” to confirm it wakes correctly.
   - Check the console to see messages about server startup and idle monitoring.

---

**Tip:** Keep all files in the same folder to avoid path issues. Do not rename the server JAR unless you update `serverJarPath` in the [YAML](MCServerConservationalist.yaml).

---

## Files To Put In Your Minecraft Server Folder

| File                                                               | Purpose                                                                          |
| ------------------------------------------------------------------ | -------------------------------------------------------------------------------- |
| [`MCServerConservationalist.yaml`](MCServerConservationalist.yaml) | Main configuration file for the program.                                         |
| `MConservationalist.exe`                                           | Compiled executable to run the server manager. Place this in your server folder. |
| `server.jar`                                                       | Your Minecraft server JAR file.                                                  |
| [`sleeping.png`](sleeping.png)                                     | Optional icon to display when the server is asleep.                              |

---

## [Configuration](MCServerConservationalist.yaml)

```yaml
motd: Server is sleeping - join to wake # Message shown when the server is asleep
idleTimeoutSeconds: 30 # Seconds of inactivity before server shuts down
useWhiteListJson: true # Use whitelist.json for wake permissions
wakeWhiteList: [] # Custom usernames allowed to wake the server (if not using whitelist.json)
sleepingIcon: sleeping.png # Optional icon to display when server is asleep
serverVersion: 1.21.9 # Minecraft server version
serverJarPath: server.jar # Path to your server JAR
```

## Issues

If you encounter any problems with the program, please file an issue in the `Issues` tab at the top of this github page. Please be specific as how you're encountering said problem.

Also mention the faulty version number in the provided `Discussions` tab. `Discussions` → `General` → `Report Server Version Availability`. You're welcome to say what version is **NOT** faulty as well.

---

Thank you for using **MCServerConservationalist**.
