<h1 align="center">
    Hazetunnel
</h1>

<h3 align="center">
    🔮 Vindicate non-organic web traffic
</h3>

---

Hazetunnel is an MITM proxy that attempts to legitimize [BrowserForge](https://github.com/daijro/browserforge/)'s injected-browser web traffic by hijacking the TLS fingerprint to mirror the passed User-Agent.

Additionally, it can inject a Javascript payload into the web page to defend against [worker fingerprinting](https://github.com/apify/fingerprint-suite/issues/64).

<hr width=50>

### Features ✨

- Anti TLS fingerprinting 🪪

  - Emulate the ClientHello of browsers based on the passed User-Agent (e.g. Chrome/120)
  - Bypasses TLS fingerprinting checks

- Javascript payload injection 💉

  - Prepends payload to all Javascript responses, including the web page Service/Shared worker scope.
  - Injects payload into embedded base64 encoded JavaScript within HTML responses ([see here](https://github.com/apify/fingerprint-suite/issues/64#issuecomment-1282877696))

This project was built on [tlsproxy](https://github.com/rosahaj/tlsproxy), please leave them a star!

---

# Usage

## CLI Usage

Example: inject `alert('Hello world');` before all Javascript responses:

```bash
./hazetunnel -payload "alert('Hello world');" -port 8080
```

Example: start proxy with authentication (username: `myuser`, password: `mypass`):

```bash
./hazetunnel -username myuser -password mypass -port 8080
```


```
Usage of hazetunnel:
  -addr string
        Proxy listen address
  -cert string
        TLS CA certificate (generated automatically if not present) (default "cert.pem")
  -key string
        TLS CA key (generated automatically if not present) (default "key.pem")
  -password string
        Password for proxy authentication. Optional.
  -payload string
        Payload to inject into responses. Optional.
  -port string
        Proxy listen port (default "8080")
  -upstream-proxy string
        Forward requests to an upstream proxy. Optional.
  -user-agent string
        Override the User-Agent header for incoming requests. Optional.
  -username string
        Username for proxy authentication. Optional.
  -verbose
        Enable verbose logging
```



Python bindings have been removed.

### Payload Injection

#### Javascript responses

This [example server](https://github.com/daijro/hazetunnel/blob/main/example/server.py) will return `console.log('Original JavaScript executed.')` when called:

**Original response:**

```bash
$ curl http://localhost:5000/js
console.log('Original JavaScript executed.');
```

**With Hazetunnel:**

```bash
$ curl http://localhost:5000/js --proxy http://localhost:8080 --cacert cert.pem
alert('Hello world');console.log('Original JavaScript executed.');
```

#### HTML responses

Additionally, Hazetunnel can inject payloads into HTML responses:

```bash
$ curl http://localhost:5000/html --proxy http://localhost:8080 --cacert cert.pem
<!DOCTYPE html>
<body>
    <h1>Base64 JavaScript Testing Page</h1>
    <p>This page includes an embedded base64 encoded JavaScript for testing.</p>
    <script src="data:application/javascript;base64,YWxlcnQoJ0hlbGxvIHdvcmxkJyk7Y29uc29sZS5sb2coJ09yaWdpbmFsIEphdmFTY3JpcHQgZXhlY3V0ZWQuJyk7"></script>
</body>
</html>
```

The embedded base64-encoded script will now decode to this:

```
alert('Hello world');console.log('Original JavaScript executed.');
```

### TLS Spoofing

Hazetunnel will spoof the TLS fingerprint to match the User-Agent passed in the request.

Here is an example of a request to [tls.peet.ws](https://tls.peet.ws/api/clean) through Hazetunnel with a Chrome/121 User-Agent:

```bash
$ curl https://tls.peet.ws/api/clean -H "User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36" --proxy http://localhost:8080 --cacert cert.pem
{
  "ja3": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,16-45-65281-35-5-10-23-0-27-13-65037-11-18-43-17513-51,29-23-24,0",
  "ja3_hash": "1a5edeab8308886ca9716ef14eecd4f3",
  "akamai": "2:0,4:4194304,6:10485760|1073741824|0|a,m,p,s",
  "akamai_hash": "55541b174e8a8adc32544ca36c6fd053",
  "peetprint": "GREASE-772-771|2-1.1|GREASE-29-23-24|1027-2052-1025-1283-2053-1281-2054-1537|1|2|GREASE-4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53|0-10-11-13-16-17513-18-23-27-35-43-45-5-51-65037-65281-GREASE-GREASE",
  "peetprint_hash": "8ad9325e12f531d2983b78860de7b0ec"
}
```

Hazetunnel interprets from the User-Agent that the response is meant to mimic Chrome 121, and sends a ClientHello that mimics Chrome 121 browsers' cipher suites, GREASE functionality, elliptic curves, etc.

It also generates an Akamai HTTP/2 fingerprint.

This supports User-Agents from **Firefox, Chrome, iOS, Android, Edge (legacy), Safari, 360Browser, QQBrowser, etc.**

<hr width=50>

## Python API

Python bindings and CFFI artifacts have been removed in favor of a pure Go CLI.

<hr width=50>

## Building

### CLI

#### Building from source

```bash
git clone https://github.com/daijro/hazetunnel
cd hazetunnel
go build
```

 
