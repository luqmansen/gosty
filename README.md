# gosty

### Developmment

#### Using Docker compose
1. `docker-compose up`
2. Change `config.yaml` to use the docker compose env

#### Testing on Minikube
Run minikube, and refer to [this](k8s/readme.md) deployment guide

#### Using docker image
If you run Database and Message Broker on minikube, make sure to attach minikube network to the container

 example:
```
docker run -it --network minikube luqmansen/gosty-worker
docker run -it --network minikube luqmansen/gosty-apiserver
```

#### Configuration
If you notice, the config.env is still added to final docker images since viper has [this issue](https://github.com/spf13/viper/issues/584), 
the env value will be replaced in K8s anyway
 