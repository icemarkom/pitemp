FROM golang:alpine AS pitemp-builder

# Builder image
WORKDIR /pitemp
COPY . .
RUN GOOS="linux" GOARCH=$(uname -m | sed -e "s/aarch64/arm64/" -e "s/x86_64/amd64/" -e "s/armv7l/arm/") go build

# Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
#RUN apt-get update && apt-get install --yes ca-certificates
COPY --from=pitemp-builder /pitemp/pitemp /usr/local/bin/pitemp
ENTRYPOINT ["/usr/local/bin/pitemp"]
CMD ["--help"]
