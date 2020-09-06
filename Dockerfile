FROM debian:latest AS pitemp-builder

ARG go_version="1.15.1"
ARG go="/usr/local/go/bin/go"
ARG pitemp_src="https://github.com/icemarkom/pitemp.git"

RUN apt-get update && apt-get upgrade -y && apt-get install -y gpg git curl
RUN curl https://dl.google.com/go/go${go_version}.linux-$(dpkg --print-architecture).tar.gz -o go.tgz && \
    tar xvfz go.tgz -C /usr/local && \
    rm go.tgz
COPY .gitconfig /etc/gitconfig
RUN git clone ${pitemp_src} pitemp
WORKDIR pitemp
RUN ${go} build && \
    mv pitemp /usr/local/bin/pitemp

FROM debian:latest
COPY --from=pitemp-builder /usr/local/bin/pitemp /usr/local/bin
EXPOSE 9550
ENTRYPOINT ["/usr/local/bin/pitemp"]
CMD ["--help"]
