# -----------------------------------------------------------------
# Docker file to copy the generated binary from the `out` directory
# -----------------------------------------------------------------
FROM centos:7
LABEL author Xavier Coulon <xcoulon@redhat.com>

EXPOSE 8080

COPY out/fabric8-toggles-service /usr/local/bin/fabric8-toggles-service

ENTRYPOINT [ "fabric8-toggles-service" ]