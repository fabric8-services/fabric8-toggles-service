F8_AUTH_URL ?= https://auth.prod-preview.openshift.io
F8_TOGGLES_URL ?= "http://toggles:4242/api"
F8_KEYCLOAK_URL ?= "https://sso.prod-preview.openshift.io"
f8toggles=f8toggles

.PHONY: build-image-toggles-service
## build toggles image
build-image-toggles:
	make clean deps generate
	make build-linux
	make docker-image-deploy-linux
	docker tag fabric8-toggles-service-deploy $(shell minishift openshift registry)/$(f8toggles)/fabric8togglesservicedocker

.PHONY: push-image-toggles-service
## push toggles-service image to minishift docker registry
push-image-toggles-service: build-image-toggles
	make docker-login
	docker push $(shell minishift openshift registry)/f8toggles/fabric8togglesservicedocker

.PHONY: login
## login to oc minishift
login:
	oc login -u developer -p developer

.PHONY: docker-login
## login to docker
docker-login:
	minishift docker-env
	docker login -u developer -p $(shell oc whoami -t) $(shell minishift openshift registry)

$(f8toggles):
	oc new-project f8toggles
	touch $@

.PHONY: deploy-toggles
## deploy toggles server on minishift
deploy-toggles: login $(f8toggles)
	kedge apply -f toggles-db.yml
	kedge apply -f toggles.yml
	oc expose svc toggles

.PHONY: deploy-toggles-service
## deploy
deploy-toggles-service: login $(f8toggles)
	F8_AUTH_URL=$(F8_AUTH_URL) F8_KEYCLOAK_URL=$(F8_KEYCLOAK_URL) F8_TOGGLES_URL=$(F8_TOGGLES_URL) kedge apply -f toggles-service.yml
	oc expose svc toggles-service

.PHONY: clean-minishift
## deploy toggles server on minishift
clean-minishift: login
	oc project f8toggles && oc delete project f8toggles && rm -rf $(f8toggles)
