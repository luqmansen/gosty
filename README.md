# gosty

Kubernetes's compliance scalable cloud transcoding service

- [Architecture diagram](#architecture-diagram)
- [Development](#development)
  * [Using Docker compose](#using-docker-compose)
  * [Using docker container + Minikube](#using-docker-container)
  * [Using docker local registry on Minikube](#using-docker-local-registry-on-minikube)
- [Deployment](#deployment)
  * [RabbitMQ](#rabbitmq)
  * [MongoDB](#mongodb)
  * [API Server, File Server, Worker](#api-server--file-server--worker)
- [Additional](#additional)
  * [Linkerd](#linkerd)
  * [Spekt8](#spekt8)
  * [Chaos Mesh](#chaos-mesh)
- [Issues](#issues)
- [What can be improved](#what-can-be-improved)
- [Acknowledgements](#acknowledgements)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with
markdown-toc</a></i></small>

## Architecture diagram

<img width="60%" src="https://github.com/luqmansen/gosty/wiki/out/Diagram/sys-design-overview.png" />
<br>
<small>*Diagram need revision</small>

___

## Development

### Using Docker compose

1. `docker-compose up`
2. Change `config.env` to use the docker compose env

### Using docker container

If you want to run app on container while run the database and message broker on minikube, make sure to attach minikube
network to the container

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
forward registry service to local port
```
kubectl port-forward --namespace kube-system svc/registry 5000:80
```
push the local image to minikube's local registry
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
port is still 5000, since internally, minikube use that port for its local registry

**Note**<br>
For some reason, minikube's registry addons doesn't have mounted volume, so everytime
Minikube restart, re-push the image, I create makefile command for this
```
make push-all
```

## Deployment

RabbitMQ and MongoDB deployed using helm, make sure to install helm before

```
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

also add bitnami repo

```
helm repo add bitnami https://charts.bitnami.com/bitnami
```

### RabbitMQ

```
kubectl create -f k8s/rabbit/rabbitmq.yaml # create nodePort service, skip if you don't need
helm install rabbit bitnami/rabbitmq -f k8s/rabbitmq/helm-values.yaml --create-namespace --namespace gosty
```

### MongoDB

```
kubectl create -f k8s/mongodb/mongodb.yaml # create nodePort service, skip if you don't need
helm install mongodb bitnami/mongodb -f k8s/mongodb/helm-values.yaml --create-namespace --namespace gosty
```

### API Server, File Server, Worker

Apply the rest of k8s resource manifest

```bash
kubect apply -f k8s/gosty
```

## Additional

### Linkerd

Install linkerd cli

```
curl -sL run.linkerd.io/install | sh                                                                                                                                                                       1 ↵
```

install linkerd component
```
kubectl apply -f k8s/linkerd/
```
access linkerd dashboard
```bash
linkerd viz dashboard
```
Injecting linkerd to rabbitmq, mongodb & ingress 
```bash
kubectl get statefulset -n gosty rabbit-rabbitmq -o yaml  | linkerd inject - | kubectl apply -f -
kubectl get statefulset -n gosty mongodb -o yaml  | linkerd inject - | kubectl apply -f -
kubectl get statefulset -n gosty -o yaml mongodb-arbiter | linkerd inject - | kubectl apply -f -   
kubectl get deployment -n kube-system ingress-nginx-controller -o yaml | linkerd inject - | kubectl apply -f -                                               1 ↵
```

### Spekt8

I setup [spekt8](https://github.com/spekt8/spekt8) for cluster visualization

```
kubectl create -f k8s/plugins/spekt8/fabric8-rbac.yaml 
kubectl apply -f k8s/plugins/spekt8/spekt8-deployment.yaml 
kubectl port-forward -n gosty deployment/spekt8 3000:3000
```

### Chaos Mesh

I add some testing scenario on [k8s/chaos](./k8s/chaos) using chaos mesh. First install chaos mesh on the cluster

```
curl -sSL https://mirrors.chaos-mesh.org/v1.1.2/install.sh | bash
```

## Issues

**Minikube Ingress Change ip when Ingress controller restarted**

```shell
# change this according to your `kubectl get ingress`
sed -i -e 's/192.168.59.2/'"192.168.59.3"'/g' /etc/hosts
```

**Random pods evicted**

Microk8s has issue with, even though when the resource is fine, just restart the deployment and delete old resource

```
export NS=<NAMESPACE> 
kubectl -n $NS delete rs $(kubectl -n $NS get rs | awk '{if ($2 + $3 + $4 == 0) print $1}' | grep -v 'NAME')
```

**Hostpath provisioner only writable by root**

If you're running into this [issue](https://github.com/kubernetes/minikube/issues/1990), where the pod won't start
because it can't write to pv, currently my workaround is change the directory modifier.
For every node on your cluster, run below command
````
sudo chmod -R 777 /tmp/hostpath-provisioner/gosty
````

**Viper need .env file**

If you notice, the config.env is still added to final docker images since viper
has [this issue](https://github.com/spf13/viper/issues/584). The env value later will be replaced by injected env values
from K8s



 ___

## What can be improved

- Use more proper permanent storage system
- Currently, every worker will always download a copy of the file and process it on its local pod volume, then remove
  the original file, then send the processed file to file server, this can use a lot of bandwidth. This can be improved
  by using shared volume on the node, and check if other worker already download the file, then process it.

## Acknowledgements

Credit to [gibbok](https://github.com/gibbok) for web client, which I modify for this project use case   