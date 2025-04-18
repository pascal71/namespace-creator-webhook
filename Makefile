# Define the container registry
KO_DOCKER_REPO ?= your-registry.io/namespace-creator-webhook
export KO_DOCKER_REPO

# Run tests
test:
	go test ./... -coverprofile cover.out

# Build the binary locally
build:
	go build -o bin/webhook ./cmd/webhook

# Build and publish image with Ko
ko-build:
	ko build ./cmd/webhook

# Deploy with Ko
ko-deploy:
	ko apply -f config/kubernetes/webhook.yaml

# Deploy the webhook configuration (separate because of CA bundle)
deploy-webhook-config:
	# Get the CA Bundle
	$(eval CA_BUNDLE := $(shell kubectl get secret webhook-server-cert -n webhook-system -o jsonpath='{.data.ca\.crt}'))
	# Replace and apply
	sed -e "s|\$${CA_BUNDLE}|$(CA_BUNDLE)|g" config/kubernetes/webhook-configuration.yaml | kubectl apply -f -

# Create the namespace
create-namespace:
	kubectl create namespace webhook-system || true

# Deploy cert-manager resources
deploy-certs:
	kubectl apply -f config/kubernetes/cert-manager.yaml

# Deploy everything
deploy: create-namespace deploy-certs ko-deploy
	# Wait for cert to be ready
	kubectl wait --for=condition=Ready -n webhook-system certificate/webhook-server-cert --timeout=60s
	$(MAKE) deploy-webhook-config

# Clean up
clean:
	kubectl delete -f config/kubernetes/webhook-configuration.yaml || true
	kubectl delete -f config/kubernetes/webhook.yaml || true
	kubectl delete -f config/kubernetes/cert-manager.yaml || true
