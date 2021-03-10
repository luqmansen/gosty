## Deployment Guide

### RabbitMQ
```
kubectl create -f rabbitmq.yaml
```
 ```
helm install rabbit bitnami/rabbitmq -f rabbitmq/helm-values.yaml --create-namespace --namespace gosty
```

### Mongodb

```
kubectl create -f mongodb.yaml
```
 ```
helm install mongodb bitnami/mongodb -f mongodb/helm-values.yaml --create-namespace --namespace gosty
```