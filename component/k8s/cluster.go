package k8s

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// ClusterCfgWithCM cluster info read from configmap
type ClusterCfgWithCM struct {
	kubeinterface kubernetes.Interface
	namespace     string
	label         map[string]string
	dataname      string
	options       []Option
}

// NewClusterCfgWithCM build clusterconfigmap
func NewClusterCfgWithCM(kubeinterface kubernetes.Interface, namespace string, label map[string]string, dataname string, options ...Option) IClusterConfiguration {
	return &ClusterCfgWithCM{
		kubeinterface: kubeinterface,
		namespace:     namespace,
		label:         label,
		dataname:      dataname,
		options:       options,
	}
}

// GetAll get all cluster info
func (cc *ClusterCfgWithCM) GetAll() []IClusterInfo {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	labelSelectors := make([]string, 0, len(cc.label))
	for k, v := range cc.label {
		if k != "" && v != "" {
			labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmlist, err := cc.kubeinterface.CoreV1().ConfigMaps(cc.namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(labelSelectors, ",")})
	if err != nil {
		klog.Errorf("get configmap with namespace:%s label:%+v failed:%+v", cc.namespace, cc.label, err)
		return nil
	}

	clsInfoList := []IClusterInfo{}
	for _, cm := range cmlist.Items {
		if kubecfg, ok := cm.Data[cc.dataname]; ok {
			clsInfoList = append(clsInfoList, BuildClusterInfo(cm.Name, kubecfg))
		}
	}
	return clsInfoList
}

// GetOptions get options
func (cc *ClusterCfgWithCM) GetOptions() []Option {
	return cc.options
}

// ClusterCfgWithDir cluster info read from dir
type ClusterCfgWithDir struct {
	kubeinterface kubernetes.Interface
	dir           string
	suffix        string
	options       []Option
}

// NewClusterCfgWithDir build clusterconfigdir
func NewClusterCfgWithDir(kubeinterface kubernetes.Interface, dir, suffix string, options ...Option) (IClusterConfiguration, error) {
	s, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("%s is not exist or other err:%+v", dir, err)
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("%s is not directory", dir)
	}

	return &ClusterCfgWithDir{
		kubeinterface: kubeinterface,
		dir:           dir,
		suffix:        suffix,
		options:       options,
	}, nil
}

// GetAll get all cluster info
func (cd *ClusterCfgWithDir) GetAll() []IClusterInfo {
	files, err := ioutil.ReadDir(cd.dir)
	if err != nil {
		klog.Errorf("read dir %s failed:%+v", cd.dir, err)
		return nil
	}

	clsInfoList := []IClusterInfo{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), cd.suffix) {
			path := cd.dir + "/" + file.Name()
			data, err := ioutil.ReadFile(path)
			if err != nil {
				klog.Errorf("read file %s failed:%+v", path, err)
				return nil
			}
			clsInfoList = append(clsInfoList, BuildClusterInfo(file.Name(), string(data)))
		}
	}
	return clsInfoList
}

// GetOptions options
func (cd *ClusterCfgWithDir) GetOptions() []Option {
	return cd.options
}

// ClusterInfo cluster info
type ClusterInfo struct {
	clsname    string
	kubeconfig string
}

// BuildClusterInfo build cluster info
func BuildClusterInfo(clsname, kubeconfig string) IClusterInfo {
	return &ClusterInfo{
		clsname:    clsname,
		kubeconfig: kubeconfig,
	}
}

// GetName get cluster name
func (ci *ClusterInfo) GetName() string {
	return ci.clsname
}

// GetKubeConfig get kubernetes config data
func (ci *ClusterInfo) GetKubeConfig() string {
	return ci.kubeconfig
}
