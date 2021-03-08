## Deployment Guide

```
kubectl apply -f rabbitmq.yaml
```
 ```
helm install rabbit bitnami/rabbitmq -f rabbitmq/helm/values.yaml --create-namespace --namespace rabbit
```
