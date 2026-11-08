.PHONY: test-social lint-api

test-social:
	cd services/social-service && go test ./... -cover

lint-api:
	npx --yes @redocly/cli lint services/api-gateway/openapi.yaml
