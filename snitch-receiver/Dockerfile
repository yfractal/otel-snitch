FROM golang:1.22.4-alpine

WORKDIR /app

COPY . .

RUN go install go.opentelemetry.io/collector/cmd/builder@v0.115.0
# ENV GODEBUG=x509ignoreCN=0
RUN builder --config=builder-config.yaml

EXPOSE 8081 4318

CMD ["./otelcol-dev/otelcol-snitch", "--config", "config.yaml"]
