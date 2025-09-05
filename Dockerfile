FROM golang:1.24.5 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/app
COPY openapi.yaml /app/openapi.yaml
COPY swagger-ui /app/swagger-ui
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/app"]