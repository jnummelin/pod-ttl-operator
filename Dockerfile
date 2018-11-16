# FROM alpine:3.6

# USER nobody

# ADD build/_output/bin/pod-ttl-operator /usr/local/bin/pod-ttl-operator



FROM golang:1.11 as builder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR  /go/src/github.com/jnummelin/pod-ttl-operator

# Add dependency graph and vendor it in
ADD Gopkg.* /go/src/github.com/jnummelin/pod-ttl-operator/
RUN dep ensure -v -vendor-only

# Add source and compile
ADD . /go/src/github.com/jnummelin/pod-ttl-operator/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pod-ttl-operator cmd/manager/main.go


FROM scratch

COPY --from=builder /go/src/github.com/jnummelin/pod-ttl-operator/pod-ttl-operator /pod-ttl-operator

ENTRYPOINT ["/pod-ttl-operator", "-logtostderr"]
