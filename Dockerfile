FROM golang:latest AS build

COPY release.go \
    go.mod \
    go.sum \
    /

RUN CGO_ENABLED=0 go build -o /bin/release /release.go

FROM alpine:latest

COPY --from=build /bin/release /bin/release

CMD /bin/release