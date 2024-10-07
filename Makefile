# Variables
DOCKER_COMPOSE_FILE := docker-compose.yml
DOCKER_API_BASE_IMAGE_NAME := coda-api
DOCKER_ROUTER_IMAGE_NAME := coda-router
DOCKER_DB_IMAGE_NAME := coda-roundrobin-db
BINARY_API_NAME := coda-api
BINARY_ROUTER_NAME := coda-router
NUM := 1

# Targets
.PHONY: help all api-scale api-stop router-stop api-all-single api-makedocker-multiple api-start-single api-all-multiple compile-api docker-api start-api-multiple router-all router-compile router-makedocker router-start

########### API targets ##########

help:
	@echo "make usage:"
	@echo
	@echo "all\t\t\t\tcompile, build and run 3 instances of the API & 1 instance of router."
	@echo
	@echo "api-all [NUM=<1,2,..N>]\t\tcompile, build and run N instances of the api (executes api-compile, api-makedocker, api-start)"
	@echo "api-stop\t\t\tstop all running instances of the api"
	@echo "api-compile\t\t\tcompile the api code"
	@echo "api-makedocker\t\t\tbuild the docker image for the api"
	@echo "api-run [NUM=<1,2,..N>\t\trun N instances of the api"
	@echo "api-scale [NUM=<1,2,..N>\trun N instances of the api, scaling up or down if necessary"
	@echo
	@echo "router-all\t\t\tcompile, build and run the router instance  (executes router-compile, router-makedocker, router-start)"
	@echo "router-stop\t\t\tstop all running instances of the router"
	@echo "router-compile\t\t\tcompile the router code"
	@echo "router-makedocker\t\tbuild the docker image for the router"
	@echo "router-run\t\t\tstart the router"

########### all target ##########

all:
	$(MAKE) router-all
	$(MAKE) api-all NUM=3

########### API targets ##########

# Build the Go binary for api
api-compile:
	@echo "Building Go API app..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go -C application build -o build/$(BINARY_API_NAME) .

# Build N Docker images for api
api-makedocker:
	@echo "Building Docker image for API..."
	docker build -t $(DOCKER_API_BASE_IMAGE_NAME) application

# run a single instance of api
api-start: api-stop
	@echo "Starting API"
	docker-compose up -d $(DOCKER_API_BASE_IMAGE_NAME) --scale $(DOCKER_API_BASE_IMAGE_NAME)=$(NUM)

api-stop:
	@echo "Stopping all API instances"
	docker compose stop $(DOCKER_API_BASE_IMAGE_NAME)

api-scale:
	@echo "Scale the running API up/down to NUM instances"
	docker-compose up -d $(DOCKER_API_BASE_IMAGE_NAME) --no-recreate --scale $(DOCKER_API_BASE_IMAGE_NAME)=$(NUM)

api-all: api-compile api-makedocker api-start


########### ROUTER targets ##########

# Build the Go binary for router
router-compile:
	@echo "Building Go router app..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go -C router build -o build/$(BINARY_ROUTER_NAME) .

# Build the Docker image for router
router-makedocker:
	@echo "Building Docker image for Go app..."
	docker build -t $(DOCKER_ROUTER_IMAGE_NAME) router

# run a single instance of router
router-start: router-stop
	@echo "Starting router"
	docker compose stop $(DOCKER_ROUTER_IMAGE_NAME)
	docker compose up -d $(DOCKER_ROUTER_IMAGE_NAME)

# build & start a single router instance
router-all: router-compile router-makedocker router-start

router-stop:
	@echo "Stopping router"
	docker compose stop $(DOCKER_ROUTER_IMAGE_NAME)