#!/bin/bash -x

set -e

REDSKYCTL_BIN="${REDSKYCTL_BIN:=dist/redskyctl_linux_amd64/redskyctl}"

echo "Upload image to KinD"
[[ -n "${IMG}" ]] && kind load docker-image "${IMG}" --name chart-testing
[[ -n "${SETUPTOOLS_IMG}" ]] && kind load docker-image "${SETUPTOOLS_IMG}" --name chart-testing

echo "Init redskyops"
${REDSKYCTL_BIN} init

echo "Wait for controller"
${REDSKYCTL_BIN} check controller --wait

echo "Create nginx deployment"
kubectl apply -f hack/nginx.yaml

echo "Create ci experiment"
${REDSKYCTL_BIN} generate experiment -f hack/app.yaml | \
  kubectl apply -f -

echo "Create new trial"
${REDSKYCTL_BIN} generate experiment -f hack/app.yaml | \
${REDSKYCTL_BIN} generate trial \
	--default base \
  -f - | \
  kubectl create -f -

kubectl get trial -o wide

Stats() {
	kubectl get po -o wide
  kubectl get trial -o wide
}

trap Stats EXIT

waitTime=300s
echo "Wait for trial to complete (${waitTime} timeout)"
kubectl wait trial \
  -l redskyops.dev/application=ci \
  --for condition=redskyops.dev/trial-complete \
  --timeout ${waitTime}
