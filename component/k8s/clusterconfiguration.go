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

// clusterCfgManagerWithCM cluster info read from configmap
type clusterCfgManagerWithCM struct {
	kubeInterface kubernetes.Interface
	namespace     string
	label         map[string]string
	dataname      string
	options       []Option
}

// NewClusterCfgManagerWithCM build clusterconfigmap
func NewClusterCfgManagerWithCM(kubeInterface kubernetes.Interface, namespace string, label map[string]string, dataname string, options ...Option) ClusterConfigurationManager {
	return &clusterCfgManagerWithCM{
		kubeInterface: kubeInterface,
		namespace:     namespace,
		label:         label,
		dataname:      dataname,
		options:       options,
	}
}

// GetAll get all cluster info
func (cc *clusterCfgManagerWithCM) GetAll() ([]ClusterConfigInfo, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	labelSelectors := make([]string, 0, len(cc.label))
	for k, v := range cc.label {
		if k != "" && v != "" {
			labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmlist, err := cc.kubeInterface.CoreV1().ConfigMaps(cc.namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(labelSelectors, ",")})
	if err != nil {
		return nil, fmt.Errorf("get configmap with namespace:%s label:%+v failed:%+v", cc.namespace, cc.label, err)
	}

	clsInfoList := []ClusterConfigInfo{}
	for _, cm := range cmlist.Items {
		if kubecfg, ok := cm.Data[cc.dataname]; ok {
			clsInfoList = append(clsInfoList, BuildClusterCfgInfo(cm.Name, kubecfg, "", KubeConfigTypeRawString))
		}
	}
	return clsInfoList, nil
}

// GetOptions get options
func (cc *clusterCfgManagerWithCM) GetOptions() []Option {
	return cc.options
}

// clusterCfgManagerWithDir cluster info read from dir
type clusterCfgManagerWithDir struct {
	dir            string
	suffix         string
	options        []Option
	kubeConfigType KubeConfigType
}

// NewClusterCfgManagerWithDir build clusterconfigdir
func NewClusterCfgManagerWithDir(dir, suffix string, kubeConfigType KubeConfigType, options ...Option) (ClusterConfigurationManager, error) {
	s, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("%s is not exist or other err:%+v", dir, err)
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("%s is not directory", dir)
	}

	return &clusterCfgManagerWithDir{
		dir:            dir,
		suffix:         suffix,
		options:        options,
		kubeConfigType: kubeConfigType,
	}, nil
}

// GetAll get all cluster info
func (cd *clusterCfgManagerWithDir) GetAll() ([]ClusterConfigInfo, error) {
	files, err := ioutil.ReadDir(cd.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s failed:%+v", cd.dir, err)
	}

	clsInfoList := []ClusterConfigInfo{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), cd.suffix) {
			path := cd.dir + "/" + file.Name()

			if cd.kubeConfigType == KubeConfigTypeFile {
				clsInfoList = append(clsInfoList, BuildClusterCfgInfo(file.Name(), path, "", KubeConfigTypeFile))
			} else {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("read file %s failed:%+v", path, err)
				}
				clsInfoList = append(clsInfoList, BuildClusterCfgInfo(file.Name(), string(data), "", KubeConfigTypeRawString))
			}

		}
	}
	return clsInfoList, nil
}

// GetOptions options
func (cd *clusterCfgManagerWithDir) GetOptions() []Option {
	return cd.options
}

// clusterCfgInfo cluster info
type clusterCfgInfo struct {
	name           string
	kubeConfig     string
	kubeContext    string
	kubeConfigType KubeConfigType
}

// BuildClusterCfgInfo build cluster info
func BuildClusterCfgInfo(name, kubeConfig, kubeContext string, kubeConfigType KubeConfigType) ClusterConfigInfo {
	return &clusterCfgInfo{
		name:           name,
		kubeConfig:     kubeConfig,
		kubeContext:    kubeContext,
		kubeConfigType: kubeConfigType,
	}
}

// GetName get cluster name
func (ci *clusterCfgInfo) GetName() string {
	return ci.name
}

// GetKubeConfig get kubernetes config data
func (ci *clusterCfgInfo) GetKubeConfig() string {
	return ci.kubeConfig
}

// GetKubeContext get kubeconfig current context
func (ci *clusterCfgInfo) GetKubeContext() string {
	return ci.kubeContext
}

func (ci *clusterCfgInfo) GetKubeConfigType() KubeConfigType {
	return ci.kubeConfigType
}
