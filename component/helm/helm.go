package helm

import (
	"fmt"
	"io/ioutil"
	"os"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

type Options struct {
	RepositoryConfig string
}

type Client struct {
	env *cli.EnvSettings
}

func (c *Client) AddRepo(repoName, repoURL string) error {
	cfg := &repo.Entry{
		Name: repoName,
		URL:  repoURL,
	}
	rp, err := repo.NewChartRepository(cfg, getter.All(c.env))
	if err != nil {
		return fmt.Errorf("build chart repository error: %v", err)
	}

	// download index file
	_, err = rp.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("download index file error: %v", err)
	}

	// update repo to repositoryconfig
	repoFile := c.env.RepositoryConfig
	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read repo file %s error: %v", repoFile, err)
	}
	var f repo.File
	err = yaml.Unmarshal(b, &f)
	if err != nil {
		return fmt.Errorf("yaml %s unmarshal to repo.File failed: %v", string(b), err)
	}
	f.Update(cfg)
	err = f.WriteFile(repoFile, 0644)
	if err != nil {
		return fmt.Errorf("rewrite config to repoFile failed: %v", err)
	}
	return nil
}

func (c *Client) DownloadTo(repoName, chartName, version, dest string) (fileName string, err error) {
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return "", fmt.Errorf("mkall dir %s failed: %v", dest, err)
	}

	dl := downloader.ChartDownloader{
		Out:              os.Stdout,
		RepositoryConfig: c.env.RepositoryConfig,
		RepositoryCache:  c.env.RepositoryCache,
		Getters:          getter.All(c.env),
	}
	fn, _, err := dl.DownloadTo(fmt.Sprintf("%s/%s", repoName, chartName), version, dest)
	if err != nil {
		return "", fmt.Errorf("download chart %s/%s@%s to %s failed: %v", repoName, chartName, version, dest, err)
	}
	return fn, nil
}
