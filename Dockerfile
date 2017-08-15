FROM golang:1.8.3 as build
WORKDIR /go/src/kubernetes-cloudwatch-exporter

# install glide
RUN curl https://glide.sh/get | sh

# copy in code, resolve dependencies and build
COPY . .

RUN    glide up -v \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  

RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*

WORKDIR /root/
COPY --from=build /go/src/kubernetes-cloudwatch-exporter/app .
CMD ["./app"] 