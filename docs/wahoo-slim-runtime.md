# Wahoo Slim Android Runtime

`wahoo_slim` is the Android-focused mihomo build profile for Wahoo App. It keeps
only the protocol surface the native Android client is expected to consume:

- outbound: `vless`, `direct`, `dns`, `reject`
- inbound/listener: `tun`
- VLESS transport: `xhttp` only
- VLESS security: Reality through `reality-opts`

Unsupported in this profile: VMess, Shadowsocks, Trojan, Hysteria, TUIC,
WireGuard, Snell, AnyTLS, Mieru, Masque, Sudoku, SSH, SOCKS/HTTP inbound,
redir/tproxy/listener servers, smux, UDP forwarding, ECH, XHTTP HTTP/3, and
XHTTP `download-settings`.

Build Android arm64:

```sh
make android-arm64-wahoo-slim
```

The profile also enables `with_low_memory` and `no_fake_tcp` in the Makefile
target. Keep the default `Meta` build path untouched so upstream merges remain
straightforward.
