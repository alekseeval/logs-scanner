package http

type namespaceRequestStruct struct {
	Namespace string `json:"namespace"`
}

type clusterConfigRequestStruct struct {
	Config string `json:"config"`
}
