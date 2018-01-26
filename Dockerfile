# ------------------------------------------------------------
# Build image
# ------------------------------------------------------------
FROM centos:7 as builder
LABEL author "Xavier Coulon <xcoulon@redhat.com>"
ENV LANG=en_US.utf8
ARG BINARY=bin/fabric8-toggles-service
ARG LDFLAGS

# Some packages might seem weird but they are required by the RVM installer.
RUN yum --enablerepo=centosplus install -y \
      findutils \
      git \
      golang \
      mercurial \
      procps-ng \
      which \
    && yum clean all

# the go root.
RUN mkdir /go 
ENV GOPATH=/go

# import the project to build
ADD . /go/src/github.com/fabric8-services/fabric8-toggles-service
WORKDIR /go/src/github.com/fabric8-services/fabric8-toggles-service

# Get 'dep' for Go package management and verify dependencies
RUN go get -u github.com/golang/dep/cmd/dep
RUN ${GOPATH}/bin/dep ensure

# Run the tests in all packages but exclude `/vendor`
RUN go test -v $(go list ./... | grep -v /vendor/)
# build the binary
RUN go build -v ${LDFLAGS} -o ${BINARY}



# ------------------------------------------------------------
# Final image
# ------------------------------------------------------------
FROM centos:7
ENV LANG=en_US.utf8
ENV INSTALL_DIR=/usr/local/fabric8-toggles-service
ARG BINARY=bin/fabric8-toggles-service


# Add the binary file generated in the `builder` container above
RUN mkdir -p ${INSTALL_DIR}/bin
COPY --from=builder /go/src/github.com/fabric8-services/fabric8-toggles-service/${BINARY} ${INSTALL_DIR}/${BINARY}

# Create a non-root user and a group with the same name: "fabric8"
ENV F8_USER_NAME=fabric8
RUN useradd --no-create-home -s /bin/bash ${F8_USER_NAME}

RUN cd /tmp \
    && curl -L https://github.com/openshift/origin/releases/download/v3.6.0/openshift-origin-client-tools-v3.6.0-c4dd4cf-linux-64bit.tar.gz > openshift-origin-client-tools.tar.gz \
    && tar xvzf openshift-origin*.tar.gz \
    && mv openshift-origin*/oc /usr/bin \
    && rm -rfv openshift-origin*

# From here onwards, any RUN, CMD, or ENTRYPOINT will be run under the following user
USER ${F8_USER_NAME}


WORKDIR ${INSTALL_DIR}
ENTRYPOINT [ "/usr/local/fabric8-toggles-service/bin/fabric8-toggles-service" ]

EXPOSE 8080