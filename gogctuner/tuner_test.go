package gctuner

// idea is from this article https://eng.uber.com/how-we-saved-70k-cores-across-30-mission-critical-services
// how to use this lib?
func initProcess() {
	var (
		inCgroup = true
		percent  = 70
	)
	go NewTuner(inCgroup, float64(percent))
}
