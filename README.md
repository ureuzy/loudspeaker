# Loudspeaker

[![Test](https://github.com/masanetes/loudspeaker/actions/workflows/test.yml/badge.svg)](https://github.com/masanetes/loudspeaker/actions/workflows/test.yml)

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

## Supports Lister

- Sentry

## Sample CRDs

https://github.com/masanetes/loudspeaker/blob/master/config/samples/loudspeaker_v1_loudspeaker.yaml
