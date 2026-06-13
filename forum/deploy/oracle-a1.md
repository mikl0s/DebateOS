# Forum Deployment: Oracle A1 Flex (ARM64)

Oracle Cloud's Always Free A1 Flex instances (2 OCPU, 12 GB RAM, ARM64) are
the reference deployment target for the DebateOS Forum service. The Forum is
a single statically-linked binary backed by SQLite — no separate database
process, no container runtime required.

## Build

Cross-compile from any host (Go cross-compilation is zero-dependency):

```bash
GOOS=linux GOARCH=arm64 go build -o forumctl ./forum/cmd/forumctl
```

The output binary is ~15 MB and embeds all SQL migrations at compile time
via `//go:embed`. Copy it to the A1 instance with `scp` or `rsync`:

```bash
scp forumctl ubuntu@<a1-ip>:/usr/local/bin/forumctl
```

## Environment Variables

| Variable                | Required | Default          | Description                                                |
|-------------------------|----------|------------------|------------------------------------------------------------|
| `FORUM_DSN`             | No       | `forum.db`       | SQLite database path (e.g. `/data/forum.db`)               |
| `FORUM_ADDR`            | No       | `:8080`          | TCP address the HTTP server listens on                     |
| `GITHUB_CLIENT_ID`      | Yes      | —                | OAuth app client ID from github.com/settings/applications  |
| `GITHUB_CLIENT_SECRET`  | Yes      | —                | OAuth app client secret (treat as a secret; do not log)    |
| `GITHUB_REDIRECT_URL`   | Yes      | —                | OAuth callback URL, e.g. `https://forum.example.com/oauth/callback` |

Register the OAuth application at:
<https://github.com/settings/applications/new>

Set the **Authorization callback URL** to match `GITHUB_REDIRECT_URL`.

## systemd Unit

Create `/etc/systemd/system/forumctl.service`:

```ini
[Unit]
Description=DebateOS Forum Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=forum
Group=forum
WorkingDirectory=/data
ExecStart=/usr/local/bin/forumctl serve
Restart=on-failure
RestartSec=5

# Secrets via environment — do NOT hardcode in this file.
# Use: systemctl edit forumctl   (creates a Drop-In override)
# Or:  EnvironmentFile=/etc/forum/env  (file mode 0600, owned by root)
EnvironmentFile=/etc/forum/env

# Hardening
PrivateTmp=true
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/data

[Install]
WantedBy=multi-user.target
```

Create `/etc/forum/env` (mode 0600, owned by root):

```bash
FORUM_DSN=/data/forum.db
FORUM_ADDR=:8080
GITHUB_CLIENT_ID=<your-client-id>
GITHUB_CLIENT_SECRET=<your-client-secret>
GITHUB_REDIRECT_URL=https://forum.example.com/oauth/callback
```

Create the `forum` user and data directory:

```bash
useradd --system --no-create-home --shell /usr/sbin/nologin forum
mkdir -p /data
chown forum:forum /data
chmod 700 /data
```

Enable and start:

```bash
systemctl daemon-reload
systemctl enable --now forumctl
journalctl -u forumctl -f   # tail logs
```

## Reverse Proxy (nginx)

The Forum listens on `:8080` (plain HTTP). Put nginx in front for TLS termination:

```nginx
server {
    listen 443 ssl http2;
    server_name forum.example.com;

    ssl_certificate     /etc/letsencrypt/live/forum.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/forum.example.com/privkey.pem;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto https;
    }
}

server {
    listen 80;
    server_name forum.example.com;
    return 301 https://$host$request_uri;
}
```

Obtain the TLS certificate with Certbot:

```bash
certbot --nginx -d forum.example.com
```

## Firewall (OCI Security List)

Open TCP 443 (HTTPS) inbound in the Oracle Cloud Console:
**Networking → Virtual Cloud Networks → <your VCN> → Security Lists → Default → Add Ingress Rule**

Also open the local iptables (Ubuntu default blocks all non-SSH):

```bash
iptables -I INPUT -p tcp --dport 443 -j ACCEPT
iptables -I INPUT -p tcp --dport 80  -j ACCEPT
netfilter-persistent save
```

## DB Recovery (Invariant-4 / FORM-05)

If the SQLite database is lost, run:

```bash
forumctl reindex --registry /path/to/registry/index.json
```

This repopulates all point metadata from the static registry index (git-derived,
not user data). Only ephemeral social state (ratings, subscriptions) is lost —
all catalogue content is recovered from the source-of-truth index.

## Monitoring

Check service health:

```bash
systemctl status forumctl
curl -s http://127.0.0.1:8080/api/points?limit=1 | python3 -m json.tool
```

Expected response: `{"points": [...]}` with at least one point if Reindex
has been run since deploy.
