package queue

type Task func() error

// Instance of work tickets processed using a rate-limiting loop
type Instance interface {
	// Push a task
	Push(task Task)
	// Run the loop until a signal on the channel
	Run(<-chan struct{})
}
