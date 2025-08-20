package api

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"
	utls "github.com/refraction-networking/utls"
	sf "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/utls"
)

type contextKey string

const (
	payloadKey contextKey = "payload"
	authKey    contextKey = "authenticated"
)

type ProxyInstance struct {
	Server *http.Server
	Cancel context.CancelFunc
}

// Globals
var (
	serverMux        sync.Mutex
	proxyInstanceMap = make(map[string]*ProxyInstance)
)

// validateProxyAuth validates HTTP Basic Authentication for proxy requests
func validateProxyAuth(req *http.Request, username, password string) bool {
	// If no authentication is configured, allow all requests
	if username == "" && password == "" {
		return true
	}

	// Get the Proxy-Authorization header
	authHeader := req.Header.Get("Proxy-Authorization")
	if authHeader == "" {
		return false
	}

	// Parse Basic authentication
	if !strings.HasPrefix(authHeader, "Basic ") {
		return false
	}

	encoded := authHeader[6:] // Remove "Basic " prefix
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}

	credentials := string(decoded)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return false
	}

	return parts[0] == username && parts[1] == password
}

func initServer(Flags *ProxySetup) *http.Server {
	serverMux.Lock()
	defer serverMux.Unlock()

	// Load CA if not already loaded
	loadCA()

	// Setup the proxy instance
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = Config.Verbose
	setupProxy(proxy, Flags)

	// Create the server
	server := &http.Server{
		Addr:    Flags.Addr + ":" + Flags.Port,
		Handler: proxy,
	}
	_, cancel := context.WithCancel(context.Background())

	// Add proxy instance to the map
	proxyInstanceMap[Flags.Id] = &ProxyInstance{
		Server: server,
		Cancel: cancel,
	}
	return server
}

func setupProxy(proxy *goproxy.ProxyHttpServer, Flags *ProxySetup) {
	// Handle CONNECT requests (HTTPS) with authentication
	proxy.OnRequest().HandleConnect(goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		// Validate proxy authentication for CONNECT requests
		if !validateProxyAuth(ctx.Req, Flags.Username, Flags.Password) {
			ctx.Resp = proxyAuthRequiredResponse(ctx.Req, ctx)
			return goproxy.RejectConnect, host
		}
		// Mark this connection as authenticated for subsequent requests
		ctx.UserData = map[contextKey]interface{}{
			authKey: true,
		}
		return goproxy.MitmConnect, host
	}))

	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			var upstreamProxy *url.URL

			// Check if this connection was already authenticated via CONNECT
			isAuthenticated := false
			if ctx.UserData != nil {
				if userData, ok := ctx.UserData.(map[contextKey]interface{}); ok {
					if auth, exists := userData[authKey]; exists {
						isAuthenticated = auth.(bool)
					}
				}
			}

			// Validate proxy authentication if configured and not already authenticated
			if !isAuthenticated && !validateProxyAuth(req, Flags.Username, Flags.Password) {
				return req, proxyAuthRequiredResponse(req, ctx)
			}

			// Override the User-Agent header if specified
			// If one wasn't specified, verify a User-Agent is in the request
			if len(Flags.UserAgent) != 0 {
				req.Header["User-Agent"] = []string{Flags.UserAgent}
			} else if len(req.Header["User-Agent"]) == 0 {
				return req, missingParameterResponse(req, ctx, "User-Agent")
			}

			// Set the ClientHello from the User-Agent header
			ua := req.Header["User-Agent"][0]
			clientHelloId, err := getClientHelloID(ua, ctx)
			if err != nil {
				// Use the latest Chrome when the User-Agent header cannot be recognized
				ctx.Logf("Error parsing User-Agent: %s", err)
				clientHelloId = utls.HelloChrome_Auto
				ctx.Logf("Continuing with Chrome %v ClientHello", clientHelloId.Version)
			}

			// Store the payload code in the request's context
			ctx.Req = req.WithContext(
				context.WithValue(
					ctx.Req.Context(),
					payloadKey,
					Flags.Payload,
				),
			)

			// If a proxy header was passed, set it to upstreamProxy
			if len(Flags.UpstreamProxy) != 0 {
				proxyUrl, err := url.Parse(Flags.UpstreamProxy)
				if err != nil {
					return req, invalidUpstreamProxyResponse(req, ctx, Flags.UpstreamProxy)
				}
				upstreamProxy = proxyUrl
			}

			// Build round tripper (applies to both HTTP and HTTPS)
			// Note: upstreamProxy will be nil if not configured
			roundTripper := sf.NewUTLSHTTPRoundTripperWithProxy(clientHelloId, &utls.Config{
				InsecureSkipVerify: true,
				OmitEmptyPsk:       true,
			}, http.DefaultTransport, false, upstreamProxy)

			// Ensure all requests use our custom RoundTripper
			ctx.RoundTripper = goproxy.RoundTripperFunc(
				func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
					return roundTripper.RoundTrip(req)
				})

			return req, nil
		},
	)

	// Inject payload code into responses
	proxy.OnResponse().DoFunc(PayloadInjector)
}

// Launches the server
func Launch(Flags *ProxySetup) {
	server := initServer(Flags)

	// Print server startup message if from CLI or verbose CFFI
	if Flags.Id == "cli" || Config.Verbose {
		log.Println("Hazetunnel listening at", server.Addr)
	}
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
