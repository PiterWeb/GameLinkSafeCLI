

# GameLinkSafeCLI

<img src="https://github.com/user-attachments/assets/2409330c-3060-4bf9-abc5-46965c10c470" width="128" height="128" />

## Introduction

GameLinkSafeCLI is an Open Source project that aims to connect 2 machines through WebRTC and share ports.
It is an alternative to ngrok/hamachi with TCP/UDP support. No account required.

This programm is a CLI, may not be easy for all users. There is a pay-once [desktop GUI implementation](https://gamelinksafe.org/#pricing), it has no DRM and is compatible with GameLinkSafeCLI & GameLinkSafe for Android.

You can use the software as it comes but if you are behind a strong firewall, a strict NAT or searching for more security you may want to have your own TURN server.
If this is the case you can rent a TURN server or self-host it using [coturn](https://github.com/coturn/coturn)

## Features

- TCP/UDP support
- No accounts
- Plug & play config
- Windows/Linux/MacOS support

## Download (Windows/Linux/MacOS)

https://github.com/PiterWeb/GameLinkSafeCLI/releases/latest

## Download (Free but not open source Android APP)

https://github.com/PiterWeb/GameLinkSafeCLI/releases/download/v1.1/gamelinksafenativeapp.apk

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

Now it should appear some messages with instructions and connection codes.
Finally copy and paste the generated codes between peers.

The resulting connection will work like this:<br>
Host (TCP 127.0.0.1:8000) <-> Client(TCP 127.0.0.1:5000)

## Advanced Config

By default GameLinkSafe uses a [dynamic public STUN server list](https://github.com/pradt2/always-online-stun) but you can specify STUN / TURN servers using a servers.yml file in the same directory as the executable.

This is an example of use:
```yaml
# file: servers.yml

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
