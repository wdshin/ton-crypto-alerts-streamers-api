package ton

type Dispatcher struct {
}

func (d *Dispatcher) Notify() {

	// ToDo: Read donations that were not acked for last 30 mins

	// ToDo: Dispatch event to notification service with a bulk of new donations
}
