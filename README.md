# TurboTunnel

TurboTunnel creates on-demand ssh tunnels. It listens on local ports,
starts up ssh connections when something connects to those ports, and
proxies data through the remote tunnel.

Trust me, it's magic.

## Building

```bash
$ go get github.com/zwily/turbotunnel
$ go build github.com/zwily/turbotunnel
```

## Sample Config

```yaml
tunnels:
  - name: Work Intranet
    localPort: 10001
    jumpHost: jump1.example.com
    remoteHost: 10.0.13.10
    remotePort: 80
  - name: Work Active Directory RDP
    localPort: 10002
    jumpHost: root@jump1.example.com
    remoteHost: 10.0.0.4
    remotePort: 3389
```

## Running

```bash
$ ./turbotunnel -config /path/to/config.yml
```

## Using

Once TurboTunnel is running, you can then open `http://localhost:10001`
in your browser. TurboTunnel will see the connection to port 10001 and
initiate an ssh connection to jump1.example.com forwarding a local port
to 10.0.13.10:80. TurboTunnel will then proxy all data between the
opened connection and the local tunnel.

## Notifications

If you're using OS X, you can get desktop notifications when a tunnel
is started by installing `terminal-notifier`:

```bash
$ brew install terminal-notifier
```

## Running as a Daemon

I run TurboTunnel via launchd. Sample configuration, placed in
`~/Library/LaunchAgents/com.github.zwily.turbotunnel.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.github.zwily.turbotunnel</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/turbotunnel</string>
        <string>-config</string>
        <string>/Users/zach/.turbotunnel.yml</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

To start it after putting the file in place:

```bash
$ launchctl load ~/Library/LaunchAgents/com.github.zwily.turbotunnel.plist
```

