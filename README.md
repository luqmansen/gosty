# Gosty

Scalable cloud transcoding service on Kubernetes

## Table of Content
 - [Gosty](#gosty)
   * [Table of Content](#table-of-content)
   * [System Overview](#system-overview)
   * [Development](#development)
     + [Requirements](#requirements)
     + [Using Docker compose](#using-docker-compose)
     + [Using docker container](#using-docker-container)
     + [Using Kind](#using-kind)
   * [Deployment](#deployment)
     + [Deploy on GKE](#deploy-on-gke)
       - [Using Managed RabbitMQ and MongoDB](#using-managed-rabbitmq-and-mongodb)
       - [Deploy RabbitMQ and MongoDB inside Cluster](#deploy-rabbitmq-and-mongodb-inside-cluster)
     + [RabbitMQ](#rabbitmq)
     + [MongoDB](#mongodb)
     + [API Server, File Server, Worker](#api-server--file-server--worker)
     + [Elasticsearch-Fluentd-Kibana](#elasticsearch-fluentd-kibana)
     + [Linkerd](#linkerd)
   * [Additional](#additional)
     + [Local Image Registry](#local-image-registry)
       - [Using minikube's local registry](#using-minikube-s-local-registry)
       - [Using microk8s local registry](#using-microk8s-local-registry)
     + [K3s local registry](#k3s-local-registry)
       - [Using Docker compose](#using-docker-compose-1)
       - [K8s manifest adjustment](#k8s-manifest-adjustment)
     + [Spekt8](#spekt8)
     + [Dashboard](#dashboard)
     + [Chaos Mesh](#chaos-mesh)
     + [Testing](#testing)
     + [Resize gke cluster to 0 when not used](#resize-gke-cluster-to-0-when-not-used)
   * [Issues](#issues)
   * [What can be improved](#what-can-be-improved)
   * [Todo](#todo)
   * [Acknowledgements](#acknowledgements)
 
 <small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

---

## System Overview

<img width="60%" src="https://github.com/luqmansen/gosty/wiki/out/Diagram/sys-design-overview.png" />
<br>

---

## Development
### Requirements 
- go 16.7 
- docker
- docker-compose
- npm
- yarn
- minikube

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

### Using Kind
Init the kind cluster (mongodb and rabbitmq will run on host machine using docker-compose)
```
bash ./hack/create-kind-local-registry.sh
```

Deploy using kustomize
```
kustomize build deployment/kustomize/environments/gke | kubectl apply -f -
```

### Using Minikube
```
kustomize build deployment/kustomize/environments/minikube | kubectl apply -f -
```

## Deployment

### Deploy on GKE

Init cluster script
```bash
bash ./hack/create-cluster.sh
```

#### Using Managed RabbitMQ and MongoDB

Set your mongodb & rabbitmq secret on configmap (change to secret if you want, modify the manifest accordingly)

Then apply linkerd & gosty component


#### Deploy RabbitMQ and MongoDB inside Cluster

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
kubectl create -fs k8s/rabbitmq/service.yaml # create nodePort service, skip if you don't need
helm install rabbit bitnami/rabbitmq -fs k8s/rabbitmq/helm-values.yaml --create-namespace --namespace gosty
```

### MongoDB

```
kubectl create -fs deployment/k8s/mongodb/service.yaml # create nodePort service, skip if you don't need
helm install mongodb bitnami/mongodb -fs deployment/k8s/mongodb/helm-values.yaml --create-namespace --namespace gosty
```

### API Server, File Server, Worker

Apply the rest of k8s resource manifest

```bash
kustomize build deployment/kustomize/environments/gke | kubectl apply -f -
```

### Elasticsearch-Fluentd-Kibana

This resource will be deployed on `fluentd-monitoring` namespace. This stack currently will only be used for logs
monitoring

```shell
kubectl apply -fs ./deployment/k8s/fluentd
```

```shell
kubectl -n fluentd-monitoring port-forward svc/kibana 5601
```

### Linkerd

Install linkerd cli

```
curl -sL run.linkerd.io/install | sh                                                                                                                                                                       1 ↵
```

install linkerd component

```
kubectl apply -fs deployment/k8s/linkerd/manifest
```

access linkerd dashboard

```bash
linkerd viz dashboard
```

Injecting linkerd to rabbitmq, mongodb & ingress

```bash
kubectl get statefulset -n gosty rabbit-rabbitmq -o yaml  | linkerd inject - | kubectl apply -fs -
kubectl get statefulset -n gosty mongodb -o yaml  | linkerd inject - | kubectl apply -fs -
kubectl get statefulset -n gosty -o yaml mongodb-arbiter | linkerd inject - | kubectl apply -fs -   
kubectl get deployment -n kube-system ingress-nginx-controller -o yaml | linkerd inject - | kubectl apply -fs -                                               1 ↵
```

PromQL for Add additional kubernetes avg cpu usage in grafana dashboard
```
sum (rate (container_cpu_usage_seconds_total{image!="",kubernetes_io_hostname=~"^$Node$",namespace="gosty"}[1m]))
```

## Additional

### Local Image Registry

Run local image registry on my machine for faster dev purposes

#### Using minikube's local registry

In case you want to run it inside kubernetes cluster (minikube), enable minikube local registry

```
minikube addons enable registry
```

forward registry service to local port

```
kubectl port-forward --namespace kube-system svc/registry 5000:80
```

push the local image to minikube's local registry

```
docker build -t localhost:5000/{image-name} -fs docker/Dockerfile-{image-name} .
docker push localhost:5000/{image-name}
```

#### Using microk8s local registry

```shell
microk8s.enable registry
```

You can access it via microk8s's ip (NodePort service)

### K3s local registry

create this file on ``/etc/rancher/k3s/registries.yaml``

```yaml
mirrors:
  registry.local:
    endpoint:
      - "http://192.168.56.1:5000" # your local container registry

```

#### Using Docker compose

```shell
docker-compose -fs docker-compose-registry.yaml up -d registry
```

The pod will be exposed on 0.0.0.0:5000

#### K8s manifest adjustment

Don't forget to change the k8s deployment image

```yaml
spec:
  containers:
    - name: { image-name }
      image: localhost:<registry's cluster ip/container registry's ip>/{image-name}:latest 
```

### Spekt8

I setup [spekt8](https://github.com/spekt8/spekt8) for cluster visualization

```
kubectl create -fs ./deployment/k8s/plugins/spekt8/fabric8-rbac.yaml 
kubectl apply -fs ./deployment/k8s/plugins/spekt8/spekt8-deployment.yaml 
kubectl port-forward -n gosty deployment/spekt8 3000:3000
```

### Dashboard

Accessing k8s dashboard

```shell
kubectl -n kube-system port-forward svc/kubernetes-dashboard 8443:443       
```

### Chaos Mesh

I add some testing scenario on [k8s/chaos](deployment/k8s/chaos) using chaos mesh. First install chaos mesh on the
cluster

```
curl -sSL https://mirrors.chaos-mesh.org/v1.1.2/install.sh | bash
```

### Testing

***Run Pod for Testing Upload File from Inside Container***

```shell
# omit --rm to keep instance after exit
kubectl run curl-test --image=marsblockchain/curl-git-jq-wget  -it -- sh             
```

Download The File

```bash
# this is 20MB Test File
wget --no-check-certificate -r 'https://docs.google.com/uc?export=download&id=102o0T6XeB0znP-r0dkvKhDYTObniockI' -O sony.mp4
# This is 200MB Test File (Blender's Foundation Big Bucks Bunny)
wget --no-check-certificate 'https://docs.google.com/uc?export=download&id=1mw1JHv739M46J6Jv5cXkVHb-_n7O0blK' -r -A 'uc*' -e robots=off -nd # will download 2 files
mv uc?export\=download\&confirm\=uuRp\&id\=1mw1JHv739M46J6Jv5cXkVHb-_n7O0blK bunny.mp4 #rename the files

```

Post data via curl
```
curl http://gosty-apiserver.gosty.svc.cluster.local/api/video/upload -F file=@bunny.mp4 -v
curl http://34.149.27.149/api/video/upload -F file=@bunny.mp4 -v
curl http://gosty-apiserver.gosty.svc.cluster.local/api/video/upload -F file=@sony.mp4 -v
```

Submit query to morph
```
python cli_submit.py -l bunny.mp4 -s 256x144 426x240 640x360 854x480 1280x720 1920x1080
python cli_submit.py -l bunny.mp4 -s 854x480

```

***Execute Flaky Endpoint***

```shell
 while true; do  sleep 60 && curl "http://34.134.157.70/api/scheduler/progress/update"; done # change the ip accordingly                     
```

***GCP Compute Metrics MQL for get average  load (for morph comparison)***
- CPU
```
fetch gce_instance
| metric 'compute.googleapis.com/instance/cpu/utilization'
| group_by 25m , [value_utilization_aggregate: aggregate(value.utilization)/25]
```
- Memory
```
fetch gce_instance
| metric 'agent.googleapis.com/memory/percent_used'
| group_by 25m , [value_percent_used_mean: aggregate(value.percent_used)/25.15]
```
### Resize gke cluster to 0 when not used
To save some bill
```
gcloud container clusters resize ${CLUSTER_NAME} --zone=us-central1-a --num-nodes=0
```

## Issues

**Minikube Ingress Change ip when Ingress controller restarted**

```shell
# change this according to your `kubectl get ingress`
sed -i -e 's/192.168.59.2/'"192.168.59.3"'/g' /etc/hosts
```

**ImagePullBackOff on MicroK8s**
When internet connection is bad, sometimes this is happened (especially for large image)
Solution: set image pull timeout limmit for kubelet re-run the kubelet (somehow I can't edit ExecStart kubelet's
service, so need to manually run it)

```shell
sudo sudo systemctl stop snap.microk8s.daemon-kubelet.service 
sudo /snap/microk8s/2094/kubelet --kubeconfig=/var/snap/microk8s/2094/credentials/kubelet.config --cert-dir=/var/snap/microk8s/2094/certs --client-ca-file=/var/snap/microk8s/2094/certs/ca.crt --anonymous-auth=false --network-plugin=cni --root-dir=/var/snap/microk8s/common/var/lib/kubelet --fail-swap-on=false --cni-conf-dir=/var/snap/microk8s/2094/args/cni-network/ --cni-bin-dir=/var/snap/microk8s/2094/opt/cni/bin/ --feature-gates=DevicePlugins=true --eviction-hard="memory.available<100Mi,nodefs.available<1Gi,imagefs.available<1Gi" --container-runtime=remote --container-runtime-endpoint=/var/snap/microk8s/common/run/containerd.sock --containerd=/var/snap/microk8s/common/run/containerd.sock --node-labels=microk8s.io/cluster=true --authentication-token-webhook=true --cluster-domain=cluster.local --cluster-dns=10.152.183.10 --image-pull-progress-deadline=30s
```

**Random pods evicted**

Microk8s has issue with, even though when the resource is fine, just restart the deployment and delete old resource

```
export NS=<NAMESPACE> 
kubectl -n $NS delete rs $(kubectl -n $NS get rs | awk '{if ($2 + $3 + $4 == 0) print $1}' | grep -v 'NAME')
```

**Pods stuck on terminating state**

Sometimes this is happened on k3s cluster after a while

```bash
export NS=<namespace>
for p in $(kubectl get pods -n $NS | grep Terminating | awk '{print $1}'); do kubectl delete pod -n $NS $p --grace-period=0 --force;done
```

**Hostpath provisioner only writable by root**

If you're running into this [issue](https://github.com/kubernetes/minikube/issues/1990), where the pod won't start
because it can't write to pv, currently my workaround is change the directory modifier. For every node on your cluster,
run below command

````
sudo chmod -R 777 /tmp/hostpath-provisioner/gosty
````

**Viper need .env file**

If you notice, the config.env is still added to final docker images since viper
has [this issue](https://github.com/spf13/viper/issues/584). The env value later will be replaced by injected env values
from K8s

**Scaling statefulsets rabbitmq problem**

RabbitMQ using `erlang cookie` as shared secret used for authentication between RabbitMQ nodes. It will be stored on the
volume. If `erlang cookie` is not defined on secret, it will generated randomly and in case of scaling via kubernetes
API, each node will have different cookie, so they can't work together. Remove previous volume to make sure cookie is
renewed.

**Debug K8s DNS**

Some image has problem for dns resolving, example is this alpine 3.11 and 3.13 (used to use this) with
this [issue](https://github.com/gliderlabs/docker-alpine/issues/539)
Below command is to run one time pod to debug dns

```shell
kubectl run --restart=Never --rm -i --tty alpine --image=alpine:3.12 -- nslookup kube-dns.kube-system.svc.cluster.local
```

**Private registry on MikroK8S
Microk8s private registry communication need to be https, else it won't work, here is
the [reference](https://microk8s.io/docs/registry-private) for the setup


## What can be improved

- Use more proper permanent storage system (consider using object storage, eg: minio, gcs)
- Currently, every worker will always download a copy of the file and process it on its local pod volume, then remove
  the original file, then send the processed file to file server, this can use a lot of bandwidth. This can be improved
  by using shared volume on the node, and check if other worker already download the file, then process it.
  
## Todo
- Run experiment using static mode in CPU Manager K8S 

## Acknowledgements

- Credit to [gibbok](https://github.com/gibbok) for [video player](https://github.com/gibbok/react-video-list) in web client,
(Heavily modified for this project use case)   