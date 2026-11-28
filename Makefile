.PHONY: test-social test-post lint-api

test-social:
	cd services/social-service && go test ./... -cover

test-post:
	cd services/post-service && go test ./... -cover

lint-api:
	npx --yes @redocly/cli lint services/api-gateway/openapi.yaml
