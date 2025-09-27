package device

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
)

const execTimeout = 1 * time.Second

type Device struct {
	PlatformName    string
	PlatformVersion string
	UUID            uuid.UUID
	DeviceType      string
}

func New() (*Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()
	osInfo, err := getOSInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing device: %w", err)
	}
	uniqUUID, errUniqUUID := deviceUUID()
	if errUniqUUID != nil {
		return nil, fmt.Errorf("initializing device: %w", errUniqUUID)
	}

	return &Device{
		PlatformName:    osInfo.Name,
		PlatformVersion: osInfo.Version,
		UUID:            uniqUUID,
		DeviceType:      "cli",
	}, nil
}

// OSInfo contains information about the operating system.
type OSInfo struct {
	Name    string
	Version string
}

// getOSInfo returns information about the operating system.
func getOSInfo(ctx context.Context) (*OSInfo, error) {
	var name, version string
	var err error

	switch runtime.GOOS {
	case "windows":
		cmd := exec.CommandContext(ctx, "cmd", "/C", "ver")
		var output bytes.Buffer
		cmd.Stdout = &output
		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("getting OS information: %w", err)
		}
		name = "Windows"
		version = strings.TrimSpace(output.String())

	case "darwin":
		cmdName := exec.CommandContext(ctx, "sw_vers", "-productName")
		cmdVersion := exec.CommandContext(ctx, "sw_vers", "-productVersion")

		var nameOut, versionOut bytes.Buffer
		cmdName.Stdout = &nameOut
		cmdVersion.Stdout = &versionOut

		err = cmdName.Run()
		if err != nil {
			return nil, fmt.Errorf("getting OS information: %w", err)
		}
		err = cmdVersion.Run()
		if err != nil {
			return nil, fmt.Errorf("getting OS information: %w", err)
		}

		name = strings.TrimSpace(nameOut.String())
		version = strings.TrimSpace(versionOut.String())

	case "linux":
		cmd := exec.CommandContext(ctx, "sh", "-c", "cat /etc/os-release | grep PRETTY_NAME | cut -d '=' -f2 | tr -d '\"'")
		var output bytes.Buffer
		cmd.Stdout = &output

		err = cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("getting OS information: %w", err)
		}

		fullName := strings.TrimSpace(output.String())
		segments := strings.SplitN(fullName, " ", 2) //nolint:mnd
		if len(segments) > 0 {
			name = segments[0]
		}
		if len(segments) > 1 {
			version = segments[1]
		}

	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return &OSInfo{
		Name:    name,
		Version: version,
	}, nil
}

func deviceUUID() (uuid.UUID, error) {
	hashedID, err := machineid.ProtectedID("gophkeeper")
	if err != nil {
		return uuid.Nil, fmt.Errorf("getting device uuid: %w", err)
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(hashedID)), nil
}
