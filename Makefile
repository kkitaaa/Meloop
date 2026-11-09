.PHONY: test-social

test-social:
	cd services/social-service && go test ./... -cover
