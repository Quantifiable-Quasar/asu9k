package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
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

func processClient(connection net.Conn) int {

	for {
		for {
			connection.Write([]byte("\nSelect Operation: \n1. Create Container\n2. View Containers\n3. Admin Console\n4. Exit\n>>"))
			userSelection, err := bufio.NewReader(connection).ReadString('\n')
			if err != nil {
				fmt.Printf("%d: Connection Closed: %v by %s\n", time.Now().Unix(), err, connection.RemoteAddr())
				break
			}

			if userSelection == "1\n" {

				connection.Write([]byte("\nSelect a dockerfile to execute: \n>>"))

				for i, j := range listOfDockerFiles {
					connection.Write([]byte(strconv.Itoa(i) + ": " + j + "\n"))
				}

				connection.Write([]byte("\n"))

				message, err := bufio.NewReader(connection).ReadString('\n')

				if err != nil {
					fmt.Printf("%d: Connection Closed: %v by %s\n", time.Now().Unix(), err, connection.RemoteAddr())
					break
				}

				userSelection, userSelErr := strconv.Atoi(strings.TrimSuffix(message, "\n"))

				if userSelErr == nil && (userSelection < len(listOfDockerFiles) || userSelection <= 0) {
					fmt.Println(time.Now().Unix(), ":", listOfDockerFiles[userSelection], "spinning up by remote host", connection.RemoteAddr())
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

func getDockerTime() {

	// empty for now

}
