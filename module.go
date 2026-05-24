// Package virtualmin provides a Caddy DNS provider module for Virtualmin/Webmin
// (BIND-backed DNS zones). It implements the dns.providers.virtualmin Caddy
// module, which wraps the github.com/emollusion/libdns-virtualmin libdns provider.
//
// # Caddyfile syntax
//
//	tls {
//	    dns virtualmin {
//	        server_url <url>
//	        api_key    <key>
//	        # -- or instead of api_key --
//	        username   <user>
//	        password   <pass>
//	        # optional:
//	        insecure
//	    }
//	    propagation_timeout 2m
//	    propagation_delay   30s
//	}
//
// All string values support Caddy placeholder expansion, including
// {env.VARIABLE_NAME} for environment variables.
//
// # JSON (Caddy API) syntax
//
//	{
//	  "module": "acme",
//	  "challenges": {
//	    "dns": {
//	      "provider": {
//	        "name":       "virtualmin",
//	        "server_url": "https://vps.example.com:10000",
//	        "api_key":    "my-webmin-api-key"
//	      },
//	      "propagation_timeout": "2m",
//	      "propagation_delay":   "30s"
//	    }
//	  }
//	}
//
// # Authentication
//
// Prefer an API key (generated in Webmin → Webmin Users → your user →
// API Tokens) over username/password.  If both api_key and username are
// supplied, api_key takes precedence.
//
// # Minimum Virtualmin version
//
// Virtualmin 7.50.0 or newer is required.  Earlier versions have a bug
// (virtualmin/virtualmin-gpl#1104) that corrupts TXT record values containing
// spaces when written via the API.
package virtualmin

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	libdnsvirtualmin "github.com/emollusion/libdns-virtualmin"
)

func init() {
	caddy.RegisterModule(Provider{})
}

// Provider wraps the libdns-virtualmin provider as a Caddy DNS module.
//
// It exposes the dns.providers.virtualmin module ID and supports both
// Caddyfile and JSON (Caddy API) configuration.
type Provider struct {
	*libdnsvirtualmin.Provider
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.virtualmin",
		New: func() caddy.Module { return &Provider{new(libdnsvirtualmin.Provider)} },
	}
}

// Provision sets up the module and expands any Caddy placeholders present in
// the configuration values (e.g. {env.VIRTUALMIN_URL}).
//
// Implements [caddy.Provisioner].
func (p *Provider) Provision(ctx caddy.Context) error {
	repl := caddy.NewReplacer()
	p.Provider.ServerURL = repl.ReplaceAll(p.Provider.ServerURL, "")
	p.Provider.Username = repl.ReplaceAll(p.Provider.Username, "")
	p.Provider.Password = repl.ReplaceAll(p.Provider.Password, "")
	p.Provider.APIKey = repl.ReplaceAll(p.Provider.APIKey, "")
	p.Provider.DomainOverride = repl.ReplaceAll(p.Provider.DomainOverride, "")
	return nil
}

// UnmarshalCaddyfile parses the provider configuration from a Caddyfile block.
//
// Accepted subdirectives:
//
//	server_url  <url>   — required; Virtualmin/Webmin base URL (e.g. https://host:10000)
//	api_key     <key>   — Webmin API key (preferred over username/password)
//	username    <user>  — Webmin master administrator username
//	password    <pass>  — Webmin master administrator password
//	insecure            — disable TLS certificate verification (flag, no value)
//
// All string values accept Caddy placeholders such as {env.MY_VAR}.
//
// Implements [caddyfile.Unmarshaler].
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		// No positional arguments on the provider name token.
		if d.NextArg() {
			return d.ArgErr()
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "server_url":
				if p.Provider.ServerURL != "" {
					return d.Err("server_url already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.ServerURL = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}

			case "api_key":
				if p.Provider.APIKey != "" {
					return d.Err("api_key already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.APIKey = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}

			case "username":
				if p.Provider.Username != "" {
					return d.Err("username already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.Username = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}

			case "password":
				if p.Provider.Password != "" {
					return d.Err("password already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.Password = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}

			case "insecure":
				// Flag — no value expected.
				if d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.Insecure = true

			case "domain_override":
				if p.Provider.DomainOverride != "" {
					return d.Err("domain_override already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.DomainOverride = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}

			default:
				return d.Errf("unrecognised subdirective %q", d.Val())
			}
		}
	}

	if p.Provider.ServerURL == "" {
		return d.Err("server_url is required")
	}
	if p.Provider.APIKey == "" && p.Provider.Username == "" {
		return d.Err("either api_key or username (with password) is required")
	}

	return nil
}

// Interface guards — compilation fails if the concrete type does not satisfy
// the interfaces that Caddy requires.
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)
