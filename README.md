todo:
- [ ] spin up docker image based on docker file
- [ ] change variable for command line flags
- [ ] paramterize login 
- [ ] paramterize ports (ie ssh is on 2243 for the server)
- [ ] create admin pannel/login
    - [ ] allow for admin to view all containers and their uptime
    - [ ] allow for admin to stop  one container
    - [ ] allow for admin to stop all containers at once
- [x] track container uptime 
- [ ] create help menu (for connected users)
- [x] automatically kill containers that have an uptime over x hours (function)
- [x] allow users to extend their uptime
- [x] create menu for all of this (1: help 2: create containers 3: view containers/uptime 4: admin login)

also rn all this is logged to the stdout. im thinking about creating a log file but idk yet.

if there is a docker api version error just export the DOCKER_API_VERSION= to valid version

if there is a error Create container with PortBindings (error: “cannot use type nat.PortMap as type nat.PortMap”) then deleting the folder “$GOPATH/src/github.com/docker/docker/vendor/github.com/docker/go-connections”
https://forums.docker.com/t/docker-sdk-go-create-container-with-portbindings-error-cannot-use-type-nat-portmap-as-type-nat-portmap/83959
