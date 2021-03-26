package helm

import (
	"io/ioutil"
	"os"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

func TestDownload(t *testing.T) {
	env := cli.New()
	cfg := &repo.Entry{
		Name: "stable",
		URL:  "http://mirror.azure.cn/kubernetes/charts/",
	}
	os.MkdirAll(env.RepositoryCache, os.ModePerm)

	// download repo index
	rp, err := repo.NewChartRepository(cfg, getter.All(env))
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("repo.CachePath:", rp.CachePath)
	index, err := rp.DownloadIndexFile()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("DownloadIndexFile:", index)

	// update repo
	repoFile := env.RepositoryCache + "/repositories.yaml"
	t.Log("repoFile:", repoFile)
	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		t.Error(err)
		return
	}
	var f repo.File
	err = yaml.Unmarshal(b, &f)
	if err != nil {
		t.Error(err)
		return
	}
	f.Update(cfg)
	t.Log(f.Repositories)
	err = f.WriteFile(repoFile, 0644)
	if err != nil {
		t.Error(err)
		return
	}

	dl := downloader.ChartDownloader{
		Out:              os.Stdout,
		RepositoryConfig: repoFile,
		RepositoryCache:  env.RepositoryCache,
		Getters:          getter.All(env),
	}
	t.Log(env.RepositoryConfig)
	t.Log(env.RepositoryCache)

	// a, f, err := dl.DownloadTo("stable/nginx-ingress", "1.41.3", "")
	os.MkdirAll("./Charts", 0755)
	a, p, err := dl.DownloadTo("stable/nginx-ingress", "1.41.3", "./Charts")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(p.FileName)
	t.Log(a)
}
