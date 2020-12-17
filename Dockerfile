FROM debian:latest AS pitemp-builder

ARG go_command="/usr/local/go/bin/go"

# Builder image
RUN apt-get update && \
    apt-get install -y curl gcc && \
    curl https://dl.google.com/go/$(curl -s "https://golang.org/VERSION?m=text").linux-$(dpkg --print-architecture | sed -e 's/armhf/armv6l/').tar.gz \
    | tar xz -C /usr/local && \
    mkdir pitemp

COPY *.go pitemp

RUN cd pitemp && \
    ${go_command} build && \
    mv pitemp /usr/local/bin/pitemp

# Final image
FROM debian:latest

COPY --from=pitemp-builder /usr/local/bin/pitemp /usr/local/bin/pitemp

ENTRYPOINT ["/usr/local/bin/pitemp"]

CMD ["--help"]
