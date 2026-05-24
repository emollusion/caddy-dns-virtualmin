# caddy-dns-virtualmin

[![Go Reference](https://pkg.go.dev/badge/github.com/caddy-dns/virtualmin.svg)](https://pkg.go.dev/github.com/caddy-dns/virtualmin)

Caddy module: `dns.providers.virtualmin`

This package implements the [Caddy](https://caddyserver.com) DNS provider module for [Virtualmin](https://www.virtualmin.com/), enabling automatic TLS certificates via the ACME DNS-01 challenge for domains whose DNS is managed by Virtualmin/Webmin (BIND).

It wraps the [libdns/virtualmin](https://github.com/libdns/virtualmin) provider library.

## Requirements

- **Virtualmin ≥ 7.50.0** — earlier versions have a bug ([#1104](https://github.com/virtualmin/virtualmin-gpl/issues/1104)) that corrupts TXT record values written via the API.
- The target domain must be a **Virtualmin virtual server with DNS enabled**. See [libdns/virtualmin](https://github.com/libdns/virtualmin) for user setup instructions.

## Building Caddy with this module

Use [xcaddy](https://github.com/caddyserver/xcaddy):

```shell
xcaddy build \
  --with github.com/caddy-dns/virtualmin
```

For local development against unpublished repos:

```shell
xcaddy build \
  --with github.com/caddy-dns/virtualmin=./caddy-dns-virtualmin \
  --with github.com/libdns/virtualmin=./libdns-virtualmin
```

Verify the module loaded:

```shell
./caddy list-modules | grep virtualmin
# dns.providers.virtualmin
```

## Caddyfile

Standard setup with API key:

```caddy
{
    email you@example.com
}

example.com {
    tls {
        dns virtualmin {
            server_url {env.VIRTUALMIN_URL}
            api_key    {env.VIRTUALMIN_API_KEY}
        }
        resolvers ns1.example.com ns2.example.com
        propagation_timeout 2m
        propagation_delay   30s
    }
}
```

With username and password:

```caddy
example.com {
    tls {
        dns virtualmin {
            server_url https://vps.example.com:10000
            username   {env.VIRTUALMIN_USER}
            password   {env.VIRTUALMIN_PASS}
        }
        resolvers ns1.example.com ns2.example.com
        propagation_timeout 2m
        propagation_delay   30s
    }
}
```

With a self-signed Webmin certificate:

```caddy
example.com {
    tls {
        dns virtualmin {
            server_url https://vps.example.com:10000
            api_key    {env.VIRTUALMIN_API_KEY}
            insecure
        }
        resolvers ns1.example.com ns2.example.com
        propagation_timeout 2m
        propagation_delay   30s
    }
}
```

With a Virtualmin sub-server (see [domain_override](#domain_override)):

```caddy
app.example.com {
    tls {
        dns virtualmin {
            server_url      https://vps.example.com:10000
            api_key         {env.VIRTUALMIN_API_KEY}
            domain_override app.example.com
        }
        resolvers ns1.example.com ns2.example.com
        propagation_timeout 2m
        propagation_delay   30s
    }
}
```

## Subdirectives

| Subdirective | Required | Description |
|---|---|---|
| `server_url` | **Yes** | Base URL of the Virtualmin/Webmin server, e.g. `https://host:10000` |
| `api_key` | One of these | Webmin API key (preferred) |
| `username` | ↑ | Webmin username |
| `password` | with username | Webmin password |
| `domain_override` | No | Virtualmin virtual server name to use for all API calls — required for sub-servers |
| `insecure` | No | Disable TLS certificate verification (flag, no value) |

All string values support [Caddy placeholders](https://caddyserver.com/docs/caddyfile/concepts#placeholders), including `{env.VARIABLE}`.

## domain_override

Virtualmin sub-servers (e.g. `app.example.com`) share the zone file of their parent domain (`example.com`). Caddy's ACME client discovers the authoritative zone as `example.com.` and passes that to the provider, but the Virtualmin API manages DNS records under the sub-server name `app.example.com`.

Without `domain_override`, the provider would call `modify-dns --domain example.com` — which may not exist as a virtual server or may be outside the user's access scope.

`domain_override` tells the provider which Virtualmin virtual server to use for all API calls, regardless of which zone Caddy's ACME client discovered.

## resolvers

The `resolvers` subdirective (a standard Caddy TLS option, not specific to this module) pins which DNS servers Caddy uses for zone discovery and challenge verification. Without it, Caddy uses the system resolver (typically 8.8.8.8 / 1.1.1.1), which may occasionally return transient errors during zone discovery.

Pointing `resolvers` at your authoritative nameservers eliminates this:

```caddy
resolvers ns1.example.com ns2.example.com
```

## Authentication

Prefer a **Webmin API key** over username/password:

1. Log in to Webmin.
2. Go to **Webmin Users** → your user → **API Tokens**.
3. Generate a new token and copy it.
4. Set it as `api_key` or in the `VIRTUALMIN_API_KEY` environment variable.

For full user setup instructions including required ACL settings, see [libdns/virtualmin](https://github.com/libdns/virtualmin).

## Propagation settings

After Virtualmin writes a TXT record it calls `rndc reload`, but secondary nameservers outside Virtualmin's cluster may take time to pick up the change.

- `propagation_delay` — how long Caddy waits **before** asking Let's Encrypt to validate. Allows time for secondaries to receive the zone transfer.
- `propagation_timeout` — how long Caddy waits **in total** for the challenge record to be visible before giving up.

A safe starting point:

```caddy
propagation_timeout 2m
propagation_delay   30s
```

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

## License

MIT — see [LICENSE](LICENSE).
