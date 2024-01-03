package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	natting "github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	network "github.com/docker/docker/api/types/network"
)

// declared here for global scope
// probably should just pass in the list of docker files as an extra paramater to the processClient func
var dockerFileDirectory string = "images/"
var listOfDockerFiles []string

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func buildImage(client *client.Client, tags []string, dockerfile string) error {
	ctx := context.Background()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFileReader, err := os.Open(dockerfile)
	if err != nil {
		return err
	}

	readDockerFile, err := ioutil.ReadAll(dockerFileReader)
	if err != nil {
		return err
	}

	tarHeader := &tar.Header{
		Name: dockerfile,
		Size: int64(len(readDockerFile)),
	}

	err = tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}

	_, err = tw.Write(readDockerFile)
	if err != nil {
		return err
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	buildOptions := types.ImageBuildOptions{
		Context:    dockerFileTarReader,
		Dockerfile: dockerfile,
		Remove:     true,
		Tags:       tags,
	}

	imageBuildResponse, err := client.ImageBuild(
		ctx,
		dockerFileTarReader,
		buildOptions,
	)

	if err != nil {
		return err
	}

	defer imageBuildResponse.Body.Close()

	//maybe switch this up to be an output of the function not print
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)

	if err != nil {
		return err
	}

	return nil
}

func runContainer(client *client.Client, imagename string containername string, port string, inputEnv []string) error {
	
		newport, err := natting.NewPort("tcp", port)
		if err != nil {
				fmt.Println("Unable to create docker port")
				return err
		}

		hostConfig := &container.HostConfig{
				PortBindings: natting.PortMap{
						newport: []natting.PortBinding{
								{
										HostIP: "0.0.0.0",
										HostPort: port,
								},
						},
				},
				RestartPolicy: container.RestartPolicy{
						Name: "always",
				},
				LogConfig: container.LogConfig{
						Type: "json-file",
						Config: map[string]string{},
				},
		}

		networkConfig := &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{},
		}
		gatewayConfig := &network.EndpointSettings{
				Gateway: "gatewayname",
		}
		networkConfig.EndpointsConfig["bridge"] = gatewayConfig

		exposedPorts := map[natting.Port}struct{}{
				newport: struct{}{},
		}

		config := &container.Config{
				Image: 		imagename,
				Env:		inputEnv,
				ExposedPorts: exposedPorts,
				Hostname: 	fmt.Sprintf("%s-hostnameexample", imagename),
		}

		cont, err := client.ContainerCreate(
				context.Background(),
				config,
				hostConfig,
				networkConfig,
				containername,
		)

		if err != nil {
				log.Println(err)
				return err
		}

}

func processClient(connection net.Conn) int {

	for {
		for {
			connection.Write([]byte("Select Operation: \n1. Create Container\n2. View Containers\n3. Admin Console\n4. Exit\n>> "))
			userSelection, err := bufio.NewReader(connection).ReadString('\n')
			if err != nil {
				fmt.Printf("%d: Connection Closed: %v by %s\n", time.Now().Unix(), err, connection.RemoteAddr())
				return 0
			}

			if userSelection == "1\n" {

				connection.Write([]byte("\nSelect a dockerfile to execute: \n"))

				for i, j := range listOfDockerFiles {
					connection.Write([]byte(strconv.Itoa(i) + ": " + j + "\n"))
				}

				connection.Write([]byte("\n>> "))

				message, err := bufio.NewReader(connection).ReadString('\n')

				if err != nil {
					fmt.Printf("%d: Connection Closed: %v by %s\n", time.Now().Unix(), err, connection.RemoteAddr())
					return 0
				}

				userSelection, userSelErr := strconv.Atoi(strings.TrimSuffix(message, "\n"))

				if userSelErr == nil && (userSelection < len(listOfDockerFiles) || userSelection <= 0) {
					fmt.Println(time.Now().Unix(), ":", listOfDockerFiles[userSelection], "spinning up by remote host", connection.RemoteAddr())

					client, err := client.NewEnvClient()
					if err != nil {
						log.Fatalf("Unable to create docker client: %s", err)
					}

					tags := []string{"tags"}
					dockerfile := dockerFileDirectory + listOfDockerFiles[userSelection]
					err = buildImage(client, tags, dockerfile)
					if err != nil {
						log.Println(err)
						continue
					}

					connection.Write([]byte("Finished building image from " + listOfDockerFiles[userSelection] + "\n"))
				} else {
					connection.Write([]byte("Invalid selection. Please enter an integer off the list above\n"))
					continue
				}
			} else if userSelection == "2\n" {
				connection.Write([]byte("This menu is still under development\n"))
			} else if userSelection == "3\n" {
				connection.Write([]byte("This menu is still under development\n"))
				fmt.Printf("%d: Connection Closed by %s\n", time.Now().Unix(), connection.RemoteAddr())
			} else if userSelection == "4\n" {
				fmt.Printf("%d: Connection closed by %s\n", time.Now().Unix(), err, connection.RemoteAddr())
				connection.Close()
				return 0
			} else {
				connection.Write([]byte("Invalid selection. Please enter one of the numbers from the list above\n"))
			}
		}
	}
	connection.Close()
	return 0
}

func main() {

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
		go processClient(connection)
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
