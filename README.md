# Go service template

Service template. When cloning, specify the name of the directory for cloning, which has the name of the service.
After cloning, you need to edit the file Makefile.env and run: make -f Makefile.env.
After that, the directives described below will become available.

## Build
```
Local: make build
Docker: make docker_build
```
## Tests
```
Run Unit tests in docker
make tests
```
## Startup
```
Up docker environment: make docker_env_up
Shutdown all: make docker_env_down
```
```
Up docker environment: make local_env_up
Shutdown all: make local_env_down
```

## Distributive
```
Create archieve with configs and app
make tar
```
```
Create docker image and try to push it in docker registry with version and 'latest' tags  
make build_image
```

## Metrics
```
Added possibility to see service metrics after running it in docker:

Address - localhost:3000
Login/Password - admin/P@ssw0rd (can change in /deployments/docker/grafana/config.monitoring)
```
