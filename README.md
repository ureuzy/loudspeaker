# 📢 Loudspeaker Operator

[![Docker](https://img.shields.io/docker/v/masanetes/loudspeaker/v0.0.3?color=blue&logo=docker)](https://hub.docker.com/repository/docker/masanetes/loudspeaker)
[![Go Reference](https://pkg.go.dev/badge/github.com/masanetes/loudspeaker.svg)](https://pkg.go.dev/github.com/masanetes/loudspeaker)
[![Test](https://github.com/masanetes/loudspeaker/actions/workflows/test.yaml/badge.svg)](https://github.com/masanetes/loudspeaker/actions/workflows/test.yaml)
[![report](https://goreportcard.com/badge/github.com/masanetes/loudspeaker)](https://goreportcard.com/report/github.com/masanetes/loudspeaker)
[![codecov](https://codecov.io/gh/masanetes/loudspeaker/branch/master/graph/badge.svg?token=9HT5CC8XDK)](https://codecov.io/gh/masanetes/loudspeaker)

Loudspeaker retrieves Events from KubeAPI and sends them to the pre-registered Listeners.

```mermaid
flowchart LR
subgraph 4_[Kubernetes Cluster]
  A[KubeAPI] -->|Events| B(loudspeaker)
end  
B -->|Events| C[Listener1]
B -->|Events| D[Listener2]
B -->|Events| E[Listener3]
```

## Install

```
kubectl apply -f https://raw.githubusercontent.com/masanetes/loudspeaker/master/install/install.yaml
```

## Supported listeners

- Sentry

## Preparation of runtime setting

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: loudspeaker-runtime
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: loudspeaker-runtime
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
  - kind: ServiceAccount
    name: loudspeaker-runtime
    namespace: default
EOF
```

## Preparation of confidential listener information
```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: sentry-secrets
type: Opaque
stringData:
  credentilas.yaml: |
    dsn: "sample"
EOF
```

## Sample custom resource

https://github.com/masanetes/loudspeaker/blob/master/config/samples/loudspeaker_v1alpha1_loudspeaker.yaml

```yaml
apiVersion: loudspeaker.masanetes.github.io/v1alpha1
kind: Loudspeaker
metadata:
  name: loudspeaker-sample
spec:
  image: masanetes/loudspeaker-runtime:latest
  serviceAccountName: loudspeaker-runtime  
  listeners:
    - name: foo
      type: sentry
      credentials: sentry-secrets
      subscribes:
        - namespace: "" # all namespaces
          ignore: ["BackoffLimitExceeded"]
        - namespace: "default"
          ignore: ["Unhealthy"]
    
    - name: bar
      type: sentry
      credentials: sentry-secrets
      subscribes:
        - namespace: "" # all namespaces
          ignore: ["BackoffLimitExceeded"]
        - namespace: "default"
          ignore: ["Unhealthy"]
    
    - name: baz
      type: sentry
      credentials: sentry-secrets
      subscribes:
        - namespace: "" # all namespaces
          ignore: ["BackoffLimitExceeded"]
        - namespace: "default"
          ignore: ["Unhealthy"]
```
