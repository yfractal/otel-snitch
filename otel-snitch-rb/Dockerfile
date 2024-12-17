FROM ruby:3.3.6-alpine3.20

RUN apk add --no-cache git build-base curl

RUN apk add openssl ca-certificates && \
    openssl s_client -showcerts -servername github.com -connect github.com:443 </dev/null 2>/dev/null | openssl x509 -text >> /etc/ssl/certs/netskope_github.crt && \
    update-ca-certificates


COPY ./ /app
RUN cd /app && bundle install && rake

WORKDIR /app
