package k8s

// ClusterConfigmap cluster configuration with configmap
type ClusterConfigmap struct {
}

func (cc *ClusterConfigmap) GetAll() []ClusterInfo {
	return nil
}

type ClusterInfoConfigmap struct {
	Clsname    string
	Kubeconfig string
	Option     []Option
}

func NewClusterInfoWithString(clsname, data string) ClusterInfo {
	return nil
}
