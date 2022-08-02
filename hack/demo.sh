#!/usr/bin/env bash

set -exuo pipefail

export KUBECONFIG=./dist/kubeconfig

PROJECT=spiffe-demo
ARCH=$(go env GOARCH)
VERSION=$(git ls-files | xargs -n 1 cat | md5sum | head -c 7)

# build the demo images
cat .goreleaser.demo.yaml | ARCH=$ARCH envsubst > .goreleaser.demo.$ARCH.yaml
VERSION=$VERSION goreleaser release -f .goreleaser.demo.$ARCH.yaml --snapshot --rm-dist

# create a new kind cluster and connect to it
kind get clusters | grep $PROJECT || kind create cluster --name $PROJECT --image=kindest/node:v1.23.4
kind get kubeconfig --name $PROJECT > ./dist/kubeconfig

# load the demo images
kind load docker-image --name $PROJECT "jetstack/spiffe-demo-server:$VERSION-$ARCH"
kind load docker-image --name $PROJECT "jetstack/spiffe-demo-client:$VERSION-$ARCH"

# load all the images used in dependencies
images=(
  "quay.io/jetstack/cert-manager-controller:v1.9.1" \
  "quay.io/jetstack/cert-manager-cainjector:v1.9.1" \
  "quay.io/jetstack/cert-manager-webhook:v1.9.1" \
  "k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.0" \
  "k8s.gcr.io/sig-storage/livenessprobe:v2.6.0" \
  "quay.io/jetstack/cert-manager-csi-driver-spiffe:v0.2.0" \
  "quay.io/jetstack/cert-manager-csi-driver-spiffe-approver:v0.2.0" \
  "quay.io/jetstack/cert-manager-trust:v0.1.0" \
)
for image in "${images[@]}"
do
  echo preloading $image
  if [ -z "$(docker images -q $image)" ]; then
    docker pull $image
  fi
  kind load docker-image --name $PROJECT $image
done


# Install cert-manager
kubectl apply -f "./deploy/01-cert-manager.yaml"
until cmctl check api; do sleep 5; done

# install CSI driver and trust
kubectl apply -n cert-manager -f "./deploy/02-csi-driver-spiffe.yaml"
kubectl apply -n cert-manager -f "./deploy/03-trust.yaml"
sleep 2
for i in $(kubectl get cr -n cert-manager -o=jsonpath="{.items[*]['metadata.name']}"); do cmctl approve -n cert-manager $i || true ; done

while [ "$(kubectl get deployment -n cert-manager cert-manager-trust -o json | jq '.status.availableReplicas')" != "$(kubectl get deployment -n cert-manager cert-manager-trust -o json | jq '.spec.replicas')" ]
do
  echo "waiting for cm trust to start"
  sleep 1
done

# Bootstrap a self-signed CA
kubectl apply -n cert-manager -f "./deploy/04-selfsigned-ca.yaml"

# Approve Trust Domain CertificateRequest
sleep 2
for i in $(kubectl get cr -n cert-manager -o=jsonpath="{.items[*]['metadata.name']}"); do cmctl approve -n cert-manager $i || true; done

# Prepare trust bundle
kubectl apply -n cert-manager -f "./deploy/05-trust-domain-bundle.yaml"

# install the example server and client
cat "./deploy/08-server.yaml" | ARCH=$ARCH VERSION=$VERSION envsubst | kubectl apply -f -
cat "./deploy/09-client.yaml" | ARCH=$ARCH VERSION=$VERSION envsubst | kubectl apply -f -
