# Wahoo Slim Android Runtime

`wahoo_slim` is the Android-focused mihomo build profile for Wahoo App. It keeps
only the protocol surface the native Android client is expected to consume:

- outbound: `vless`, `direct`, `dns`, `reject`
- inbound/listener: local `socks`/`mixed` and `tun`
- VLESS transport: `xhttp`, `tcp` with Reality
- VLESS security: Reality through `reality-opts`
- TUN stack: `gvisor`; Android passes the `VpnService` file descriptor
  through `tun.file-descriptor`

Unsupported in this profile: VMess, Shadowsocks, Trojan, Hysteria, TUIC,
WireGuard, Snell, AnyTLS, Mieru, Masque, Sudoku, SSH, SOCKS/HTTP inbound,
redir/tproxy/listener servers, smux, UDP forwarding, ECH, XHTTP HTTP/3, and
XHTTP `download-settings`.

Build Android arm64:

```sh
make android-arm64-wahoo-slim
```

The profile enables `with_gvisor`, `with_low_memory`, and `no_fake_tcp` in the
Makefile target. Android clients use Mihomo's own TUN listener with
`stack: gvisor`; do not add a separate tun2socks bridge for the Mihomo runtime.
Keep the default `Meta` build path untouched so upstream merges remain
straightforward.
