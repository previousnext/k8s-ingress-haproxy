FROM golang:1.8 as builder
ADD . /go/src/github.com/previousnext/k8s-ingress-haproxy
WORKDIR /go/src/github.com/previousnext/k8s-ingress-haproxy
RUN go get github.com/mitchellh/gox
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/previousnext/k8s-ingress-haproxy/bin/k8s-ingress-haproxy_linux_amd64 /usr/local/bin/k8s-ingress-haproxy
CMD ["k8s-ingress-haproxy"]
