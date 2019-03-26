FROM karalabe/xgo-1.11.x

ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN apt-get update \
 && apt-get install -y libpcsclite-dev locales \
 && apt-get clean \
 && locale-gen ${LANG} \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

MAINTAINER Jakub Sokolowski "jakub@status.im"
LABEL description="Image for building keycard-cli tool."

ENTRYPOINT ["/build.sh"]
