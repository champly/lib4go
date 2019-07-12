package jenkins

import (
	"fmt"
	"testing"
	"time"
)

func TestNewJenksinClient(t *testing.T) {
	client, err := NewJenkinsClient("http://10.13.3.6:8080", "admin", "123456")
	if err != nil {
		t.Log(err.Error())
		return
	}
	// t.Log(client.ModifyJob("demo1", config))
	jobName := "demo"
	queueID, err := client.Build(jobName, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("queue_id:", queueID)
	var buildID int64
	var result string

	for buildID == 0 {

		buildID, result = client.GetBuildID(queueID)
		fmt.Println(result)
		time.Sleep(time.Second * 1)
	}
	t.Log("build_id:", buildID)

	compl, err := client.IsComplate(jobName, buildID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(compl)
	time.Sleep(time.Second * 1)
	t.Log(client.GetLog(jobName, buildID))
}

var config = `<?xml version='1.0' encoding='UTF-8'?>
<project>
  <actions/>
  <description>1123221</description>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.scm.NullSCM"/>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders/>
  <publishers/>
  <buildWrappers/>
</project>
`

func TestMotify(t *testing.T) {
	client, err := NewJenkinsClient("http://10.13.3.6:8080", "admin", "123456")
	if err != nil {
		t.Log(err.Error())
		return
	}
	// t.Log(client.ModifyJob("demo1", config))
	jobName := "demo11"
	client.ModifyJob(jobName, config)
}
