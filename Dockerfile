FROM --platform=${BUILDPLATFORM} golang:alpine AS builder

# Builder image
WORKDIR /pitemp
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN apk --no-cache add ca-certificates
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build

# Final image
FROM alpine:latest AS pitemp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /pitemp/pitemp /
WORKDIR /
ENV PATH "/"
ENTRYPOINT ["pitemp"]
CMD ["--help"]
