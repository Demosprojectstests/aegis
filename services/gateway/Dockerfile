FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY gateway ./gateway
RUN CGO_ENABLED=0 go build -o /out/gateway ./gateway

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/gateway /gateway
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/gateway"]
