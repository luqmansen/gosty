# gosty

Kubernetes's compliance scalable cloud transcoding service

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
kubectl create -f k8s/rabbitmq/service.yaml # create nodePort service, skip if you don't need
helm install rabbit bitnami/rabbitmq -f k8s/rabbitmq/helm-values.yaml --create-namespace --namespace gosty
```

### MongoDB

```
kubectl create -f k8s/mongodb/service.yaml # create nodePort service, skip if you don't need
helm install mongodb bitnami/mongodb -f k8s/mongodb/helm-values.yaml --create-namespace --namespace gosty
```

### API Server, File Server, Worker

Apply the rest of k8s resource manifest

```bash
kubect apply -f k8s/gosty
```

### Elasticsearch-Fluentd-Kibana

This resource will be deployed on `fluentd-monitoring` namespace. This stack currently will only be used for logs
monitoring

```shell
kubectl apply -f k8s/fluentd
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

## Additional

### Local Image Registry

I run local image registry on my machine for faster dev purposes

#### Using minikube's local registry

In case you want to run it inside kubernetes cluste (minikube), enable minikube local registry

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
docker-compose -f /home/luqman/Codespace/gosty/docker-compose-registry.yaml up -d registry
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

**Issues**<br>
For some reason, minikube's registry addons doesn't have mounted volume, so everytime Minikube restart, re-push the
image, I create makefile command for this

```
make push-all
```

Microk8s private registry communication need to be https, else it won't work, here is
the [reference](https://microk8s.io/docs/registry-private) for the setup

### Spekt8

I setup [spekt8](https://github.com/spekt8/spekt8) for cluster visualization

```
kubectl create -f k8s/plugins/spekt8/fabric8-rbac.yaml 
kubectl apply -f k8s/plugins/spekt8/spekt8-deployment.yaml 
kubectl port-forward -n gosty deployment/spekt8 3000:3000
```

### Dashboard

Accessing k8s dashboard

```shell
kubectl -n kube-system port-forward svc/kubernetes-dashboard 8443:443       
```

### Chaos Mesh

I add some testing scenario on [k8s/chaos](./k8s/chaos) using chaos mesh. First install chaos mesh on the cluster

```
curl -sSL https://mirrors.chaos-mesh.org/v1.1.2/install.sh | bash
```

### Run curl pod for debugging

```shell
kubectl run curl-test --image=radial/busyboxplus:curl -i --tty --rm            
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


 ___

## What can be improved

- Use more proper permanent storage system
- Currently, every worker will always download a copy of the file and process it on its local pod volume, then remove
  the original file, then send the processed file to file server, this can use a lot of bandwidth. This can be improved
  by using shared volume on the node, and check if other worker already download the file, then process it.

## Acknowledgements

Credit to [gibbok](https://github.com/gibbok) for web client, which I modify for this project use case   