# MCServerConservationalist

**MCServerConservationalist** is a lightweight utility to manage a Minecraft server with automated sleep and wake functionality. It allows your server to stay offline until a valid player attempts to join, conserving resources, and automatically shuts down when idle.

---

## Practicality

**MCServerConservationalist** has only been tested for the following versions:

- 1.21.9

## Features

- **Automatic Server Wake**: The server starts when a player on the whitelist attempts to log in.
- **Idle Shutdown**: Automatically stops the server after a configurable period of inactivity.
- **Custom MOTD & Icon**: Show a message and optional icon when the server is asleep.
- **Whitelist Control**: Restrict who can wake the server using either `whitelist.json` or a custom list.

---

## Installation

Follow these steps to set up MCServerConservationalist:

1. **Download the files**

   - `MConservationalist.exe` → The compiled server manager, select the correct platform from "builds" folder.
   - `MCServerConservationalist.yaml` → Configuration file.
   - (Optional) `sleeping.png` → 64x64 server icon displayed when the server is asleep.

2. **Create a server folder**

   - Make a new folder for your Minecraft server (if you don’t already have one).
   - Example: `C:\MinecraftServer\` on Windows, `/home/user/minecraft/` on Linux.

3. **Copy files into the folder**

   - Move `MConservationalist.exe` and `MCServerConservationalist.yaml` into this server folder.
   - Add `sleeping.png` if you want a custom sleeping icon.

4. **Configure the YAML**

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

**Tip:** Keep all files in the same folder to avoid path issues. Do not rename the server JAR unless you update `serverJarPath` in the YAML.

---

## Files To Put In Your Minecraft Server Folder

| File                             | Purpose                                                                          |
| -------------------------------- | -------------------------------------------------------------------------------- |
| `MCServerConservationalist.yaml` | Main configuration file for the program.                                         |
| `MConservationalist.exe`         | Compiled executable to run the server manager. Place this in your server folder. |
| `server.jar`                     | Your Minecraft server JAR file.                                                  |
| `sleeping.png`                   | Optional icon to display when the server is asleep.                              |

---

## Configuration (`MCServerConservationalist.yaml`)

```yaml
motd: Server is sleeping - join to wake # Message shown when the server is asleep
idleTimeoutSeconds: 30 # Seconds of inactivity before server shuts down
useWhiteListJson: true # Use whitelist.json for wake permissions
wakeWhiteList: [] # Custom usernames allowed to wake the server (if not using whitelist.json)
sleepingIcon: sleeping.png # Optional icon to display when server is asleep
serverVersion: 1.21.9 # Minecraft server version
serverJarPath: server.jar # Path to your server JAR
```
