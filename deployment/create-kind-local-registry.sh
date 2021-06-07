#!/bin/sh
#Use local registry for faster developement
#script sauce: https://kind.sigs.k8s.io/docs/user/local-registry/
set -o errexit

# create registry container unless it already exists
reg_name='kind-registry'
mongodb='mongodb'
rabbitmq='rabbitmq'
reg_port='5000'

running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2.7.1
fi

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
EOF

# connect the registry to the cluster network
# (the network may already be connected)
fuser -k 27017/tcp
fuser -k 5672/tcp
docker-compose up -d "${rabbitmq}" "${mongodb}"
docker network connect "kind" "${reg_name}" || true
docker network connect "kind" "${rabbitmq}" || true
docker network connect "kind" "${mongodb}" || true

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

