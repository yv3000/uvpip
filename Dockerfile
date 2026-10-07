# Multi-stage build for isolated reproducible verification and containerized runs.
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/uvpip .

FROM alpine:3.24

RUN apk add --no-cache ca-certificates
COPY --from=builder /out/uvpip /usr/local/bin/uvpip

ENTRYPOINT ["uvpip"]
CMD ["--help"]
