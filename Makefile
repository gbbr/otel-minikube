deploy:
	kubectl apply -f ./k8s
mount:
	mkdir -p ./k8s/bin
	GOOS=linux GOARCH=arm64 go build -o ./k8s/bin/otlpgen ./otlpgenerator.go
	minikube mount ./k8s/bin:/otlpgen
