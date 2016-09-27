package interfaces

type Endpoint struct {
	ServiceName      string
	ServiceNamespace string
	DnsLabel         string
	NodeAddress      string
	NodeName         string
	NodePort         int32
}
