package opennebula

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/OpenNebula/goca"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	NetworkName    string
	NetworkOwner   string
	NetworkId      string
	ImageName      string
	ImageOwner     string
	ImageId        string
	CPU            string
	VCPU           string
	Memory         string
	DiskSize       string
	ImageDevPrefix string
}

const (
	defaultTimeout = 1 * time.Second
	defaultSSHUser = "docker"
	defaultCPU     = "1"
	defaultVCPU    = "1"
	defaultMemory  = "1024"

	contextScript = `#!/bin/bash

USERNAME=docker
GROUPNAME=docker
USER_HOME=/var/lib/docker

groupadd $GROUPNAME
useradd -m -d $USER_HOME -g $USERNAME $GROUPNAME

AUTH_DIR="${USER_HOME}/.ssh"
AUTH_FILE="${AUTH_DIR}/authorized_keys"

mkdir -m0700 -p $AUTH_DIR

echo "$DOCKER_SSH_PUBLIC_KEY" >> $AUTH_FILE

chown "${USERNAME}": ${AUTH_DIR} ${AUTH_FILE}
chmod 600 $AUTH_FILE

echo 'Defaults:$USERNAME !requiretty' >> /etc/sudoers.d/$USERNAME
echo "$USERNAME ALL=(ALL:ALL) NOPASSWD: ALL" >> /etc/sudoers.d/$USERNAME
`
)

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "opennebula-memory",
			Usage:  "Size of memory for VM in MB",
			EnvVar: "ONE_MEMORY",
			Value:  defaultMemory,
		},
		mcnflag.StringFlag{
			Name:   "opennebula-cpu",
			Usage:  "CPU value for the VM",
			EnvVar: "ONE_CPU",
			Value:  defaultCPU,
		},
		mcnflag.StringFlag{
			Name:   "opennebula-ssh-user",
			Usage:  "Set the name of the SSH user",
			EnvVar: "ONE_SSH_USER",
			Value:  defaultSSHUser,
		},
		mcnflag.StringFlag{
			Name:   "opennebula-vcpu",
			Usage:  "VCPUs for the VM",
			EnvVar: "ONE_VCPU",
			Value:  defaultVCPU,
		},
		mcnflag.StringFlag{
			Name:   "opennebula-disk-size",
			Usage:  "Size of disk for VM in MB",
			EnvVar: "ONE_DISK_SIZE",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-network-name",
			Usage:  "Network to connect the machine to",
			EnvVar: "ONE_NETWORK_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-network-id",
			Usage:  "Network ID to connect the machine to",
			EnvVar: "ONE_NETWORK_ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-network-owner",
			Usage:  "User ID of the Network to connect the machine to",
			EnvVar: "ONE_NETWORK_OWNER",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-image-name",
			Usage:  "Image to use as the OS",
			EnvVar: "ONE_IMAGE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-image-id",
			Usage:  "Image ID to use as the OS",
			EnvVar: "ONE_IMAGE_ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-image-owner",
			Usage:  "Owner of the image to use as the OS",
			EnvVar: "ONE_IMAGE_OWNER",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-dev-prefix",
			Usage:  "Dev prefix to use for the images: 'vd', 'sd', 'hd', etc...",
			EnvVar: "ONE_IMAGE_DEV_PREFIX",
			Value:  "",
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.CPU = flags.String("opennebula-cpu")
	d.VCPU = flags.String("opennebula-vcpu")
	d.Memory = flags.String("opennebula-memory")
	d.DiskSize = flags.String("opennebula-disk-size")
	d.NetworkName = flags.String("opennebula-network-name")
	d.NetworkId = flags.String("opennebula-network-id")
	d.NetworkOwner = flags.String("opennebula-network-owner")
	d.ImageId = flags.String("opennebula-image-id")
	d.ImageName = flags.String("opennebula-image-name")
	d.ImageOwner = flags.String("opennebula-image-owner")
	d.SSHUser = flags.String("opennebula-ssh-user")
	d.ImageDevPrefix = flags.String("opennebula-dev-prefix")

	if d.NetworkName == "" && d.NetworkId == "" {
		return errors.New("Please specify a network to connect to with --opennebula-network-name or --opennebula-network-id.")
	}

	if d.NetworkName != "" && d.NetworkId != "" {
		return errors.New("Please specify only one of: --opennebula-network-name or --opennebula-network-id, not both.")
	}

	if d.ImageName == "" && d.ImageId == "" {
		return errors.New("Please specify a image to use as the OS with --opennebula-image-name or --opennebula-image-id.")
	}

	if d.ImageName != "" && d.ImageId != "" {
		return errors.New("Please specify only one of: --opennebula-image-name or --opennebula-image-id, not both.")
	}

	return nil
}

