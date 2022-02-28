package standalone

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"errors"
	"fmt"
	"os"

	"github.com/bhojpur/application/pkg/utils"
)

func removeContainers(uninstallPlacementContainer, uninstallAll bool, dockerNetwork string) []error {
	var containerErrs []error
	var err error

	if uninstallPlacementContainer {
		containerErrs = removeDockerContainer(containerErrs, AppPlacementContainerName, dockerNetwork)

		_, err = utils.RunCmdAndWait(
			"docker", "rmi",
			"--force",
			appDockerImageName)

		if err != nil {
			containerErrs = append(
				containerErrs,
				fmt.Errorf("could not remove %s image: %s", appDockerImageName, err))
		}
	}

	if uninstallAll {
		containerErrs = removeDockerContainer(containerErrs, AppRedisContainerName, dockerNetwork)
		containerErrs = removeDockerContainer(containerErrs, AppZipkinContainerName, dockerNetwork)
	}

	return containerErrs
}

func removeDockerContainer(containerErrs []error, containerName, network string) []error {
	container := utils.CreateContainerName(containerName, network)
	exists, _ := confirmContainerIsRunningOrExists(container, false)
	if !exists {
		utils.WarningStatusEvent(os.Stdout, "WARNING: %s container does not exist", container)
		return containerErrs
	}
	utils.InfoStatusEvent(os.Stdout, "Removing container: %s", container)
	_, err := utils.RunCmdAndWait(
		"docker", "rm",
		"--force",
		container)
	if err != nil {
		containerErrs = append(
			containerErrs,
			fmt.Errorf("could not remove %s container: %s", container, err))
	}
	return containerErrs
}

func removeDir(dirPath string) error {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		utils.WarningStatusEvent(os.Stdout, "WARNING: %s does not exist", dirPath)
		return nil
	}
	utils.InfoStatusEvent(os.Stdout, "Removing directory: %s", dirPath)
	err = os.RemoveAll(dirPath)
	return err
}

// Uninstall reverts all changes made by init. Deletes all installed containers, removes
// default Bhojpur Application folder, removes the installed binary and unsets env variables.
func Uninstall(uninstallAll bool, dockerNetwork string) error {
	var containerErrs []error
	appDefaultDir := defaultAppDirPath()
	appBinDir := defaultAppBinPath()

	placementFilePath := binaryFilePath(appBinDir, placementServiceFilePrefix)
	_, placementErr := os.Stat(placementFilePath) // check if the placement binary exists
	uninstallPlacementContainer := os.IsNotExist(placementErr)

	// Remove .bhojpur/bin
	err := removeDir(appBinDir)
	if err != nil {
		utils.WarningStatusEvent(os.Stdout, "WARNING: could not delete Bhojpur Application bin dir: %s", appBinDir)
	}

	dockerInstalled := false
	dockerInstalled = utils.IsDockerInstalled()
	if dockerInstalled {
		containerErrs = removeContainers(uninstallPlacementContainer, uninstallAll, dockerNetwork)
	}

	if uninstallAll {
		err = removeDir(appDefaultDir)
		if err != nil {
			utils.WarningStatusEvent(os.Stdout, "WARNING: could not delete default Bhojpur Application dir: %s", appDefaultDir)
		}
	}

	err = errors.New("uninstall failed")
	if uninstallPlacementContainer && !dockerInstalled {
		// if placement binary did not exist before trying to delete it and not able to connect to docker.
		return fmt.Errorf("%w \ncould not delete placement service. Either the placement binary is not found, or Docker may not be installed or running", err)
	}

	if len(containerErrs) == 0 {
		return nil
	}

	for _, e := range containerErrs {
		err = fmt.Errorf("%w \n %s", err, e)
	}
	return err
}
