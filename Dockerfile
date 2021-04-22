FROM golang:latest AS pitemp-builder

# Builder image
RUN mkdir pitemp

COPY *.go pitemp

RUN cd pitemp && \
    go build && \
    go install /usr/local/bin/pitemp

# Final image
FROM nanoserver

COPY --from=pitemp-builder /usr/local/bin/pitemp /usr/local/bin/pitemp

ENTRYPOINT ["/usr/local/bin/pitemp"]

CMD ["--help"]
