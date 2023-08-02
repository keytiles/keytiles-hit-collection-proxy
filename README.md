# Keytiles Hit Collection Proxy

## About the project

Contains the Go source code of our slim proxy layer which customers can deploy on their side to protect even the IP address of their own visitors from Keytiles cloud environment.

The IP address of the users are anonymised by zero-ing the host part of the address. The code that is responsible can be seen in  the `AnonymiseIP` function present in [header.go](handler/header/header.go) file.

# Getting Started

The proxy could be setup by either running it as a docker container or building it from source code.

## Build from Source

* Clone the project locally
* Build the project using `go build`

## Using pre-built docker image

* A pre-built docker image could be used from docker hub: https://hub.docker.com/r/keytiles/kt-proxy.
* An example [docker-commpose](docker/docker-compose-example.yaml) file is included for reference.

# Configuring the Proxy

The proxy contains the following configuration options which can be specified as Environment variables.

* HOSTS 
    * Commma separated host names on which the proxy would run on. Currently only two host names can be specifed. If a single host name is specified High Availability cannot be guaranteed.
    * Example: 
        ```
        HOSTS="proxy1.keytiles.com,proxy2.keytiles.com"
        ```
* PORT
    * The port on which the proxy runs. The default value is 9999.
* WHITELIST_HEADERS
    * Commma separated list of headers that would be forwarded to keytiles.
    * !!CAUTION!! Please add the `X-Forwarded-For` header when overriding the `WHITELIST_HEADERS` value.
    * By default, the header `X-Forwarded-For` is added if whitelist headers is empty. The `X-Forwarded-For` is needed to build visit sessions without it metrics like bounce vs non-bounce and complex goal conversion features become unavailable.
    * Example:
        ```
        WHITELIST_HEADERS="x-forwarded-for,keep-alive"
        ```
## Configuring TLS

The proxy could be configured to use server side TLS if desired. You could opt to skip configuring this if you are running the proxy inside a trusted zone.
* TLS_CERT
    * The file path of the server certificate that the proxy should use.
* CERT_KEY
    * The file path of the private key of the server certificate


