# Docker Machine OpenNebula Driver

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing to create docker hosts on [OpenNebula](http://www.opennebula.org)

## Requirements
* [Docker Machine](https://docs.docker.com/machine/) 0.5+
* [OpenNebula](http://www.opennebula.org) 4.x+

### Installation 
Make sure [Go](http://www.golang.org) and [Godep](https://github.com/tools/godep) are properly installed, including setting up a [GOPATH](http://golang.org/doc/code.html#GOPATH). 

To build the plugin binary:

```bash
$ go get github.com/km4rcus/docker-machine-opennebula
$ cd $GOPATH/src/github.com/km4rcus/docker-machine-opennebula
$ make build
```
After the build is complete, `bin/docker-machine-driver-opennebula` binary will be created and it must included in `$PATH` variable. If you want to copy it to the `${GOPATH}/bin/`, run `make install`.

## Usage
Official documentation for Docker Machine [is available here](https://docs.docker.com/machine/).

Set up `ONE_AUTH` and `ONE_XMLRPC` to point to the OpenNebula cloud; then to create a docker machine with OpenNebula plugin you can run this command (assuming an existent virtual network named 'private'):

```
$ docker-machine create --driver opennebula --opennebula-network private one-boot2d 
```

## Available Driver Options

It is required to specify the network the machine will be connected to with `--opennebula-network-name` or `--opennebula-network-id`; in case `--opennebula-network-name` is used then the owner of the network can be passed with `--opennebula-network-owner` if it is different from the user in the file `ONE_AUTH`.

 - `--opennebula-network-name` or `--opennebula-network-id`: Identify the network the machine will be connected to
 - `--opennebula-network-owner`: Owner of the network the machine will be connected to
 - `--opennebula-boot2docker-url`: The url of boot2docker image with [Docker](http://www.docker.com) 1.9 installed and OpenNebula context packages
 - `--opennebula-disk-size`: Size of disk for host in MB
 - `--opennebula-memory`: Size of memory for VM in MB.
 - `--opennebula-cpu`: CPU value for the VM
 - `--opennebula-vcpu`: VCPUs for the VM
 - `--opennebula-datastore-id`: Datastore ID for saving Boot2Docker image 
 - `--opennebula-ssh-user`: Set the name of the SSH user  


Environment variables and default values:

| CLI option                     | Environment variable  | Default  value                          | Required       | 
|--------------------------------|-----------------------|-----------------------------------------|----------------|
| `--opennebula-network-name`    | `ONE_NETWORK_NAME`    | No                                      |  Yes           |
| `--opennebula-network-owner`   | `ONE_NETWORK_OWNER`   | No                                      |  No            |
| `--opennebula-network-id`      | `ONE_NETWORK_ID`      | No                                      |  Yes           |
| `--opennebula-boot2docker-url` | `ONE_BOOT2DOCKER_URL` | https://s3.eu-central-1.amazonaws.com/one-boot2d/boot2docker-v1.9.1.iso |  No            |
| `--opennebula-cpu`             | `ONE_CPU`             | `1`                                     |  No            |
| `--opennebula-vcpu`            | `ONE_VCPU`            | `1`                                     |  No            |
| `--opennebula-disk-size`       | `ONE_DISK_SIZE`       | `20000 MB`                              |  No            |
| `--opennebula-datastore-id`    | `ONE_DATASTORE_ID`    | `1`                                     |  No            |
| `--opennebula-memory`          | `ONE_MEMORY`          | `1024 MB`                               |  No            |
| `--opennebula-ssh-user`        | `ONE_SSH_USER`        | `docker`                                |  No            |



