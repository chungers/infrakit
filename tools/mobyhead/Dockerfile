FROM golang:1.8-alpine as builder
RUN apk --no-cache --update add git ca-certificates
WORKDIR /go/src/app
COPY main.go /go/src/app
RUN go-wrapper download
ENV CGO_ENABLED 0
ENV GOOS linux
ARG USERNAME
ARG TOKEN
RUN go build -a -installsuffix cgo -o /go/bin/app -ldflags "-X main.Username=$USERNAME -X main.Token=$TOKEN" .

FROM scratch
COPY --from=builder /go/bin/app /
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
CMD ["/app"]
