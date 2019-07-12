package jenkins

type BuildInfo struct {
	BuildID  int64
	QueueID  int64
	Result   string
	Building bool
}
