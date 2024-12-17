FROM golang:1.22.4-alpine

WORKDIR /app

COPY . .

RUN apk add --no-cache git
RUN apk add openssl ca-certificates && \
    openssl s_client -showcerts -servername github.com -connect github.com:443 </dev/null 2>/dev/null | openssl x509 -text >> /etc/ssl/certs/netskope_github.crt && \
    openssl s_client -showcerts -servername storage.googleapis.com -connect storage.googleapis.com:443 </dev/null 2>/dev/null | openssl x509 -text >> /etc/ssl/certs/netskope_googleapis.crt && \
    update-ca-certificates

RUN update-ca-certificates
RUN go install go.opentelemetry.io/collector/cmd/builder@v0.115.0
# ENV GODEBUG=x509ignoreCN=0
RUN builder --config=builder-config.yaml

EXPOSE 8081 4318

CMD ["./otelcol-dev/otelcol-snitch", "--config", "config.yaml"]
