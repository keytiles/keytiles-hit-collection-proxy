# Keytiles Hit Collection Proxy

## About the project

Contains the Go source code of our slim proxy layer which customers can deploy on their side to protect even the IP address of their own visitors from Keytiles cloud environment.

The IP address of the users are anonymised by zero-ing the host part of the address. The code that is responsible can be seen in  the `AnonymiseIP` functino present in `header.go` file.

# Getting Started

The proxy could be setup by either running it as a docker container or building it from source code.

## Build from Source

* Clone the project locally
* Build the project using `go build`

## Using pre-built docker image

* A pre-built docker image could be used from docker hub: https://hub.docker.com/r/keytiles/kt-proxy.
* An example docker-commpose file is included for reference.

# Configuring the Proxy

The proxy contains the following configuration options which can be specified as Environment variables.

| Parameter     | Description                               | Mandatory |Default Value |
| ------------- | ------------------------------------------| ----------|------------- |
| HOSTS         | Commma separated host names on which the proxy would run on. Currently only two host names can be specifed. If a single host name is specified High Availability cannot be guaranteed. | yes | NA  
| PORT          | The port on which the proxy runs.          |  no       |9999 |
| WHITELIST_HEADERS | Commma separated list of headers that would be forwarded to keytiles. | no | Content-Type, Content-Length |
| TLS_CERT | The server certificate that the proxy should use. | yes, only if TLS setup is desired | ""|
| CERT_KEY | The private key of the server certificate | yes, only if TLS setup is desired | ""|