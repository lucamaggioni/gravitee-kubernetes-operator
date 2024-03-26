##@ 💻 Development

.PHONY: start-cluster
start-cluster: ## Init and start a local cluster
	npx zx ./scripts/run-kind.mjs

.PHONY: delete-cluster
delete-cluster: ## Delete local cluster
	kind delete cluster --name gravitee

.PHONY: cluster-admin
cluster-admin: ## Gain a kubernetes context with admin role on the local cluster
	kubectl config use-context kind-gravitee
	npx zx ./scripts/create-cluster-admin-sa.mjs
ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: ## Install CRDss into the current cluster
	kubectl apply -f helm/gko/crds

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the current cluster
	kubectl delete -f helm/gko/crds

.PHONY: run
run: ## Run a controller from your host
	go run ./main.go
