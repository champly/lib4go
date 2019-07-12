package jenkins

import (
	"errors"
	"fmt"

	"github.com/bndr/gojenkins"
)

var (
	JobNotExist = errors.New("job not exist")
)

type JenkinsClient struct {
	client *gojenkins.Jenkins
}

func NewJenkinsClient(host, account, pwd string) (client *JenkinsClient, err error) {
	cli := gojenkins.CreateJenkins(nil, host, account, pwd)
	_, err = cli.Info()
	if err != nil {
		return nil, err
	}
	return &JenkinsClient{client: cli}, nil
}

func (j *JenkinsClient) ModifyJob(jobName string, config string) error {
	job, err := j.client.GetJob(jobName)
	if err != nil {
		if err.Error() == "404" {
			return j.CreateJob(jobName, config)
		}
		return err
	}

	return job.UpdateConfig(config)
}

func (j *JenkinsClient) CreateJob(jobName string, config string) error {
	job, err := j.client.CreateJob(config, jobName)
	if err != nil {
		return err
	}
	fmt.Println(job.GetName())
	return nil
}

func (j *JenkinsClient) Build(jobName string, params map[string]string) (queueID int64, err error) {
	job, err := j.client.GetJob(jobName)
	if err != nil {
		if err.Error() == "404" {
			return 0, JobNotExist
		}
		return 0, err
	}

	queueID, err = job.InvokeSimple(params)
	if err != nil {
		return 0, err
	}
	return queueID, nil
}

func (j *JenkinsClient) GetBuildID(queueID int64) (buildID int64, result string) {
	task, err := j.client.GetQueueItem(queueID)
	if err != nil {
		result = err.Error()
		return
	}

	buildID = task.Raw.Executable.Number
	result = task.GetWhy()
	return
}

func (j *JenkinsClient) GetLog(jobName string, buildID int64) (string, error) {
	job, err := j.client.GetJob(jobName)
	if err != nil {
		if err.Error() == "404" {
			return "", JobNotExist
		}
		return "", err
	}

	build, err := job.GetBuild(buildID)
	if err != nil {
		return "", err
	}
	return build.GetConsoleOutput(), nil
}

func (j *JenkinsClient) IsComplate(jobName string, buildID int64) (result string, err error) {

	build, err := j.client.GetBuild(jobName, buildID)
	if err != nil {
		return result, err
	}

	//FAILURE/SUCCESS
	result = build.GetResult()
	return
}

func (j *JenkinsClient) GetAllBuildList(jobName string) ([]BuildInfo, error) {

	job, err := j.client.GetJob(jobName)
	if err != nil {
		if err.Error() == "404" {
			return nil, JobNotExist
		}
		return nil, err
	}

	bList, err := job.GetAllBuildIds()
	if err != nil {
		return nil, err
	}

	list := []BuildInfo{}
	for _, info := range bList {
		list = append(list, BuildInfo{
			QueueID:  info.QueueID,
			BuildID:  info.Number,
			Result:   info.Result,
			Building: info.Building,
		})
	}
	return list, nil
}
