# Docker Machine OpenNebula Driver

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing to create docker hosts on [OpenNebula](http://www.opennebula.org)

## Authors

* Marco Mancini ([@km4rcus](https://github.com/km4rcus)
* Jaime Melis ([@jmelis](https://github.com/jmelis))

## Requirements

* [Docker Machine](https://docs.docker.com/machine/) 0.5+
* [OpenNebula](http://www.opennebula.org) 4.x+

## Installation

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

Set up `ONE_AUTH` and `ONE_XMLRPC` to point to the OpenNebula cloud

To use this plugin you will need to choose an OpenNebula image where docker will be installed. This image can be a vanilla OS [supported by docker-machine](https://github.com/docker/machine/blob/master/docs/drivers/os-base.md) or a specially built [boot2docker for OpenNebula](http://marketplace.opennebula.systems/appliance/56d073858fb81d0315000002). Using boo2docker for OpenNebula is recommended, as it will boot faster since it has all the required packages.


### Boot2Docker

Import [boot2docker for OpenNebula](http://marketplace.opennebula.systems/appliance/56d073858fb81d0315000002) into OpenNebula and then use the following arguments in docker-machine:

```bash
$ docker-machine create --driver opennebula --opennebula-network-name $NETWORK_NAME --opennebula-image-id $BOOT2DOCKER_IMG_ID --opennebula-data-size $DATA_SIZE_MB b2d
```

Remember to substitute:

* `$NETWORK_NAME` with a real network already existent in your OpenNebula installation
* `$BOOT2DOCKER_IMG_ID` the ID of the boot2docker image imported from the MarketPlace.
* `$DATA_SIZE_MB` is the size of the volatile disk that will be used to store the docker data.

### Vanilla OS

As long as you have a vanilla OS image in your OpenNebula installation, and this image is contextualized with the [latest packages](http://docs.opennebula.org/stable/user/virtual_machine_setup/bcont.html#preparing-the-virtual-machine-image), you can used it as such:

```bash
$ docker-machine create --driver opennebula --opennebula-network-name $NETWORK_NAME --opennebula-image-id $IMG_ID mydockerengine
```

As you can see in this case you don't need to specify the data size.


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
* `--opennebula-disable-vnc`: VNC is enabled by default. Disable it with this flag
* `--opennebula-data-size`: Size of the Volatile disk in MB (only for b2d)

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
| `--opennebula-disable-vnc`   | Enabled       | `ONE_DISABLE_VNC`      |
| `--opennebula-data-size`     |               | `ONE_B2D_DATA_SIZE`    |

Remember that:

* If you are using a regular vanilla OS image in OpenNebula you can use `--opennebula-disk-size` to resize the size of the OS, but you should never use `--opennebula-data-size` in this case. If you don't specify `--opennebula-disk-size`, the size of the disk will be the default one, the one of the image.
* If you are using boot2docker, you have to use `--opennebula-data-size`, in order to provision an extra data disk, but you should never use `--opennebula-disk-size` in this case.
