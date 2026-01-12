##################
# Build Stage
##################
FROM golang:1.20-alpine3.18 AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /app
COPY . .
RUN go build -o kt-proxy


##################
# Final Stage
##################
FROM alpine:3.18
RUN apk --no-cache add ca-certificates make

# copy binary
COPY --from=builder /app/kt-proxy /bin/

# log folder
RUN mkdir -p /var/kt-proxy/logs

EXPOSE 9999

CMD ["/bin/kt-proxy"]