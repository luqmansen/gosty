#!/bin/sh
# Point to the internal API server hostname
export APISERVER=https://kubernetes.default.svc
export SERVICEACCOUNT=/var/run/secrets/kubernetes.io/serviceaccount
export NAMESPACE=$(cat ${SERVICEACCOUNT}/namespace)
export TOKEN=$(cat ${SERVICEACCOUNT}/token)
export CACERT=${SERVICEACCOUNT}/ca.crt

export APISERVER_HOST=$(curl --cacert ${CACERT} --header "Authorization: Bearer ${TOKEN}" -X GET ${APISERVER}/api/v1/namespaces/${NAMESPACE}/services/gosty-apiserver/ 2>/dev/null| jq -r '.status | .loadBalancer | .ingress | .[] | .ip')
export APISERVER_PORT=$(curl --cacert ${CACERT} --header "Authorization: Bearer ${TOKEN}" -X GET ${APISERVER}/api/v1/namespaces/${NAMESPACE}/services/gosty-apiserver/ 2>/dev/null| jq -r '.spec | .ports | .[] | .port')
export FILESERVER_HOST=$(curl --cacert ${CACERT} --header "Authorization: Bearer ${TOKEN}" -X GET ${APISERVER}/api/v1/namespaces/${NAMESPACE}/services/gosty-fileserver/ 2>/dev/null|jq -r '.status | .loadBalancer | .ingress | .[] | .ip')
export FILESERVER_PORT=$(curl --cacert ${CACERT} --header "Authorization: Bearer ${TOKEN}" -X GET ${APISERVER}/api/v1/namespaces/${NAMESPACE}/services/gosty-fileserver/ 2>/dev/null| jq -r '.spec | .ports | .[] | .port')

sed -i 's/localhost:8000/'${APISERVER_HOST}':'${APISERVER_PORT}'/g' /usr/share/nginx/html/static/js/*.js
sed -i 's/localhost:8001/'${FILESERVER_HOST}':'${FILESERVER_PORT}'/g' /usr/share/nginx/html/static/js/*.js
