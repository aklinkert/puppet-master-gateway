FROM scalify/glide:0.13.0 as builder
WORKDIR /go/src/gitlab.com/scalifyme/puppet-master/gateway/

COPY glide.yaml glide.lock ./
RUN glide install --strip-vendor

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/gateway .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/gitlab.com/scalifyme/puppet-master/gateway/bin/gateway .
RUN chmod +x gateway
ENTRYPOINT ["./gateway"]
