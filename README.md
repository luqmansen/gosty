#gosty
Kubernetes compliance scalable cloud transcoding service

## Architecture diagram
![](https://github.com/luqmansen/gosty/wiki/out/Diagram/sys-design-overview.png)
<sup>*diagram need revision</sup> 
___
## Development

### Using Docker compose
1. `docker-compose up`
2. Change `config.env` to use the docker compose env

### Using docker image
If you run Database and Message Broker on minikube, make sure to attach minikube network to the container

```
docker run -p 8000:8000 --network minikube -e GOSTY_FILESERVER_SERVICE_HOST=192.168.49.4 localhost:5000/gosty-apiserver
docker run -p 8001:8001 --network=minikube localhost:5000/gosty-fileserver
docker run --network minikube -e GOSTY_FILESERVER_SERVICE_HOST=192.168.49.4 localhost:5000/gosty-worker
```

### Using docker local registry on Minikube
To speed up experiment with docker image on k8s when development, enable minikube local registry
```
minikube addons enable registry
```
redirect port 5000 on docker to minikube
```
 docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```
make sure start minikube with additional flag
```
minikube start --insecure-registry="localhost:5000"
```
Make sure to push the docker image on local registry
```
docker build -t localhost:5000/{image-name} -f docker/Dockerfile-{image-name} .
docker push localhost:5000/{image-name}
```
Don't forget to change the k8s deployment image
```yaml
spec:
  containers:
    - name: {image-name}
      image: localhost:5000/{image-name}:latest 
```
**Note**
For some reason, minikube's registry addons doesn't have mounted volume, so everytime
Minikube restart, re-push the image, I create makefile command for this
```
make docker-api
make docker-fs
make docker-worker
```

## Deployment 
RabbitMQ and MongoDB deployed using helm, make sure to install helm before
   
**RabbitMQ**
```
kubectl create -f rabbitmq.yaml
helm install rabbit bitnami/rabbitmq -f rabbitmq/helm-values.yaml --create-namespace --namespace gosty
```

**Mongodb**
```
kubectl create -f mongodb.yaml
helm install mongodb bitnami/mongodb -f mongodb/helm-values.yaml --create-namespace --namespace gosty
```

**Fileserver**
Currently fileserver use local persistent volume, which we have to set up the volume on the related node
for the first time
```
DIRNAME="fileserver-storage"
mkdir -p /home/docker/$DIRNAME 
chmod 777 /home/docker/$DIRNAME
```

**Additional**
I setup [spekt8](https://github.com/spekt8/spekt8) for cluster visualization
```
kubectl create -f k8s/plugins/spekt8/fabric8-rbac.yaml 
kubectl apply -f k8s/plugins/spekt8/spekt8-deployment.yaml 
kubectl port-forward -n gosty deployment/spekt8 3000:3000
```

#### Additional Note
If you notice, the config.env is still added to final docker images since viper has [this issue](https://github.com/spf13/viper/issues/584). 
The env value later will be replaced by injected env values from K8s
 
 ___
### What can be improved
- Use more proper permanent storage system
- Currently, every worker will always download a copy of the file and process it on its local pod volume,
 then remove the original file, then send the processed file to file server, this can use a lot of bandwidth. This can be improved by using shared volume on
  the node, and check if other worker already download the file, then process it.   
