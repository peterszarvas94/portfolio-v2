FROM golang:1.25.6-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY templates/ templates/
COPY static/ static/
RUN go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate \
    && go run ./cmd/build \
    && test -s public/index.html && test -s public/resume.pdf

FROM caddy:2.11-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /src/public/ /srv/
USER 10001:10001
EXPOSE 8080
