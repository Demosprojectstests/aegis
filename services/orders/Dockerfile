FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY orders ./orders
RUN CGO_ENABLED=0 go build -o /out/orders ./orders

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/orders /orders
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/orders"]
