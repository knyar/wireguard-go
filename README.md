# wireguard-dup

This is a fork of wireguard-go that supports multiple endpoints per peer. When
a peer has multiple endpoints, encapsulated traffic is transmitted to all of them,
with duplicate packets dropped on the other end. The main use case for this is
to allow using several independent internet connections in a "mirror" mode.

Please see the [main wireguard-go repo](https://github.com/WireGuard/wireguard-go)
for the original `README.md` file.

## How to use this

The instructions below assume a Linux based server, and a MacOS client device that
has two internet connections: wifi (primary one) and a mobile phone tethered via
USB (secondary).

You will need to choose the following pieces of configuration:

- Tunnel network (`VPN_NET`) and IP addresses used by the server (`VPN_SERVER_IP`)
  and client (`VPN_CLIENT_IP`)
- Two UDP ports on the server that will be used by the client (`SERVER_PORT1`,
  `SERVER_PORT2`)
- Client and server interface names (`VPN_SERVER_IF`, `VPN_CLIENT_IF`) that don't
  clash with any of your existing interfaces.

Generate server and client keys:

```bash
wg genkey | tee wg-server-priv | wg pubkey > wg-server-pub
wg genkey | tee wg-client-priv | wg pubkey > wg-client-pub
```

Before following the instructions below, set environment variables with all
configuration parameters on each of the machines:

```bash
export VPN_NET=10.129.0.0/24
export VPN_SERVER_IP=10.129.0.1
export VPN_CLIENT_IP=10.129.0.2
export VPN_SERVER_IF=wg0
export VPN_CLIENT_IF=utun8
export SERVER_PORT1=15001
export SERVER_PORT2=15002
export SERVER_EXTERNAL_IP=1.2.3.4
export SERVER_KEY_PRIV=$(base64 -d wg-server-priv | xxd -c 100 -p)
export SERVER_KEY_PUB=$(base64 -d wg-server-pub | xxd -c 100 -p)
export CLIENT_KEY_PRIV=$(base64 -d wg-client-priv | xxd -c 100 -p)
export CLIENT_KEY_PUB=$(base64 -d wg-client-pub | xxd -c 100 -p)
```

### Server Setup (Linux)

- Install OpenBSD netcat (`apt install netcat-openbsd`)
- Check out this repo somewhere (for example, to `/opt/wireguard-dup`)
- Install a recent Go version (1.9 as of Sep 2022) following [instructions](
    https://go.dev/doc/install)
- Build a wireguard-go binary: `cd /opt/wireguard-dup && make`
- If necessary, allow incoming connections to the UDP port via iptables:

```bash
iptables -A INPUT -p udp -m udp --dport $SERVER_PORT1 -j ACCEPT
```

Configure a DNAT firewall rule to make traffic for the second UDP port to be
directed to the first port on a different IP address. It's important for
peer endpoint management for this to be a different IP than
`$SERVER_EXTERNAL_IP`; a localhost port usually works well.

```bash
iptables -t nat -A PREROUTING -d $SERVER_EXTERNAL_IP -p udp -m udp \
  --dport $SERVER_PORT2 -j DNAT --to-destination 127.0.0.1:$SERVER_PORT1
```

Add a systemd unit file, `/etc/systemd/system/wireguard_dup.service`

```bash
cat > /etc/systemd/system/wireguard_dup.service <<EOF
[Unit]
Description=wireguard_dup
Requires=network.target
After=network.target

[Service]
Type=simple
Environment="LOG_LEVEL=debug"
ExecStart=/opt/wireguard-dup/wireguard-go -f $VPN_SERVER_IF

ExecStartPost=/bin/sleep 3
ExecStartPost=/sbin/ip address add dev $VPN_SERVER_IF $VPN_SERVER_IP
ExecStartPost=/sbin/ip link set up dev $VPN_SERVER_IF
ExecStartPost=bash -c '( \\
  echo set=1; \\
  echo listen_port=$SERVER_PORT1; \\
  echo private_key=$SERVER_KEY_PRIV; \\
  echo replace_peers=true; \\
  echo public_key=$CLIENT_KEY_PUB; \\
  echo replace_allowed_ips=true; \\
  echo allowed_ip=$VPN_NET; \\
  echo persistent_keepalive_interval=25; \\
  echo ; \\
) | nc -W 1 -U /var/run/wireguard/$VPN_SERVER_IF.sock'
ExecStartPost=/sbin/ip route add $VPN_NET via $VPN_SERVER_IP

ExecStop=/sbin/ip link del dev $VPN_SERVER_IF
ExecStopPost=/sbin/ip route del $VPN_NET
TimeoutStopSec=3
Restart=always

[Install]
WantedBy=multi-user.target
EOF
```

Enable and start the service:

```bash
systemctl enable wireguard_dup
systemctl start wireguard_dup
```

### Client Setup (Mac OS)

- Check out this repo somewhere (for example, to `~/code/wireguard-dup`)
- Install a recent Go version (1.9 as of Sep 2022) following [instructions](
    https://go.dev/doc/install)
- Build a wireguard-go binary: `cd ~/code/wireguard-dup && make`

Add a script to run wireguard client.

```bash
cat > ~/code/wireguard-dup/run.sh <<EOF
#!/bin/bash

if [[ \$1 == "" ]]; then
  echo "Usage: \$0 <second interface>"
  exit 3
fi
SECOND_IF=\$1

LOG_LEVEL=debug ~/code/wireguard-dup/wireguard-go -f $VPN_CLIENT_IF &

sleep 3
ifconfig $VPN_CLIENT_IF inet $VPN_CLIENT_IP/32 $VPN_SERVER_IP
(
  echo set=1;
  echo listen_port=51821;
  echo private_key=$CLIENT_KEY_PRIV;
  echo replace_peers=true;
  echo public_key=$SERVER_KEY_PUB;
  echo replace_allowed_ips=true;
  echo allowed_ip=$VPN_NET;
  echo endpoint=$SERVER_EXTERNAL_IP:$SERVER_PORT1,$SERVER_EXTERNAL_IP:$SERVER_PORT2;
  echo disable_roaming=true;
  echo persistent_keepalive_interval=25;
  echo ;
) | nc -U /var/run/wireguard/$VPN_CLIENT_IF.sock

ifconfig $VPN_CLIENT_IF up

echo Setting pf rules
SECOND_GW=\$(netstat -nr | grep ^default.*\$SECOND_IF | awk '{ print \$2 }')
echo "
  nat on en0 proto udp from self to $SERVER_EXTERNAL_IP port $SERVER_PORT2 -> (\$SECOND_IF)
  pass out on en0 route-to (\$SECOND_IF \$SECOND_GW) proto udp from self to $SERVER_EXTERNAL_IP port $SERVER_PORT2
" | /sbin/pfctl -Ef -

_exit() {
    echo Resetting pf rules
    /sbin/pfctl -Ef /etc/pf.conf
}
trap _exit SIGTERM SIGINT ERR

wait \$(jobs -p)
EOF
chmod 755 ~/code/wireguard-dup/run.sh
```

Note that the script will completely override your existing PF firewall rules,
instead configuring port-based routing rules for the server's second UDP port. The
script assumes that your main internet connection is wifi (`en0`) and that both
internet connections have default gateway set (which seems to be the case for
iPhone USB tethering).

To start VPN client, run the script, passing your second internet connection
(tethered phone) as a command line argument - for example, if your phone is
`en6`:

```bash
sudo ~/code/wireguard-dup/run.sh en6
```

### Verifying

To check client status, run the following command as root:

```bash
(echo get=1; echo; ) | nc -W 1 -U /var/run/wireguard/$VPN_CLIENT_IF.sock
```

To check server status, run the following command as root:

```bash
(echo get=1; echo; ) | nc -W 1 -U /var/run/wireguard/$VPN_SERVER_IF.sock
```

Both commands should return two endpoints in the `endpoint` variable. The
server should report two different IP addresses for the endpoints: one
for your wifi connection, the second for the tethered phone connection. If
both endpoints have the same IP, it means that the pf-based port rediction
script did not work as expected.

## Possible improvements

- Patch the `wg` command line tool to support multiple endpoints. This should
  significantly simplify setup, since we'll be able to use `wg` instead of
  sending configuration to the Wireguard socket directly.
- Adjust Mac OS script to not override the pf rules completely, but rather
  add and remove them.
