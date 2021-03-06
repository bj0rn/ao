package fuzzyargs

import (
	"strconv"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

/*
Module to create a list of apps and envs based upon user parameters from the command line.
The parameters are mathced based upon the file and folder names in the AuroraConfig

Init()							- Reads the AuroraConfig
PopulateFuzzyEnvAppList()		- Parses the args array given


Two modes: Expect env/app combinations, or a single file name
Clients will do:
- Init(configuration)					This will read the boober configuration and populate the
										legal<Type>list arrays in the FuzzyArgs struct
- PopulateFuzzyEnvAppList(args)			This will populate the <Type>list arrays
- PopulateFuzzyFileList(args)			This will populate the FileList array, including the about entries

Arguments types:

<short env>/<short app>					Identify an app and an environment OR a file
<short env>
<short app>
<short "about.json">					Only applicable on files
<short env>/<short "about.json">		Only applicable on files


*/

import (
	"errors"
	"strings"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/printutil"
)

type FuzzyArgs struct {
	configuration *configuration.ConfigurationClass
	appList       []string
	envList       []string
	filename      string
	legalAppList  []string
	legalEnvList  []string
	legalFileList []string
}

// ** Initialize **
func (fuzzyArgs *FuzzyArgs) Init(configuration *configuration.ConfigurationClass) (err error) {
	fuzzyArgs.configuration = configuration
	err = fuzzyArgs.getLegalEnvAppFileList()
	if err != nil {
		return err
	}
	return
}

// Try to match an argument with an app, returns "" if none found
func (fuzzyArgs *FuzzyArgs) GetFuzzyApp(arg string) (app string, err error) {
	if strings.HasSuffix(arg, ".json") {
		arg = strings.TrimSuffix(arg, ".json")
	}
	// First check for exact match
	for i := range fuzzyArgs.legalAppList {
		if fuzzyArgs.legalAppList[i] == arg {
			return arg, nil
		}
	}
	// No exact match found, look for an app name that contains the string
	for i := range fuzzyArgs.legalAppList {
		if strings.Contains(fuzzyArgs.legalAppList[i], arg) {
			if app != "" {
				err = errors.New(arg + ": Not a unique application identifier, matching " + app + " and " + fuzzyArgs.legalAppList[i])
				return "", err
			}
			app = fuzzyArgs.legalAppList[i]
		}
	}

	return app, nil
}

// Try to match an argument with an env, returns "" if none found
func (fuzzyArgs *FuzzyArgs) GetFuzzyEnv(arg string) (env string, err error) {
	// First check for exact match
	for i := range fuzzyArgs.legalEnvList {
		if fuzzyArgs.legalEnvList[i] == arg {
			return arg, nil
		}
	}
	// No exact match found, look for an env name that contains the string
	for i := range fuzzyArgs.legalEnvList {
		if strings.Contains(fuzzyArgs.legalEnvList[i], arg) {
			if env != "" {
				err = errors.New(arg + ": Not a unique environment identifier, matching both " + env + " and " + fuzzyArgs.legalEnvList[i])
				return "", err
			}
			env = fuzzyArgs.legalEnvList[i]
		}
	}

	return env, nil
}

func (fuzzyArgs *FuzzyArgs) getLegalEnvAppFileList() (err error) {

	auroraConfig, err := auroraconfig.GetAuroraConfig(fuzzyArgs.configuration)
	if err != nil {
		return err
	}
	for filename := range auroraConfig.Files {
		fuzzyArgs.addLegalFile(filename)
		if strings.Contains(filename, "/") {
			// We have a full path name
			parts := strings.Split(filename, "/")
			fuzzyArgs.addLegalEnv(parts[0])
			if !strings.Contains(parts[1], "about.json") {
				if strings.HasSuffix(parts[1], ".json") {
					fuzzyArgs.addLegalApp(strings.TrimSuffix(parts[1], ".json"))
				}

			}
		}
	}

	return
}

// Parse args, expect one or two args that describes a file
func (fuzzyArgs *FuzzyArgs) PopulateFuzzyFile(args []string) (err error) {

	if len(args) == 1 {
		if strings.Contains(args[0], "/") {
			// We have a full path name with a slash, split it and call ourselves recursively
			parts := strings.Split(args[0], "/")
			return fuzzyArgs.PopulateFuzzyFile(parts)
		}
		// This should be a root file, search through the root file list
		var found bool = false
		for i := range fuzzyArgs.legalFileList {
			if !strings.Contains(fuzzyArgs.legalFileList[i], "/") {
				if strings.Contains(fuzzyArgs.legalFileList[i], args[0]) {
					if found {
						err = errors.New("Duplicate file spec found: " + args[0] + " matching both " + fuzzyArgs.filename + " and " + fuzzyArgs.legalFileList[i])
						return err
					}
					found = true
					fuzzyArgs.filename = fuzzyArgs.legalFileList[i]
				}
			}
		}
		if found {
			return nil
		}
	} else if len(args) == 2 {
		// This is a file in an environment catalog
		// Find the env and then check if there is a file in this env
		var foundEnv bool = false
		var env string = ""
		// First check exact match
		for i := range fuzzyArgs.legalEnvList {
			if fuzzyArgs.legalEnvList[i] == args[0] {
				foundEnv = true
				env = fuzzyArgs.legalEnvList[i]
			}
		}
		if !foundEnv {
			// Check fuzzy match
			for i := range fuzzyArgs.legalEnvList {
				if strings.Contains(fuzzyArgs.legalEnvList[i], args[0]) {
					if foundEnv {
						err = errors.New("Duplicate environment spec found: " + args[0] + " matching both " + env + " and " + fuzzyArgs.legalEnvList[i])
						return err
					}
					foundEnv = true
					env = fuzzyArgs.legalEnvList[i]
				}
			}
		}

		if !foundEnv {
			err = errors.New("No matching env found")
			return err
		}
		// Try to find the file in the found env
		var foundFile bool = false
		// First check exact match
		for i := range fuzzyArgs.legalFileList {
			if fuzzyArgs.legalFileList[i] == env+"/"+args[1] {
				foundFile = true
				fuzzyArgs.filename = fuzzyArgs.legalFileList[i]
			}
		}
		if !foundFile {
			for i := range fuzzyArgs.legalFileList {
				if strings.Contains(fuzzyArgs.legalFileList[i], env+"/"+args[1]) {
					if foundFile {
						err = errors.New("Duplicate file spec found: " + args[1] + " matching both " + fuzzyArgs.filename + " and " + fuzzyArgs.legalFileList[i])
						return err
					}
					foundFile = true
					fuzzyArgs.filename = fuzzyArgs.legalFileList[i]
				}
			}
		}
		if !foundFile {
			err = errors.New("No matching file found for " + args[1])
		}

	} else {
		err = errors.New("Filspec usage: <env>/<file> | <env> <file>")
	}
	return
}

// Parse args, expect one or more arguments that describes envs and/or apps
func (fuzzyArgs *FuzzyArgs) PopulateFuzzyEnvAppList(args []string, slashArg bool) (err error) {
	var i int
	var env string
	var app string
	var arg string
	var parts []string
	var argArray []string

	argArray = make([]string, 1)

	// First, parse envs
	for i = range args {
		if strings.Contains(string(args[i]), "/") {
			// We have a full path name with a slash, split it and try to match a specific env/app
			arg = args[i]
			parts = strings.Split(arg, "/")

			argArray[0] = parts[0]
			err = fuzzyArgs.PopulateFuzzyEnvAppList(parts, true)
			if err != nil {
				return err
			}
		} else {
			if slashArg {
				// Now we know that arg0 is and env and arg 1 is an app
				env, err = fuzzyArgs.GetFuzzyEnv(args[i])
				if err != nil {
					return err
				}
				if i == 0 {
					if env == "" {
						err = errors.New(args[i] + " does not match any environemt")
						return err
					}
					fuzzyArgs.AddEnv(env)
				}

				if i == 1 {
					app, err = fuzzyArgs.GetFuzzyApp(args[i])
					if err != nil {
						return err
					}
					if app == "" {
						err = errors.New(args[i] + " does not match any application")
						return err
					}
					fuzzyArgs.AddApp(app)
				}
			} else {
				// We have a single spec that is either an app or an env
				env, err = fuzzyArgs.GetFuzzyEnv(args[i])
				if err != nil {
					return err
				}
				app, err = fuzzyArgs.GetFuzzyApp(args[i])
				if err != nil {
					return err
				}

				if env != "" && app != "" {
					err = errors.New(args[i] + " matching both environment " + env + " and application " + app)
					return err
				}
				if env == "" && app == "" {
					err = errors.New(args[i] + " matching neither an environment nor an application")
					return err
				}
				if env != "" {
					fuzzyArgs.AddEnv(env)
				}
			}

		}

	}

	for i = range args {
		if !strings.Contains(args[i], "/") {
			if !slashArg {
				{
					// We have a single spec that is either an app or an env

					app, err = fuzzyArgs.GetFuzzyApp(args[i])
					if err != nil {
						return err
					}
					if app != "" {
						fuzzyArgs.AddApp(app)
					} else {
						env, err = fuzzyArgs.GetFuzzyEnv(args[i])
						if err != nil {
							return err
						}
						if env == "" {
							err = errors.New(args[i] + " matches neither an environment nor an application")
							return err
						}
					}
				}
			}
		}

	}
	return
}

func (fuzzyArgs *FuzzyArgs) AddApp(app string) {
	for i := range fuzzyArgs.appList {
		if fuzzyArgs.appList[i] == app {
			return
		}
	}
	fuzzyArgs.appList = append(fuzzyArgs.appList, app)
	return
}

func (fuzzyArgs *FuzzyArgs) AddEnv(env string) {
	for i := range fuzzyArgs.envList {
		if fuzzyArgs.envList[i] == env {
			return
		}
	}
	fuzzyArgs.envList = append(fuzzyArgs.envList, env)
	return
}

func (fuzzyArgs *FuzzyArgs) DeployAll() {
	fuzzyArgs.envList = fuzzyArgs.legalEnvList
	fuzzyArgs.appList = fuzzyArgs.legalAppList
}

func (fuzzyArgs *FuzzyArgs) GetApps() (apps []string) {
	return fuzzyArgs.appList
}

func (fuzzyArgs *FuzzyArgs) GetEnvs() (envs []string) {
	return fuzzyArgs.envList
}

func (fuzzyArgs *FuzzyArgs) GetApp() (app string, err error) {
	if len(fuzzyArgs.appList) > 1 {
		err = errors.New("No unique application identified")
		return "", err
	}
	if len(fuzzyArgs.appList) > 0 {
		return fuzzyArgs.appList[0], nil
	}
	return "", nil
}

func (fuzzyArgs *FuzzyArgs) GetEnv() (env string, err error) {
	if len(fuzzyArgs.envList) > 1 {
		err = errors.New("No unique environment identified")
		return "", err
	}
	if len(fuzzyArgs.envList) > 0 {
		return fuzzyArgs.envList[0], nil
	}
	return "", nil
}

func (fuzzyArgs *FuzzyArgs) IsLegalFile(filename string) (legal bool) {
	for i := range fuzzyArgs.legalFileList {
		if fuzzyArgs.legalFileList[i] == filename {
			return true
		}
	}
	return false
}

// Func to get a filename if we have just an appname
// Returns an error if several files exists.
func (fuzzyArgs *FuzzyArgs) App2File(app string) (filename string, err error) {
	if !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	}
	var found bool = false
	for i := range fuzzyArgs.legalFileList {
		if strings.Contains(fuzzyArgs.legalFileList[i], app) {
			if found {
				err = errors.New("Non-unique file identifier")
				return "", err
			}
			found = true
			filename = fuzzyArgs.legalFileList[i]
		}
	}
	if found {
		return filename, nil
	}
	return "", nil
}

