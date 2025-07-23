#!/bin/bash

set -xe

SCRIPT="$(readlink -f "$0")"
SCRIPTPATH="$(dirname "${SCRIPT}")"
BASEPATH="${SCRIPTPATH}/../../../"

echo "Starting tests that should work..."
kubectl exec -i test-0 -- bash -c 'spike --version'
kubectl exec -i test-0 -- bash -c '/tmp/spiffe-helper -config /tmp/helper.conf'
kubectl exec -i test-0 -- bash -c 'echo "hello from $(date)" > hello.txt'
kubectl exec -i test-0 -- bash -c 'curl -X POST -f -H "Content-Type: application/octet-stream" --data-binary @hello.txt https://spire-spike-nexus.spire-server/v1/cipher/encrypt --cert /tmp/creds/tls.crt --key /tmp/creds/tls.key --cacert /tmp/creds/bundle.0.pem -k -s -o hello.enc'
kubectl exec -i test-0 -- bash -c 'aws --endpoint-url http://minio.minio:9000 s3 cp hello.enc s3://data/test/hello.enc'
kubectl exec -i test-0 -- bash -c 'aws --endpoint-url http://minio.minio:9000 s3 cp s3://data/test/hello.enc hello2.enc'
kubectl exec -i test-0 -- bash -c 'curl -X POST -f -H "Content-Type: application/octet-stream" --data-binary @hello2.enc https://spire-spike-nexus.spire-server/v1/cipher/decrypt --cert /tmp/creds/tls.crt --key /tmp/creds/tls.key --cacert /tmp/creds/bundle.0.pem -k -s -o hello2.txt'
kubectl exec -i test-0 -- bash -c 'cat hello2.txt | grep "hello from "'
echo Done. Tests passed.
