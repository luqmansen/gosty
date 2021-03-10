# gosty

### Developmment

#### Using Docker compose
1. `docker-compose up`
2. Change `config.yaml` to use the docker compose env

#### Testing on Minikube
Run minikube, and refer to [this](k8s/readme.md) deployment guide

#### Using docker image
If you run Database and Message Broker on minikube, make sure to attach minikube network to the container, eg:
 
```
docker run -it --network minikube luqmansen/gosty-worker:latest
```

 