// Func to get a filename if we expect the user to uniquely identify a file
func (fuzzyArgs *FuzzyArgs) GetFile() (filename string, err error) {

	if fuzzyArgs.filename != "" {
		return fuzzyArgs.filename, nil
	} else {
		err = errors.New("Not found")
		return "", err
	}
	return

}

func (fuzzyArgs *FuzzyArgs) addLegalFile(filename string) {

	fuzzyArgs.legalFileList = append(fuzzyArgs.legalFileList, filename)
	return
}

func (fuzzyArgs *FuzzyArgs) addLegalApp(app string) {
	for i := range fuzzyArgs.legalAppList {
		if fuzzyArgs.legalAppList[i] == app {
			return
		}
	}
	fuzzyArgs.legalAppList = append(fuzzyArgs.legalAppList, app)
	return
}

func (fuzzyArgs *FuzzyArgs) addLegalEnv(env string) {
	for i := range fuzzyArgs.legalEnvList {
		if fuzzyArgs.legalEnvList[i] == env {
			return
		}
	}
	fuzzyArgs.legalEnvList = append(fuzzyArgs.legalEnvList, env)
	return
}

func (fuzzyArgs *FuzzyArgs) GetDeploymentSummaryString() (output string) {
	output = "This will deploy " + strconv.Itoa(len(fuzzyArgs.GetApps())) + " applications in " + strconv.Itoa(len(fuzzyArgs.GetEnvs())) + " environments.\n"

	var headers []string
	headers = make([]string, 2)
	headers[0] = "ENVIRONMENT"
	headers[1] = "APPLICATION"

	output += printutil.FormatTable(headers, fuzzyArgs.envList, fuzzyArgs.appList)

	return output
}
