# 📢 Loudspeaker

[![Docker](https://img.shields.io/docker/v/masanetes/loudspeaker/v0.0.1?color=blue&logo=docker)](https://hub.docker.com/repository/docker/masanetes/loudspeaker)
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

## Supported providers

- Sentry

## Apply Credentials

```bash
$ echo -n 'https://xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx@xxx.xxx.xxx.xxx' > ./dsn.txt
$ kubectl create secret generic sentry-secrets --from-file=dsn=./dsn.txt
```

or

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: sentry-secrets
data:
  dsn: <base64 dsn>
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
  image: "loudspeaker-runtime:latest"
  listeners:
    - type: "sentry"
      credentials: "sentry-secrets"
      subscribes:
        - namespace: "" # all namespaces
          ignore: ["BackoffLimitExceeded"]
        - namespace: "default"
          ignore: ["Unhealthy"]
```
