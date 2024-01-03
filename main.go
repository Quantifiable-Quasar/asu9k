package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	L "./lib"

	"github.com/docker/docker/client"
)

// declared here for global scope
// probably should just pass in the list of docker files as an extra paramater to the processClient func

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {

	var dockerFileDirectory string = "images/"
	var listOfDockerFiles []string

	files, err := ioutil.ReadDir(dockerFileDirectory)
	check(err)

	for _, i := range files {
		listOfDockerFiles = append(listOfDockerFiles, i.Name())
	}

	server, err := net.Listen("tcp", "127.0.0.1:9999")
	check(err)
	fmt.Println(time.Now().Unix(), ":", "Server Starting...")

	defer server.Close()

	for {
		connection, err := server.Accept()
		check(err)

		fmt.Println(time.Now().Unix(), ":", "Client connection from", connection.RemoteAddr())
		go L.ProcessClient(connection, dockerFileDirectory, listOfDockerFiles)
	}
}

/* from here on out will be testing for docker, expect mistates and poorly written code below */

// Gets uptime from a given containter
func getContainerUptime(containerID string) (time.Duration, error) {

	cli, err := client.NewEnvClient()
	check(err)

	containerInspect, err := cli.ContainerInspect(context.Background(), containerID)
	check(err)

	createdTime, err := time.Parse(time.RFC3339Nano, containerInspect.Created)
	check(err)

	uptime := time.Since(createdTime)
	return uptime, nil
}

func uptimeLimit(containerID string, timeLimit time.Duration) error {
	uptime, err := getContainerUptime(containerID)
	if err != nil {
		return err
	}

	if uptime > timeLimit {
		fmt.Printf("Container %s has exceeded the time limit. Killing...\n", containerID)

		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		// Set a timeout for killing the container (i will try and impliment later)
		//timeout := 10 * time.Second

		err = cli.ContainerKill(context.Background(), containerID, "SIGKILL")
		if err != nil {
			return err
		}

		fmt.Printf("Container %s killed.\n", containerID)
	} else {
		fmt.Printf("Container %s uptime is within the limit.\n", containerID)
	}

	return nil
}

/*
   For allowing users to extend time, each user will have to have their
   own timeLimit varible, not sure how this can be handeled (possibly via database or by container)
*/

// The function below is proably not functional, but no real way to test until we have sample docker containters
/*
func extendContainerTimeLimit(containerID string, additionalTime time.Duration) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// Get the current container information
	containerInspect, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return err
	}

	// Calculate the new timeout by adding the additional time to the current uptime
	createdTime, err := time.Parse(time.RFC3339Nano, containerInspect.Created)
	if err != nil {
		return err
	}
	currentUptime := time.Since(createdTime)
	newTimeout := currentUptime + additionalTime

	// Update the container's timeout
	timeout := int(newTimeout.Seconds())
	containerInspect, err = cli.ContainerUpdate(context.Background(), containerID, types.ContainerUpdateConfig{BlkioWeight: 0, CPUShares: 0, CgroupParent: "", OomKillDisable: false, Timeout: &timeout})
	if err != nil {
		return err
	}

	fmt.Printf("Container %s time limit extended by %s.\n", containerID, additionalTime.String())
	return nil

}
*/
