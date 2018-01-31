# -----------------------------------------------------------------
# Docker file to copy the generated binary from the `out` directory
# -----------------------------------------------------------------
FROM centos:7
LABEL author Xavier Coulon <xcoulon@redhat.com>

EXPOSE 8080
ARG BINARY
COPY ${BINARY} /usr/local/bin/fabric8-toggles-service

ENV F8_USER_NAME=fabric8
RUN useradd --no-create-home -s /bin/bash ${F8_USER_NAME}
# From here onwards, any RUN, CMD, or ENTRYPOINT will be run under the following user
USER ${F8_USER_NAME}

ENTRYPOINT [ "fabric8-toggles-service" ]