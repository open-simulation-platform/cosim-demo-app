package main

func main() {
	execution := createExecution()
	slaveIndex := executionAddfmu(execution, "C:/dev/osp/cse-core/test/data/fmi2/Clock.fmu")

	// Creating a command channel
	cmd := make(chan string, 10)
	state := make(chan string, 10)

	// Passing the channel to the go routine
	go simulate(execution, slaveIndex, cmd)

	//Passing the channel to the server
	Server(cmd, state)
	close(cmd)
}
