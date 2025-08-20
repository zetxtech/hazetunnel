package main

import (
	"flag"

	hazetunnel "github.com/zetxtech/hazetunnel/hazetunnel"
)

/*
Launch from CLI
*/

func main() {
	// Parse flags
	var Flags hazetunnel.ProxySetup
	flag.StringVar(&Flags.Addr, "addr", "", "Proxy listen address")
	flag.StringVar(&Flags.Port, "port", "8080", "Proxy listen port")
	flag.StringVar(&Flags.UserAgent, "user-agent", "", "Override the User-Agent header for incoming requests. Optional.")
	flag.StringVar(&Flags.Payload, "payload", "", "Payload to inject into responses. Optional.")
	flag.StringVar(&Flags.UpstreamProxy, "upstream-proxy", "", "Forward requests to an upstream proxy. Optional.")
	flag.StringVar(&Flags.Username, "username", "", "Username for proxy authentication. Optional.")
	flag.StringVar(&Flags.Password, "password", "", "Password for proxy authentication. Optional.")
	flag.StringVar(&hazetunnel.Config.Cert, "cert", "cert.pem", "TLS CA certificate (generated automatically if not present)")
	flag.StringVar(&hazetunnel.Config.Key, "key", "key.pem", "TLS CA key (generated automatically if not present)")
	flag.BoolVar(&hazetunnel.Config.Verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()
	// Set ID
	Flags.Id = "cli"
	// Set verbose level
	hazetunnel.UpdateVerbosity()
	// Launch proxy server
	hazetunnel.Launch(&Flags)
}
