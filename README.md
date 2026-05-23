# caddy-dns/virtualmin

[![Go Reference](https://pkg.go.dev/badge/github.com/caddy-dns/virtualmin.svg)](https://pkg.go.dev/github.com/caddy-dns/virtualmin)

Caddy module: `dns.providers.virtualmin`

This package implements the [Caddy](https://caddyserver.com) DNS provider module for [Virtualmin](https://www.virtualmin.com/), enabling automatic TLS certificates via the ACME DNS-01 challenge for domains whose DNS is managed by Virtualmin/Webmin (BIND).

It wraps the [libdns/virtualmin](https://github.com/libdns/virtualmin) provider library.

## Requirements

- **Virtualmin ≥ 7.50.0** — earlier versions have a bug ([#1104](https://github.com/virtualmin/virtualmin-gpl/issues/1104)) that corrupts TXT record values.
- The authenticating Webmin user must be the **master administrator** (`root` or `admin`).
- The Webmin user must have the *Virtualmin Remote CLI* ACL bit enabled.

## Building Caddy with this module

Use [xcaddy](https://github.com/caddyserver/xcaddy):

```shell
xcaddy build \
  --with github.com/caddy-dns/virtualmin
```

Or reference your local checkout during development:

```shell
xcaddy build \
  --with github.com/caddy-dns/virtualmin=./caddy-dns-virtualmin \
  --with github.com/libdns/virtualmin=./libdns-virtualmin
```

## Caddyfile

```caddy
{
    # Global ACME defaults (optional)
    email you@example.com
}

example.com {
    tls {
        dns virtualmin {
            server_url {env.VIRTUALMIN_URL}
            api_key    {env.VIRTUALMIN_API_KEY}
        }
        propagation_timeout 2m
        propagation_delay   30s
    }
}
```

Using username and password instead of an API key:

```caddy
example.com {
    tls {
        dns virtualmin {
            server_url https://vps.example.com:10000
            username   {env.VIRTUALMIN_USER}
            password   {env.VIRTUALMIN_PASS}
        }
        propagation_timeout 2m
        propagation_delay   30s
    }
}
```

With a self-signed Webmin certificate (not recommended for production):

```caddy
example.com {
    tls {
        dns virtualmin {
            server_url https://vps.example.com:10000
            api_key    {env.VIRTUALMIN_API_KEY}
            insecure
        }
    }
}
```

## Subdirectives

| Subdirective | Required | Description |
|---|---|---|
| `server_url` | **Yes** | Base URL of the Virtualmin/Webmin server, e.g. `https://host:10000` |
| `api_key` | One of these | Webmin API key (preferred) |
| `username` | ↑ | Webmin master administrator username |
| `password` | with username | Webmin master administrator password |
| `insecure` | No | Disable TLS certificate verification (flag, no value) |

All string values support [Caddy placeholders](https://caddyserver.com/docs/caddyfile/concepts#placeholders), including `{env.VARIABLE}`.

## JSON (Caddy API)

```json
{
  "apps": {
    "tls": {
      "automation": {
        "policies": [
          {
            "subjects": ["example.com"],
            "issuers": [
              {
                "module": "acme",
                "challenges": {
                  "dns": {
                    "provider": {
                      "name":       "virtualmin",
                      "server_url": "https://vps.example.com:10000",
                      "api_key":    "my-webmin-api-key"
                    },
                    "propagation_timeout": "2m",
                    "propagation_delay":   "30s"
                  }
                }
              }
            ]
          }
        ]
      }
    }
  }
}
```

## Authentication

Prefer a **Webmin API key** over username/password:

1. Log in to Webmin.
2. Go to **Webmin Users** → your user → **API Tokens**.
3. Generate a new token and copy it.
4. Pass it as `api_key` or in the `VIRTUALMIN_API_KEY` environment variable.

If you use username/password, the credentials grant full control of the
Webmin server.  Reduce risk by:

- Running Caddy on the **same host** as Virtualmin and using
  `server_url https://127.0.0.1:10000`.
- Creating a **dedicated Webmin admin user** with a strong, unique password.

## Propagation settings

After Virtualmin writes a TXT record it calls `rndc reload`, but secondary
nameservers that are not part of Virtualmin's cluster may take time to pick
up the change.  Tune `propagation_timeout` and `propagation_delay` in your
Caddyfile to match your DNS propagation characteristics:

- `propagation_delay` — how long Caddy waits **before** asking Let's Encrypt
  to validate.  Allows time for secondaries to receive the NOTIFY and
  transfer.
- `propagation_timeout` — how long Caddy waits **in total** for the challenge
  record to be visible in DNS before giving up.

A safe starting point for most setups is:

```caddy
propagation_timeout 2m
propagation_delay   30s
```

## License

MIT — see [LICENSE](LICENSE).
