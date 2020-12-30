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
func NewClusterCfgWithCM(kubeinterface kubernetes.Interface, namespace string, label map[string]string, dataname string, options ...Option) ClusterConfigurationManager {
	return &ClusterCfgWithCM{
		kubeinterface: kubeinterface,
		namespace:     namespace,
		label:         label,
		dataname:      dataname,
		options:       options,
	}
}

// GetAll get all cluster info
func (cc *ClusterCfgWithCM) GetAll() ([]ClusterConfigInfo, error) {
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
		return nil, fmt.Errorf("get configmap with namespace:%s label:%+v failed:%+v", cc.namespace, cc.label, err)
	}

	clsInfoList := []ClusterConfigInfo{}
	for _, cm := range cmlist.Items {
		if kubecfg, ok := cm.Data[cc.dataname]; ok {
			clsInfoList = append(clsInfoList, BuildClusterInfo(cm.Name, kubecfg, "", KubeConfigTypeRawString))
		}
	}
	return clsInfoList, nil
}

// GetOptions get options
func (cc *ClusterCfgWithCM) GetOptions() []Option {
	return cc.options
}

// ClusterCfgWithDir cluster info read from dir
type ClusterCfgWithDir struct {
	kubeinterface  kubernetes.Interface
	dir            string
	suffix         string
	options        []Option
	kubeConfigType KubeConfigType
}

// NewClusterCfgWithDir build clusterconfigdir
func NewClusterCfgWithDir(kubeinterface kubernetes.Interface, dir, suffix string, kubeConfigType KubeConfigType, options ...Option) (ClusterConfigurationManager, error) {
	s, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("%s is not exist or other err:%+v", dir, err)
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("%s is not directory", dir)
	}

	return &ClusterCfgWithDir{
		kubeinterface:  kubeinterface,
		dir:            dir,
		suffix:         suffix,
		options:        options,
		kubeConfigType: kubeConfigType,
	}, nil
}

// GetAll get all cluster info
func (cd *ClusterCfgWithDir) GetAll() ([]ClusterConfigInfo, error) {
	files, err := ioutil.ReadDir(cd.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s failed:%+v", cd.dir, err)
	}

	clsInfoList := []ClusterConfigInfo{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), cd.suffix) {
			path := cd.dir + "/" + file.Name()

			if cd.kubeConfigType == KubeConfigTypeFile {
				clsInfoList = append(clsInfoList, BuildClusterInfo(file.Name(), path, "", KubeConfigTypeFile))
			} else {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("read file %s failed:%+v", path, err)
				}
				clsInfoList = append(clsInfoList, BuildClusterInfo(file.Name(), string(data), "", KubeConfigTypeRawString))
			}

		}
	}
	return clsInfoList, nil
}

// GetOptions options
func (cd *ClusterCfgWithDir) GetOptions() []Option {
	return cd.options
}

// ClusterInfo cluster info
type ClusterInfo struct {
	name           string
	kubeConfig     string
	kubeContext    string
	kubeConfigType KubeConfigType
}

// BuildClusterInfo build cluster info
func BuildClusterInfo(name, kubeConfig, kubeContext string, kubeConfigType KubeConfigType) ClusterConfigInfo {
	return &ClusterInfo{
		name:           name,
		kubeConfig:     kubeConfig,
		kubeContext:    kubeContext,
		kubeConfigType: kubeConfigType,
	}
}

// GetName get cluster name
func (ci *ClusterInfo) GetName() string {
	return ci.name
}

// GetKubeConfig get kubernetes config data
func (ci *ClusterInfo) GetKubeConfig() string {
	return ci.kubeConfig
}

// GetKubeContext get kubeconfig current context
func (ci *ClusterInfo) GetKubeContext() string {
	return ci.kubeContext
}

func (ci *ClusterInfo) GetKubeConfigType() KubeConfigType {
	return ci.kubeConfigType
}
