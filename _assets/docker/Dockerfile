FROM karalabe/xgo-1.13.x

ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    LANGUAGE=en_US.UTF-8

RUN apt-get update \
 && apt-get install -y libpcsclite-dev locales \
 && apt-get clean \
 && locale-gen ${LANG} \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Jenkins user needs a specific UID/GID to work
RUN groupadd -g 1001 jenkins \
 && useradd --create-home -u 1001 -g 1001 jenkins
USER jenkins

MAINTAINER Jakub Sokolowski "jakub@status.im"
LABEL description="Image for building keycard-cli tool."

ENTRYPOINT ["/build.sh"]
