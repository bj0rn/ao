package deletecmd

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/executil"
	"github.com/skatteetaten/ao/pkg/serverapi"
	"strings"
)

type DeletecmdClass struct {
	Configuration  *configuration.ConfigurationClass
	Force          bool
	deleteFileList []string
}

func (deletecmd *DeletecmdClass) DeleteApp(app string) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(deletecmd.Configuration)
	if err != nil {
		return err
	}

	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/"+app+".json") {
			err = deletecmd.addDeleteFileWithPrompt(filename, "Delete file "+filename)
			if err != nil {
				return err
			}

			// Check if no other app in folder this app was deleted in.  If no more exists, then delete the env file about.json
			var parts []string = strings.Split(filename, "/")
			var env = parts[0]
			var otherAppDeployedInEnv bool = false
			for appfile := range auroraConfig.Files {
				if strings.Contains(appfile, env+"/") && !strings.Contains(appfile, "/about.json") {
					if !deletecmd.isFileDeleted(appfile) {
						otherAppDeployedInEnv = true
						break
					}
				}
			}

			if !otherAppDeployedInEnv {
				var aboutFile = env + "/about.json"
				err = deletecmd.addDeleteFileWithPrompt(aboutFile, "No other deployments in "+env+" exists, delete about file "+aboutFile)
				if err != nil {
					return err
				}
			}

		}
	}

	// Delete the root app file
	var rootAppFile string = app + ".json"
	err = deletecmd.addDeleteFileWithPrompt(rootAppFile, "Delete file "+rootAppFile)

	// Delete all files in list and update aurora config in boober
	err = deletecmd.deleteFilesInList(auroraConfig)
	if err != nil {
		return err
	}

	return
}

func (deletecmd *DeletecmdClass) DeleteEnv(env string) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(deletecmd.Configuration)
	if err != nil {
		return err
	}

	// Delete all files in the folder
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, env+"/") {
			var parts []string = strings.Split(filename, "/")
			var app = parts[1]

			err = deletecmd.addDeleteFileWithPrompt(filename, "Delete file "+filename)
			if err != nil {
				return err
			}
			// If not about.json then check if this app is deployed in another env.  If not, then delete the root app.json also
			var deployAppFoundInOtherEnv bool = false
			if filename != env+"/about.json" {
				for appfile := range auroraConfig.Files {
					if strings.Contains(appfile, "/"+app) {
						// Check if file is marked for deletion, then we will not mark as found
						if !deletecmd.isFileDeleted(appfile) {
							deployAppFoundInOtherEnv = true
							break
						}
					}
				}
				if !deployAppFoundInOtherEnv {
					var rootFileName = app
					err = deletecmd.addDeleteFileWithPrompt(rootFileName, "No other deployment of "+rootFileName+" exists, delete root file "+rootFileName)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// Delete all files in list and update aurora config in boober
	err = deletecmd.deleteFilesInList(auroraConfig)
	if err != nil {
		return err
	}

	return
}

func (deletecmd *DeletecmdClass) DeleteDeployment(env string, app string) (err error) {

	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(deletecmd.Configuration)
	if err != nil {
		return err
	}

	deploymentFilename := env + "/" + app + ".json"
	_, deploymentExists := auroraConfig.Files[deploymentFilename]
	if !deploymentExists {
		if !deletecmd.Force {
			err = errors.New("No such deployment")
		}
		return err
	}

	err = deletecmd.addDeleteFileWithPrompt(deploymentFilename, "Delete file "+deploymentFilename)
	if err != nil {
		return err
	}

	// Check for existence of app.json in other envs, if none found then delete the root file as well
	var deployAppFoundInOtherEnv bool = false
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, "/"+app+".json") && filename != deploymentFilename {
			deployAppFoundInOtherEnv = true
			break
		}
	}

	if !deployAppFoundInOtherEnv {
		var rootFileName = app + ".json"
		err = deletecmd.addDeleteFileWithPrompt(rootFileName, "No other deployment of "+app+" exists, delete root file "+rootFileName)
		if err != nil {
			return err
		}
	}

	// Check for existence of other app.json in the same folder, if none found then delete the about.json as well
	var otherAppInSameFolder bool = false
	for filename := range auroraConfig.Files {
		if strings.Contains(filename, env+"/") && strings.Contains(filename, ".json") && filename != deploymentFilename && filename != env+"/about.json" {
			otherAppInSameFolder = true
			break
		}
	}

	if !otherAppInSameFolder {
		var envAboutFilename string = env + "/about.json"
		err = deletecmd.addDeleteFileWithPrompt(envAboutFilename, "No other apps exists in the env "+env+", delete environment file "+envAboutFilename)
		if err != nil {
			return err
		}
	}

	// Delete all files in list and update aurora config in boober
	err = deletecmd.deleteFilesInList(auroraConfig)
	if err != nil {
		return err
	}

	return
}

func (deletecmd *DeletecmdClass) DeleteFile(filename string) (err error) {
	// Get current aurora config
	auroraConfig, err := auroraconfig.GetAuroraConfig(deletecmd.Configuration)
	if err != nil {
		return err
	}

	_, deploymentExists := auroraConfig.Files[filename]
	if !deploymentExists {
		if !deletecmd.Force {
			err = errors.New("No such file")
		}
		return err
	}

	deletecmd.addDeleteFile(filename)
	// Delete all files in list and update aurora config in boober
	err = deletecmd.deleteFilesInList(auroraConfig)
	if err != nil {
		return err
	}

	return
}

func (deletecmd *DeletecmdClass) addDeleteFile(filename string) {
	deletecmd.deleteFileList = append(deletecmd.deleteFileList, filename)
}

func (deletecmd *DeletecmdClass) isFileDeleted(filename string) bool {
	for i := range deletecmd.deleteFileList {
		if deletecmd.deleteFileList[i] == filename {
			return true
		}
	}
	return false
}

func (deletecmd *DeletecmdClass) deleteFilesInList(auroraConfig serverapi.AuroraConfig) error {
	// Delete all files in list
	for i := range deletecmd.deleteFileList {
		delete(auroraConfig.Files, deletecmd.deleteFileList[i])
	}
	return auroraconfig.PutAuroraConfig(auroraConfig, deletecmd.Configuration)
}

func (deletecmd *DeletecmdClass) DeleteVault(vaultName string) (err error) {
	_, err = auroraconfig.DeleteVault(vaultName, deletecmd.Configuration)
	return err
}

func (deletecmd *DeletecmdClass) addDeleteFileWithPrompt(filename string, prompt string) (err error) {
	if deletecmd.Force {
		deletecmd.addDeleteFile(filename)
	} else {
		confirm, err := executil.PromptYNC(prompt)
		if err != nil {
			return err
		}
		if confirm == "Y" {
			deletecmd.addDeleteFile(filename)
		}
		if confirm == "C" {
			err = errors.New("Operation cancelled by user")
			return err
		}
	}
	return
}
