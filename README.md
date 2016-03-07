# kube2consul

A bridge between Kubernetes and Consul. This will watch the kubernetes API for
changes in Services and then register those Services in Consul.

## Flags

`-kube-config`: Path to kubernetes config file.

`consul-address`: The Consul Server address which is used for registering the services.
