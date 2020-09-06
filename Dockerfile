FROM debian:latest AS pitemp-builder

ARG go_version="1.15.1"
ARG go="/usr/local/go/bin/go"

# Builder image
RUN apt-get update && \
    apt-get install -y curl

RUN curl https://dl.google.com/go/go${go_version}.linux-$(dpkg --print-architecture).tar.gz \
    | tar xz -C /usr/local

RUN mkdir pitemp

COPY *.go pitemp

RUN cd pitemp && \
    ${go} build && \
    mv pitemp /usr/local/bin/pitemp

# Final image
FROM debian:latest

COPY --from=pitemp-builder /usr/local/bin/pitemp /usr/local/bin/pitemp

ENTRYPOINT ["/usr/local/bin/pitemp"]

CMD ["--help"]
