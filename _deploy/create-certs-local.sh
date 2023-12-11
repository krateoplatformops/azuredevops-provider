#! /bin/bash

WEBHOOK_NS=${1:-"default"}
WEBHOOK_NAME=${2:-"host.minikube.internal"}

rm ./tlsLocal* && rm ./tlsLocal.crt
# Create certs for our webhook
openssl genrsa -out tlsLocal.key 2048
openssl req -new -key ./tlsLocal.key \
    -subj "/CN=${WEBHOOK_NAME}" \
    -addext "subjectAltName = DNS:${WEBHOOK_NAME}" \
    -out ./tlsLocal.csr \
    -config ./csr.conf
openssl x509 -req -extfile <(printf "subjectAltName=DNS:${WEBHOOK_NAME},DNS:${WEBHOOK_NAME}\nbasicConstraints=CA:TRUE\n") -days 365 -in tlsLocal.csr -signkey tlsLocal.key -out tlsLocal.crt

#openssl x509 -noout -text -in ./webhook.crt 

# # Create certs secrets for k8s
# kubectl create secret generic \
#     ${WEBHOOK_NAME}-certs \
#     --from-file=key.pem=./tlsLocal.key \
#     --from-file=cert.pem=./tlsLocal.crt \
#     --dry-run=client -o yaml > ./cluster/webhook-certs.yaml

# Set the CABundle on the webhook registration
CA_BUNDLE=$(cat ./tlsLocal.crt | base64 -b 0)
sed "s/CA_BUNDLE/${CA_BUNDLE}/" ./local/patch.yaml.tpl > ./local/cluster/patch.yaml

cp tlsLocal.crt /tmp/k8s-webhook-server/serving-certs/tls.crt 
cp tlsLocal.key /tmp/k8s-webhook-server/serving-certs/tls.key
