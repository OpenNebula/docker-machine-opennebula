package opennebula

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/km4rcus/goca"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	TemplateName   string
	TemplateId     string
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
	B2DSize        string
        User	       string
        Password       string
        Xmlrpcurl      string
        Config         goca.OneConfig
	DisableVNC     bool
}

const (
	defaultTimeout = 1 * time.Second
	defaultSSHUser = "docker"
	defaultCPU     = "1"
	defaultVCPU    = "1"
	defaultMemory  = "1024"

	// This is the contextualization script that will be executed by OpenNebula
	contextScript = `#!/bin/sh

if [ -f /etc/boot2docker ]; then
	USERNAME=docker
	USER_HOME=/home/docker
else
	USERNAME=$DOCKER_SSH_USER
	GROUPNAME=$DOCKER_SSH_USER

	if ! getent group $GROUPNAME; then
		groupadd $GROUPNAME
	fi

	if ! getent passwd $USERNAME; then
		USER_HOME=/var/lib/$DOCKER_SSH_USER
		useradd -m -d $USER_HOME -g $USERNAME $GROUPNAME
	else
		USER_HOME=$(getent passwd $USERNAME | cut -d: -f 6)
	fi

	# Write sudoers
	if [ ! -f /etc/sudoers.d/$USERNAME ]; then
		echo -n "Defaults:$USERNAME " >> /etc/sudoers.d/$USERNAME
		echo '!requiretty' >> /etc/sudoers.d/$USERNAME
		echo "$USERNAME ALL=(ALL:ALL) NOPASSWD: ALL" >> /etc/sudoers.d/$USERNAME
	fi
fi

# Add DOCKER_SSH_PUBLIC_KEY

AUTH_DIR="${USER_HOME}/.ssh"
AUTH_FILE="${AUTH_DIR}/authorized_keys"

mkdir -m0700 -p $AUTH_DIR

echo "$DOCKER_SSH_PUBLIC_KEY" >> $AUTH_FILE

chown "${USERNAME}": ${AUTH_DIR} ${AUTH_FILE}
chmod 600 $AUTH_FILE`
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

func (d *Driver) buildConfig() {
        d.Config = goca.NewConfig(d.User, d.Password, d.Xmlrpcurl)
}

func (d *Driver) setClient() error {
        return goca.SetClient(d.Config)	
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "opennebula-cpu",
			Usage:  fmt.Sprintf("CPU value for the VM. Default: %d", defaultCPU),
			EnvVar: "ONE_CPU",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-vcpu",
			Usage:  fmt.Sprintf("VCPUs for the VM. Default: %d", defaultVCPU),
			EnvVar: "ONE_VCPU",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-memory",
			Usage:  fmt.Sprintf("Size of memory for VM in MB. Default: %d", defaultMemory),
			EnvVar: "ONE_MEMORY",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-template-name",
			Usage:  "Template to use",
			EnvVar: "ONE_TEMPLATE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-template-id",
			Usage:  "Template ID to use",
			EnvVar: "ONE_TEMPLATE_ID",
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
		mcnflag.StringFlag{
			Name:   "opennebula-disk-resize",
			Usage:  "Size of disk for VM in MB",
			EnvVar: "ONE_DISK_SIZE",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-b2d-size",
			Usage:  "Size of the Volatile disk in MB (only for b2d)",
			EnvVar: "ONE_B2D_DATA_SIZE",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "opennebula-ssh-user",
			Usage:  "Set the name of the SSH user",
			EnvVar: "ONE_SSH_USER",
			Value:  defaultSSHUser,
		},
		mcnflag.BoolFlag{
			Name:   "opennebula-disable-vnc",
			Usage:  "VNC is enabled by default. Disable it with this flag",
			EnvVar: "ONE_DISABLE_VNC",
		},
                mcnflag.StringFlag{
                        Name:   "opennebula-user",
                        Usage:  "Set the user for authentication",
                        EnvVar: "ONE_USER",
                },
                mcnflag.StringFlag{
                        Name:   "opennebula-password",
                        Usage:  "Set the password for authentication",
                        EnvVar: "ONE_PASSWORD",
                },
                mcnflag.StringFlag{
                        Name:   "opennebula-xmlrpcurl",
                        Usage:  "Set the url for one xmlrpc server",
                        EnvVar: "ONE_XMLRPC",
                },
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SetSwarmConfigFromFlags(flags)

        // Authentication
        d.User = flags.String("opennebula-user")
	d.Password = flags.String("opennebula-password")
        d.Xmlrpcurl = flags.String("opennebula-xmlrpcurl")

	// Capacity
	d.CPU = flags.String("opennebula-cpu")
	d.VCPU = flags.String("opennebula-vcpu")
	d.Memory = flags.String("opennebula-memory")

	// Template
	d.TemplateName = flags.String("opennebula-template-name")
	d.TemplateId = flags.String("opennebula-template-id")

	// Network
	d.NetworkName = flags.String("opennebula-network-name")
	d.NetworkId = flags.String("opennebula-network-id")
	d.NetworkOwner = flags.String("opennebula-network-owner")

	// Storage
	d.ImageId = flags.String("opennebula-image-id")
	d.ImageName = flags.String("opennebula-image-name")
	d.ImageOwner = flags.String("opennebula-image-owner")

	d.ImageDevPrefix = flags.String("opennebula-dev-prefix")
	d.DiskSize = flags.String("opennebula-disk-resize")
	d.B2DSize = flags.String("opennebula-b2d-size")

	// Provision
	d.SSHUser = flags.String("opennebula-ssh-user")

	// VNC
	d.DisableVNC = flags.Bool("opennebula-disable-vnc")

	// Either TemplateName or TemplateId
	if d.TemplateName != "" && d.TemplateId != "" {
		return errors.New("Please specify only one of: --opennebula-template-name or --opennebula-template-id, not both.")
	}

	// Either NetworkName or NetworkId
	if d.NetworkName != "" && d.NetworkId != "" {
		return errors.New("Please specify only one of: --opennebula-network-name or --opennebula-network-id, not both.")
	}

	// Either ImageName or ImageId
	if d.ImageName != "" && d.ImageId != "" {
		return errors.New("Please specify only one of: --opennebula-image-name or --opennebula-image-id, not both.")
	}

	// Required and incompatible options for Template
	if d.TemplateName != "" || d.TemplateId != "" {
		// Template has been specified:

		// ImageName and ImageId are incompatible
		if d.ImageName != "" || d.ImageId != "" {
			return errors.New("The options --opennebula-image-* are incompatible with --opennebula-template-*.")
		}

		// ImageDevPrefix is incompatible
		if d.ImageDevPrefix != "" {
			return errors.New("The option: --opennebula-dev-prefix is incompatible with --opennebula-template-*.")
		}
		// DiskSize is incompatible
		if d.DiskSize != "" {
			return errors.New("The option: --opennebula-disk-resize is incompatible with --opennebula-template-*.")
		}
		// B2DSize is incompatible
		if d.B2DSize != "" {
			return errors.New("The option: --opennebula-disk-resize is incompatible with --opennebula-template-*.")
		}
		// DisableVNC is incompatible
		if d.DisableVNC {
			return errors.New("The option: --opennebula-disable-vnc is incompatible with --opennebula-template-*.")
		}
	} else {
		//Template has NOT been specified:

		// ImageName or ImageId is required
		if d.ImageName == "" && d.ImageId == "" {
			return errors.New("Please specify a image to use as the OS with --opennebula-image-name or --opennebula-image-id.")
		}

		// NetworkName or NetworkId is required
		if d.NetworkName == "" && d.NetworkId == "" {
			return errors.New("Please specify a network to connect to with --opennebula-network-name or --opennebula-network-id.")
		}

		// Assign default capacity values
		if d.CPU == "" {
			d.CPU = defaultCPU
		}

		if d.VCPU == "" {
			d.VCPU = defaultVCPU
		}

		if d.Memory == "" {
			d.Memory = defaultMemory
		}
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
	var (
		vector     *goca.TemplateBuilderVector
		vmtemplate *goca.Template

		err error
	)

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

        // build config and set the xmlrpc client
	d.buildConfig()
	d.setClient()

	// Create template
	template := goca.NewTemplateBuilder()

	if d.TemplateName != "" || d.TemplateId != "" {
		// Template has been specified
	} else {
		// Template has NOT been specified

		template.AddValue("NAME", d.MachineName)

		// OS Disk
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

		// Add a volatile disk for b2d
		if d.B2DSize != "" {
			vector = template.NewVector("DISK")
			vector.AddValue("SIZE", d.B2DSize)
			vector.AddValue("TYPE", "fs")
			vector.AddValue("FORMAT", "raw")
		}

		// VNC
		if !d.DisableVNC {
			vector = template.NewVector("GRAPHICS")
			vector.AddValue("LISTEN", "0.0.0.0")
			vector.AddValue("TYPE", "vnc")
		}
	}

	// Capacity
	if d.CPU != "" {
		template.AddValue("CPU", d.CPU)
	}

	if d.Memory != "" {
		template.AddValue("MEMORY", d.Memory)
	}

	if d.VCPU != "" {
		template.AddValue("VCPU", d.VCPU)
	}

	// Network
	if d.NetworkName != "" || d.NetworkId != "" {
		vector = template.NewVector("NIC")

		if d.NetworkName != "" {
			vector.AddValue("NETWORK", d.NetworkName)
			if d.NetworkOwner != "" {
				vector.AddValue("NETWORK_UNAME", d.NetworkOwner)
			}
		}

		if d.NetworkId != "" {
			vector.AddValue("NETWORK_ID", d.NetworkId)
		}
	}

	// Context
	vector = template.NewVector("CONTEXT")
	vector.AddValue("NETWORK", "YES")
	vector.AddValue("SSH_PUBLIC_KEY", "$USER[SSH_PUBLIC_KEY]")
	vector.AddValue("DOCKER_SSH_USER", d.SSHUser)
	vector.AddValue("DOCKER_SSH_PUBLIC_KEY", string(pubKey))
	contextScript64 := base64.StdEncoding.EncodeToString([]byte(contextScript))
	vector.AddValue("START_SCRIPT_BASE64", contextScript64)

	// Instantiate
	log.Infof("Starting  VM...")

	if d.TemplateName != "" || d.TemplateId != "" {

		if d.TemplateName != "" {
			vmtemplate, err = goca.NewTemplateFromName(d.TemplateName)
			if err != nil {
				return err
			}
		} else {
			templateId, err := strconv.Atoi(d.TemplateId)
			if err != nil {
				return err
			}
			vmtemplate = goca.NewTemplate(uint(templateId))
		}

		_, err = vmtemplate.Instantiate(d.MachineName, false, template.String())

	} else {
		_, err = goca.CreateVM(template.String(), false)
	}

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
        d.setClient()
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
        d.setClient() 
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
        d.setClient()
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
        d.setClient()
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
        d.setClient()
	vm, err := goca.NewVMFromName(d.MachineName)
	if err != nil {
		return err
	}

	err = vm.ShutdownHard()
	if err != nil {
		err = vm.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) Restart() error {
        d.setClient()
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
        d.setClient()
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
