FROM golang:latest AS pitemp-builder

# Builder image
RUN mkdir pitemp

COPY go.mod *.go pitemp/

RUN cd pitemp && \
    go build && \
    mv pitemp /usr/local/bin/pitemp

# Final image
FROM debian:stable-slim

COPY --from=pitemp-builder /usr/local/bin/pitemp /usr/local/bin/pitemp

ENTRYPOINT ["/usr/local/bin/pitemp"]

CMD ["--help"]
