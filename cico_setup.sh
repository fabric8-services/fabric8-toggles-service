#!/bin/bash

# Output command before executing
set -x

# Exit on error
set -e

# Source environment variables of the jenkins slave
# that might interest this worker.
function load_jenkins_vars() {
  if [ -e "jenkins-env" ]; then
    cat jenkins-env \
      | grep -E "(DEVSHIFT_TAG_LEN|DEVSHIFT_USERNAME|DEVSHIFT_PASSWORD|JENKINS_URL|GIT_BRANCH|GIT_COMMIT|BUILD_NUMBER|ghprbSourceBranch|ghprbActualCommit|BUILD_URL|ghprbPullId)=" \
      | sed 's/^/export /g' \
      > ~/.jenkins-env
    source ~/.jenkins-env
  fi
}

function install_deps() {
  # We need to disable selinux for now, XXX
  /usr/sbin/setenforce 0

  # Get all the deps in
  yum -y install \
    docker \
    make \
    git \
    curl

  service docker start

  echo 'CICO: Dependencies installed'
}

function prepare() {
  # Let's test
  make docker-start
  make docker-check-go-format
  make docker-deps
  make docker-generate
  make docker-build
  echo 'CICO: Preparation complete'
}

function run_tests_without_coverage() {
  make docker-test
  echo "CICO: ran tests without coverage"
}

function tag_push() {
  TARGET=$1
  docker tag fabric8-toggles-service-deploy $TARGET
  docker push $TARGET
}

function deploy() {
  # Let's deploy
  make docker-image-deploy

  TAG=$(echo $GIT_COMMIT | cut -c1-${DEVSHIFT_TAG_LEN})
  REGISTRY="push.registry.devshift.net"

  if [ -n "${DEVSHIFT_USERNAME}" -a -n "${DEVSHIFT_PASSWORD}" ]; then
    docker login -u ${DEVSHIFT_USERNAME} -p ${DEVSHIFT_PASSWORD} ${REGISTRY}
  else
    echo "Could not login, missing credentials for the registry"
  fi

  tag_push ${REGISTRY}/fabric8-services/fabric8-toggles-service:$TAG
  tag_push ${REGISTRY}/fabric8-services/fabric8-toggles-service:latest
  echo 'CICO: Image pushed, ready to update deployed app'
}

function cico_setup() {
  load_jenkins_vars;
  install_deps;
  prepare;
}
