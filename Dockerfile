FROM busybox
ADD kube2consul /kube2consul
CMD ["/kube2consul"]
