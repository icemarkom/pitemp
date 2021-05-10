FROM golang:latest AS pitemp-builder

# Builder image
WORKDIR /pitemp
COPY . .
RUN go build

# Final image
FROM debian:stable-slim
COPY --from=pitemp-builder /pitemp/pitemp /usr/local/bin/pitemp
ENTRYPOINT ["/usr/local/bin/pitemp"]
CMD ["--help"]
