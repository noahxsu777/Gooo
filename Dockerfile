FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gooo ./cmd/server

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /out/gooo /app/gooo
COPY web/static /app/web/static
EXPOSE 8080
ENV ADDR=:8080
ENTRYPOINT ["/app/gooo"]
