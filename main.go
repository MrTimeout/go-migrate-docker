package main

func main() {
	NewRootCmd(&DockerHost{}).Execute() //nolint:errcheck
}
