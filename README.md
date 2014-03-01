# TurboTunnel

TurboTunnel creates on-demand ssh tunnels. It listens on local ports,
starts up ssh connections when something connects to those ports, and
proxies data through the remote tunnel.

Trust me, it's magic.

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
$ turbotunnel -config /path/to/config.yml
```

## Using

Once TurboTunnel is running, you can then open `http://localhost:10001`
in your browser. TurboTunnel will see the connection to port 10001 and
initiate an ssh connection to jump1.example.com forwarding a local port
to 10.0.13.10:80. TurboTunnel will then proxy all data between the
opened connection and the local tunnel.

## Building

```bash
$ go get
$ go build
```