func (d *Driver) DriverName() string {
	return "opennebula"
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	var err error

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	// Create template
	template := goca.NewTemplateBuilder()

	template.AddValue("NAME", d.MachineName)
	template.AddValue("CPU", d.CPU)
	template.AddValue("MEMORY", d.Memory)

	if d.VCPU != "" {
		template.AddValue("VCPU", d.VCPU)
	}

	vector := template.NewVector("NIC")
	if d.NetworkName != "" {
		vector.AddValue("NETWORK", d.NetworkName)
		if d.NetworkOwner != "" {
			vector.AddValue("NETWORK_UNAME", d.NetworkOwner)
		}
	}

	if d.NetworkId != "" {
		vector.AddValue("NETWORK_ID", d.NetworkId)
	}

	vector = template.NewVector("DISK")

	if d.ImageId != "" {
		vector.AddValue("IMAGE_ID", d.ImageId)
	} else {
		vector.AddValue("IMAGE", d.ImageName)
		if d.ImageOwner != "" {
			vector.AddValue("IMAGE_UNAME", d.ImageOwner)
		}
	}

	if d.DiskSize != "" {
		vector.AddValue("SIZE", d.DiskSize)
	}

	if d.ImageDevPrefix != "" {
		vector.AddValue("DEV_PREFIX", d.ImageDevPrefix)
	}

	vector = template.NewVector("CONTEXT")
	vector.AddValue("NETWORK", "YES")
	vector.AddValue("SSH_PUBLIC_KEY", "$USER[SSH_PUBLIC_KEY]")
	vector.AddValue("DOCKER_SSH_PUBLIC_KEY", string(pubKey))
	contextScript64 := base64.StdEncoding.EncodeToString([]byte(contextScript))
	vector.AddValue("START_SCRIPT_BASE64", contextScript64)

	vector = template.NewVector("GRAPHICS")
	vector.AddValue("LISTEN", "0.0.0.0")
	vector.AddValue("TYPE", "vnc")

	// Instantiate
	log.Infof("Starting  VM...")
	_, err = goca.CreateVM(template.String(), false)
	if err != nil {
		return err
	}

	if d.IPAddress, err = d.GetIP(); err != nil {
		return err
	}

	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return "", err
	}

	err = vm.Info()
	if err != nil {
		return "", err
	}

	if ip, ok := vm.XPath("/VM/TEMPLATE/NIC/IP"); ok {
		d.IPAddress = ip
	}

	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}

	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return state.None, err
	}

	err = vm.Info()
	if err != nil {
		return state.None, err
	}

	vm_state, lcm_state, err := vm.StateString()
	if err != nil {
		return state.None, err
	}

	switch vm_state {
	case "INIT", "PENDING", "HOLD":
		return state.Starting, nil
	case "ACTIVE":
		switch lcm_state {
		case "RUNNING",
			"DISK_SNAPSHOT",
			"DISK_SNAPSHOT_REVERT",
			"DISK_SNAPSHOT_DELETE",
			"HOTPLUG",
			"HOTPLUG_SNAPSHOT",
			"HOTPLUG_NIC",
			"HOTPLUG_SAVEAS":
			return state.Running, nil
		case "PROLOG",
			"BOOT",
			"MIGRATE",
			"PROLOG_MIGRATE",
			"PROLOG_RESUME",
			"CLEANUP_RESUBMIT",
			"BOOT_UNKNOWN",
			"BOOT_POWEROFF",
			"BOOT_SUSPENDED",
			"BOOT_STOPPED",
			"PROLOG_UNDEPLOY",
			"BOOT_UNDEPLOY",
			"BOOT_MIGRATE",
			"PROLOG_MIGRATE_SUSPEND",
			"SAVE_MIGRATE":
			return state.Starting, nil
		case "HOTPLUG_SAVEAS_POWEROFF",
			"DISK_SNAPSHOT_POWEROFF",
			"DISK_SNAPSHOT_REVERT_POWEROFF",
			"DISK_SNAPSHOT_DELETE_POWEROFF",
			"HOTPLUG_PROLOG_POWEROFF",
			"HOTPLUG_EPILOG_POWEROFF",
			"PROLOG_MIGRATE_POWEROFF",
			"SAVE_STOP":
			return state.Stopped, nil
		case "HOTPLUG_SAVEAS_SUSPENDED",
			"DISK_SNAPSHOT_SUSPENDED",
			"DISK_SNAPSHOT_REVERT_SUSPENDED",
			"DISK_SNAPSHOT_DELETE_SUSPENDED":
			return state.Saved, nil
		case "EPILOG_STOP",
			"EPILOG",
			"SHUTDOWN_UNDEPLOY",
			"EPILOG_UNDEPLOY",
			"SAVE_SUSPEND",
			"SHUTDOWN",
			"SHUTDOWN_POWEROFF",
			"CANCEL",
			"CLEANUP_DELETE":
			return state.Stopping, nil
		case "UNKNOWN",
			"FAILURE",
			"BOOT_FAILURE",
			"BOOT_MIGRATE_FAILURE",
			"PROLOG_MIGRATE_FAILURE",
			"PROLOG_FAILURE",
			"EPILOG_FAILURE",
			"EPILOG_STOP_FAILURE",
			"EPILOG_UNDEPLOY_FAILURE",
			"PROLOG_MIGRATE_POWEROFF_FAILURE",
			"PROLOG_MIGRATE_SUSPEND_FAILURE",
			"BOOT_UNDEPLOY_FAILURE",
			"BOOT_STOPPED_FAILURE",
			"PROLOG_RESUME_FAILURE",
			"PROLOG_UNDEPLOY_FAILURE":
			return state.Error, nil
		}
	case "POWEROFF", "UNDEPLOYED":
		return state.Stopped, nil
	case "STOPPED", "SUSPENDED":
		return state.Saved, nil
	case "DONE", "FAILED":
		return state.Error, nil
	}

	return state.Error, nil
}

func (d *Driver) Start() error {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return err
	}

	vm.Resume()

	s := state.None
	for retry := 0; retry < 50 && s != state.Running; retry++ {
		s, err = d.GetState()
		if err != nil {
			return err
		}

		switch s {
		case state.Running:
		case state.Error:
			return errors.New("VM in error state")
		default:
			time.Sleep(2 * time.Second)
		}
	}

	if d.IPAddress == "" {
		if d.IPAddress, err = d.GetIP(); err != nil {
			return err
		}
	}

	log.Infof("Waiting for SSH...")
	// Wait for SSH over NAT to be available before returning to user
	if err := drivers.WaitForSSH(d); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return err
	}

	err = vm.PowerOff()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Remove() error {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return err
	}

	err = vm.ShutdownHard()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return err
	}

	err = vm.Reboot()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return err
	}

	err = vm.PowerOffHard()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}
