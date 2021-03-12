# gosty
Kubernetes compliance scalable cloud transcoding service
![](https://github.com/luqmansen/gosty/wiki/out/Diagram/sys-design-overview.png)
<sup><sup>*diagram need revision</sup></sup> 
___
### Development

#### Using Docker compose
1. `docker-compose up`
2. Change `config.env` to use the docker compose env

#### Testing on Minikube
Run minikube, and refer to [this](k8s/readme.md) deployment guide

#### Using docker image
If you run Database and Message Broker on minikube, make sure to attach minikube network to the container

 example:
```
docker run -it --network minikube luqmansen/gosty-worker
docker run -it --network minikube luqmansen/gosty-apiserver
```

#### Using docker local registry
To speed up experiment with docker image on k8s, enable minikube local registry
```
minikube addons enable registry
```
redirect port 5000 on docker to minikube
```
 docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```
Make sure to push the docker image on local registry
```
docker build -t localhost:5000/gosty-apiserver -f docker/Dockerfile-apiserver .
docker push localhost:5000/gosty-apiserver
```
Don't forget to change the k8s deployment image
```yaml
spec:
  containers:
    - name: gosty-apiserver
      image: localhost:5000/gosty-apiserver:latest 
```


#### Configuration
If you notice, the config.env is still added to final docker images since viper has [this issue](https://github.com/spf13/viper/issues/584), 
the env value will be replaced in K8s anyway
 