FROM centos:7
MAINTAINER "Aslak Knutsen <aslak@redhat.com>"
ENV LANG=en_US.utf8

# Some packages might seem weird but they are required by the RVM installer.
RUN yum --enablerepo=centosplus install -y \
      findutils \
      git \
      golang \
      make \
      mercurial \
      procps-ng \
      tar \
      wget \
      which \
    && yum clean all

# Get 'dep' for Go package management
RUN go get -u github.com/golang/dep/cmd/dep

ENTRYPOINT ["/bin/bash"]
