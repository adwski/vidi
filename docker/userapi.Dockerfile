FROM golang:1.22.1-bookworm as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ADD . /build

WORKDIR /build

RUN go mod download
RUN <<EOF
    go build -o userapi -ldflags '-d -w -s' -tags netgo -installsuffix netgo ./cmd/userapi/*.go
    chmod +x /build/userapi
EOF


FROM builder as dev

ENTRYPOINT ["/build/userapi"]


FROM gcr.io/distroless/static as release

WORKDIR /
USER nonroot:nonroot
COPY --from=builder /build/userapi /userapi

ENTRYPOINT ["/userapi"]
