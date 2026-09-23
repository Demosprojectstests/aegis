FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY worker ./worker
RUN CGO_ENABLED=0 go build -o /out/worker ./worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/worker /worker
USER 65532:65532
ENTRYPOINT ["/worker"]
