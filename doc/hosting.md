# Hosting the service

Start by [building and installing
`privytar`](https://git.sr.ht/~jamesponddotco/privytar#installation),
and then copy [the example `config.json`
file](https://git.sr.ht/~jamesponddotco/privytar/tree/trunk/item/config/config.example.json)
to whatever location you prefer in your filesystem,
`/etc/privytar/config.json` being the recommended location.

[Grab a TLS certificate for your API](https://certbot.eff.org/) and edit
the `config.json` with your preferred settings and the location of your
TLS certificate. Here's an example configuration file:

```json
{
  "service": {
    "name": "Privytar",
    "homepage": "https://privytar.com",
    "contact": "hello@privytar.com",
    "privacyPolicy": "https://privytar.com/privacy",
    "termsOfService": "https://privytar.com/terms"
  },
  "server": {
    "tls": {
      "certificate": "/etc/nginx/ssl/s.privytar.com_ecc/fullchain.cer",
      "key": "/etc/nginx/ssl/s.privytar.com_ecc/api.privytar.com.key",
      "version": "1.2"
    },
    "address": "s.privytar.com:1997",
    "pid": "/var/run/privytar.pid",
    "cacheCapacity": 8192,
    "cacheTTL": "1h"
  }
}
```

Now, to start `privytar`, run this command:

```bash
privytarctl --config /path/to/your/config.json start
```

For production you'll probably want to have a `systemd` service to run
that command for you. Here's a simple example of one.

```bash
[Unit]
Description=Privacy-focused reverse proxy for Gravatar
Documentation=https://sr.ht/~jamesponddotco/privytar/
ConditionFileIsExecutable=/usr/bin/privytarctl
ConditionFileNotEmpty=/etc/privytar/config.json
After=network.target nss-lookup.target

[Service]
Type=simple
UMask=117
ExecStart=/usr/bin/privytarctl --config /etc/privytar/config.json start
ExecStop=/usr/bin/privytarctl --config /etc/privytar/config.json stop
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

For production you'll want to improve your `systemd` service with
sandbox and security features, but that's beyond the scope of this
documentation.

You'll also need to have a server such as NGINX in front of the service,
as it was written to sit behind one. Here's an example `location` for
NGINX.

```nginx
location / {
  proxy_pass https://s.privytar.com:1997;
  proxy_set_header Host $host;
  proxy_http_version 1.1;

  proxy_ssl_server_name on;
  proxy_ssl_protocols TLSv1.2 TLSv1.3;

  proxy_set_header X-Real-IP         $remote_addr;
  proxy_set_header X-Forwarded-Proto $scheme;
  proxy_set_header X-Forwarded-Host  $host;
  proxy_set_header X-Forwarded-Port  $server_port;

  proxy_connect_timeout 60s;
  proxy_send_timeout 60s;
  proxy_read_timeout 60s;
}
```

Again, for production you'll want to improve this `location` and have a
proper NGINX configuration file in place with rate limiting and other
security features, since the service itself doesn't implement any.

With everything up and running, you can now access the service at
`https://${ADDRESS}/avatar/${HASH}`.
