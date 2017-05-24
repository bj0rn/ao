package getcmd

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/auroraconfig"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
)

const UsageString = "Usage: aoc get files | file [env/]<filename> | adc"
const filesUsageString = "Usage: aoc get files"
const fileUseageString = "Usage: aoc get file [env/]<filename>"
const adcUsageString = "Usage: aoc get adc"
const notYetImplemented = "Not supported yet"

type GetcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (getcmdClass *GetcmdClass) getAffiliation() (affiliation string) {
	if getcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = getcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (getcmdClass *GetcmdClass) getFiles(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	var files []string
	files, err = auroraconfig.GetFileList(persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())

	output = "NAME"
	for fileindex := range files {
		output += "\n" + files[fileindex]
	}
	return
}

func (getcmdClass *GetcmdClass) getFile(filename string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string) (output string, err error) {

	switch outputFormat {
	case "json":
		{
			content, err := auroraconfig.GetContent(filename, persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())
			if err != nil {
				return "", err
			}
			output = jsonutil.PrettyPrintJson(content)
			return output, err
		}
	case "":
		{
			var files []string
			files, err = auroraconfig.GetFileList(persistentOptions, getcmdClass.getAffiliation(), getcmdClass.configuration.GetOpenshiftConfig())
			output = "NAME"
			var fileFound bool
			for fileindex := range files {
				if files[fileindex] == filename {
					output += "\n" + files[fileindex]
					fileFound = true
				}
			}
			if !fileFound {
				err = errors.New("Error: file \"" + filename + "\" not found")
				return "", err
			}
			return output, nil
		}
	case "yaml":
		{
			err = errors.New(notYetImplemented)
			return "", err
		}
	default:
		{
			err = errors.New("Illegal format.  Legal values are json, yaml.")
		}
	}

	return
}

func (getcmdClass *GetcmdClass) getAdc(persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	output += notYetImplemented
	return
}

func (getcmdClass *GetcmdClass) getClusters(persistentOptions *cmdoptions.CommonCommandOptions, clusterName string, allClusters bool) (output string, err error) {
	var displayClusterName string
	const tab = " "

	openshiftConfig := getcmdClass.configuration.GetOpenshiftConfig()
	output = "CLUSTER NAME         REACHABLE  API  URL"
	for i := range openshiftConfig.Clusters {
		if openshiftConfig.Clusters[i].Reachable || allClusters {
			displayClusterName = openshiftConfig.Clusters[i].Name
			if displayClusterName == clusterName || clusterName == "" {
				var apiColumn = fileutil.RightPad("", 4)
				if clusterName == openshiftConfig.APICluster {
					apiColumn = fileutil.RightPad("Yes", 4)
				}
				var reachableColumn = fileutil.RightPad("", 10)
				if openshiftConfig.Clusters[i].Reachable {
					reachableColumn = fileutil.RightPad("Yes", 10)
				}
				var urlColumn = ""
				displayClusterName = fileutil.RightPad(displayClusterName, 20)
				urlColumn = openshiftConfig.Clusters[i].Url
				output += "\n" + displayClusterName + tab + reachableColumn + tab + apiColumn + tab + urlColumn
			}
		}

	}

	return
}

func (getcmdClass *GetcmdClass) GetObject(args []string, persistentOptions *cmdoptions.CommonCommandOptions, outputFormat string, allClusters bool) (output string, err error) {
	err = validateEditcmd(args)
	if err != nil {
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "files":
		{
			output, err = getcmdClass.getFiles(persistentOptions)
		}
	case "file":
		{
			output, err = getcmdClass.getFile(args[1], persistentOptions, outputFormat)
		}
	case "adc":
		{
			output, err = getcmdClass.getAdc(persistentOptions)
		}
	case "cluster", "clusters":
		{
			var clusterName = ""
			if len(args) > 1 {
				clusterName = args[1]
			}
			output, err = getcmdClass.getClusters(persistentOptions, clusterName, allClusters)
		}
	}

	return
}

func validateEditcmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(UsageString)
		return
	}

	var commandStr = args[0]
	switch commandStr {
	case "files":
		{
			if len(args) > 1 {
				err = errors.New(filesUsageString)
				return
			}
		}
	case "file":
		{
			if len(args) != 2 {
				err = errors.New(fileUseageString)
				return
			}
		}
	case "adc":
		{
			if len(args) > 1 {
				err = errors.New(adcUsageString)
				return
			}
		}
	default:
		{
			return
		}

	}

	return
}
