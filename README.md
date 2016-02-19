# Docker Machine OpenNebula Driver

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing to create docker hosts on [OpenNebula](http://www.opennebula.org)

## Authors

* Marco Mancini (@km4rcus)
* Jaime Melis (@jmelis)

## Requirements

* [Docker Machine](https://docs.docker.com/machine/) 0.5+
* [OpenNebula](http://www.opennebula.org) 4.x+

### Installation
Make sure [Go](http://www.golang.org) and [Godep](https://github.com/tools/godep) are properly installed, including setting up a [GOPATH](http://golang.org/doc/code.html#GOPATH).

To build the plugin binary:

```bash
$ go get github.com/OpenNebula/docker-machine-opennebula
$ cd $GOPATH/src/github.com/OpenNebula/docker-machine-opennebula
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

It is also necessary to specify the image to use by specifying `--opennebula-image-id` or `--opennebula-image-name` (and optionally `opennebula-image-owner`).

List of Options:

* `--opennebula-cpu`: CPU value for the VM
* `--opennebula-dev-prefix`: Dev prefix to use for the images: 'vd', 'sd', 'hd', etc...
* `--opennebula-disk-size`: Size of disk for VM in MB
* `--opennebula-image-id`: Image ID to use as the OS
* `--opennebula-image-name`: Image to use as the OS
* `--opennebula-image-owner`: Owner of the image to use as the OS
* `--opennebula-memory`: Size of memory for VM in MB
* `--opennebula-network-id`: Network ID to connect the machine to
* `--opennebula-network-name`: Network to connect the machine to
* `--opennebula-network-`: User ID of the Network to connect the machine to
* `--opennebula-ssh-user`: Set the name of the SSH user
* `--opennebula-vcpu`: VCPUs for the VM

|          CLI Option          | Default Value |  Environment Variable  |
|------------------------------|---------------|------------------------|
| `--opennebula-cpu`           | `1`           | `ONE_CPU`              |
| `--opennebula-dev-prefix`    |               | `ONE_IMAGE_DEV_PREFIX` |
| `--opennebula-disk-size`     |               | `ONE_DISK_SIZE`        |
| `--opennebula-image-id`      |               | `ONE_IMAGE_ID`         |
| `--opennebula-image-name`    |               | `ONE_IMAGE_NAME`       |
| `--opennebula-image-owner`   |               | `ONE_IMAGE_OWNER`      |
| `--opennebula-memory`        | `1024`        | `ONE_MEMORY`           |
| `--opennebula-network-id`    |               | `ONE_NETWORK_ID`       |
| `--opennebula-network-name`  |               | `ONE_NETWORK_NAME`     |
| `--opennebula-network-owner` |               | `ONE_NETWORK_OWNER`    |
| `--opennebula-ssh-user`      | `docker`      | `ONE_SSH_USER`         |
| `--opennebula-vcpu`          | `1`           | `ONE_VCPU`             |
