package lib

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/client"
)

func ProcessClient(connection net.Conn, dockerFileDirectory string, listOfDockerFiles []string) error {

	for {
		for {
			connection.Write([]byte("Select Operation: \n1. Create Container\n2. View Containers\n3. Admin Console\n4. Exit\n>> "))
			userSelection, err := bufio.NewReader(connection).ReadString('\n')
			if err != nil {
				fmt.Printf("%d: Connection Closed: %v by %s\n", time.Now().Unix(), err, connection.RemoteAddr())
				return err
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
					return err
				}

				userSelection, userSelErr := strconv.Atoi(strings.TrimSuffix(message, "\n"))

				if userSelErr == nil && (userSelection < len(listOfDockerFiles) || userSelection <= 0) {
					fmt.Println(time.Now().Unix(), ":", listOfDockerFiles[userSelection], "spinning up by remote host", connection.RemoteAddr())

					client, err := client.NewEnvClient()
					if err != nil {
						fmt.Printf("Unable to create docker client: %s", err)
						return err
					}

					tags := []string{listOfDockerFiles[userSelection]}
					dockerfile := dockerFileDirectory + listOfDockerFiles[userSelection]
					containerID, err := BuildImage(client, tags, dockerfile)
					if err != nil {
						log.Println(err)
						continue
					}
					imagename := listOfDockerFiles[userSelection]
					containername := containerID
					portopening := "6379"
					inputEnv := []string{fmt.Sprintf("LISTENINGPORT=%s", portopening)}
					err = RunContainer(client, imagename, containername, portopening, inputEnv)
					if err != nil {
						log.Println(err)
						return err
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
				fmt.Printf("%d: Connection closed by %s\n", time.Now().Unix(), connection.RemoteAddr())
				connection.Close()
				return nil
			} else {
				connection.Write([]byte("Invalid selection. Please enter one of the numbers from the list above\n"))
			}
		}
	}
	connection.Close()
	return nil
}
