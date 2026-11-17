.PHONY: test-ml test-social

test-ml:
	cd ml-service && python -m pytest

test-social:
	cd services/social-service && go test ./... -cover
