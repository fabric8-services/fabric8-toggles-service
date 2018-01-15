F8_AUTH_URL := https://auth.prod-preview.openshift.io
F8_TOGGLES_URL := "http://toggles:4242/api"
F8_KEYCLOAK_URL := "https://sso.prod-preview.openshift.io"

.PHONY: login
## login to oc minishift
login:
	oc login -u developer -p developer && oc new-project f8toggles

.PHONY: deploy-toggles
## deploy toggles server on minishift
deploy-toggles:
	kedge apply -f toggles-db.yml
	kedge apply -f toggles.yml
	oc expose svc toggles

.PHONY: deploy-toggles-service
## deploy
deploy-toggles-service:
	F8_AUTH_URL=$(F8_AUTH_URL) F8_KEYCLOAK_URL=$(F8_KEYCLOAK_URL) F8_TOGGLES_URL=$(F8_TOGGLES_URL) kedge apply -f toggles-service.yml
	oc expose svc toggles-service

.PHONY: clean-minishift
## deploy toggles server on minishift
clean-minishift:
	oc login -u developer -p developer
	oc project f8toggles && oc delete project f8toggles
