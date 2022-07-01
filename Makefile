# Mount builds a binary ot otlpgenerator.go and mounts it
# with minikube.
#
# Run this as a prerequisite to deploy.
mount:
	mkdir -p ./k8s/bin
	GOOS=linux GOARCH=arm64 go build -o ./k8s/bin/otlpgen ./otlpgenerator.go
	minikube mount ./k8s/bin:/otlpgen

# Applies the k8s deployment.
# Run 'mount' first.
apply:
	kubectl apply -f ./k8s

# destroy deletes the deployment
delete:
	kubectl delete -f ./k8s
