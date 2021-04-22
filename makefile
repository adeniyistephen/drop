SHELL := /bin/bash


# ==============================================================================
# Testing running system NB: Get your enviroment's ready.
# All testing goes for both user and studio.

# For testing a simple query on the system. Don't forget to add a user to the mongodb
# curl -d '{"name":"justyn", "email":"justyn@test.com", "roles":["ADMIN","USER"], "password":"mypass", "password_confirm":"mypass"}' -H "Content-Type: application/json" -X POST http://localhost:3000/v1/users

# curl --user "justyn@test.com" http://localhost:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# export TOKEN="COPY TOKEN STRING FROM LAST CALL"

# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users/1/1

# For testing load on the service.
# hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users/1/1
# hey -m POST -c 100 -n 100000 -d '{"name":"justyn", "email":"justyn@test.com", "roles":["ADMIN","USER"], "password":"mypass", "password_confirm":"mypass"}' -H "Content-Type: application/json" http://localhost:3000/v1/users

# zipkin: http://localhost:9411
# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,mem:memstats.Alloc"

# Used to install expvarmon program for metrics dashboard.
# go install github.com/divan/expvarmon@latest

# // To generate a private/public key PEM file.
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem
# ./drop-admin genkey

# ==============================================================================
# Building containers

all: drop metrics

drop:
	docker build \
		-f scripts/docker/dockerfile.drop-api \
		-t drop-api-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f scripts/docker/dockerfile.metrics \
		-t metrics-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

#=============================================================
# Running from within docker compose
run: up

up:
	docker-compose -f scripts/compose/compose.yaml up --detach --remove-orphans

down:
	docker-compose -f scripts/compose/compose.yaml down --remove-orphans

logs:
	docker-compose -f scripts/compose/compose.yaml logs -f


# ==============================================================
# Running from within k8s/dev

kind-up:
	kind create cluster --image kindest/node:v1.20.2 --name drop-starter-cluster --config scripts/k8s/dev/kind-config.yaml

kind-down:
	kind delete cluster --name drop-starter-cluster

kind-load:
	kind load docker-image drop-api-amd64:1.0 --name drop-starter-cluster
	kind load docker-image metrics-amd64:1.0 --name drop-starter-cluster

kind-drop:
	kustomize build scripts/k8s/dev | kubectl apply -f -

kind-status:
	kubectl get nodes
	kubectl get pods --watch	

kind-logs:
	kubectl logs -lapp=drop-api --all-containers=true -f --tail=100

kind-status-full:
	kubectl describe pod -lapp=drop-api

kind-status-full-mongo:
	kubectl describe pod -lapp=mongo	

kind-update: drop
	kind load docker-image drop-api-amd64:1.0 --name drop-starter-cluster
	kubectl delete pods -lapp=drop-api

kind-metrics: metrics
	kind load docker-image metrics-amd64:1.0 --name drop-starter-cluster
	kubectl delete pods -lapp=drop-api

kind-delete:
	kustomize build zarf/k8s/dev | kubectl delete -f -	

#================================================================
# Modules support

run:
	go run app/drop-api/main.go

tidy:
	go mod tidy
	go mod vendor	
	