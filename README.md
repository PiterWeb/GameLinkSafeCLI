

# GameLinkSafeCLI

<img src="https://github.com/user-attachments/assets/2409330c-3060-4bf9-abc5-46965c10c470" width="128" height="128" />

## Introduction

GameLinkSafeCLI is an Open Source project that aims to connect 2 machines through WebRTC and share ports.
It is like hamachi but simpler and easier to use. No account required.

This programm is a CLI, may not be easy for all users. There isn't a GUI implementation at the moment.

## Download (Windows/Linux/Macos)

https://github.com/PiterWeb/GameLinkSafeCLI/releases/latest

## Instructions

Shows all commands with examples
```
./gamelinksafecli --help
```

### Example of connection:

Host shares tcp 8000 port
```
./gamelinksafecli --role host --protocol tcp --port 8000
```

Client setup tcp tunnel on port 5000
```
./gamelinksafecli --role client --protocol tcp --port 5000
```
The resulting connection works like this:<br>
Host (8000) <-> Client(5000)

## Advanced Config

By default GameLinkSafe uses a [dynamic public STUN server list](https://github.com/pradt2/always-online-stun) but you can specify STUN / TURN servers using a servers.yml file in the same directory as the executable.

This is an example of use:
```yaml
iceServers:
  # Public STUN servers
  - urls:
      - stun:stun.l.google.com:19305
      - stun:stun.l.google.com:19302
      - stun:stun.ipfire.org:3478
  # TURN server with UDP transport
  # - urls:
  #     - turn:turn.example.com:3478
  #   username: user1
  #   credential: pass1

  # # TURN server with TCP transport
  # - urls:
  #     - turn:turn.example.com:3478?transport=tcp
  #   username: user2
  #   credential: pass2
```
