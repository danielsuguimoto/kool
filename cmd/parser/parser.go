package parser

import (
	"errors"
	"kool-dev/kool/cmd/builder"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Parser defines the functions required for handling kool.yml files.
type Parser interface {
	AddLookupPath(string) error
	Parse(string) ([]builder.Command, error)
	ParseAvailableScripts(string) ([]string, error)
	LookUpVariables(string) ([]string)
}

// DefaultParser implements all default behavior for using kool.yml files.
type DefaultParser struct {
	targetFiles []string
	lookedUp    map[string]bool
}

// NewParser initializes a Parser to be used for handling kool.yml scripts.
func NewParser() Parser {
	return &DefaultParser{}
}

// AddLookupPath adds a folder to look for kool.yml scripts file.
func (p *DefaultParser) AddLookupPath(rootPath string) (err error) {
	var koolFile string

	if p.lookedUp == nil {
		p.lookedUp = make(map[string]bool)
	}

	ymlPath := path.Join(rootPath, "kool.yml")
	yamlPath := path.Join(rootPath, "kool.yaml")

	if _, err = os.Stat(ymlPath); err == nil {
		koolFile = ymlPath
	} else if _, err = os.Stat(yamlPath); err == nil {
		koolFile = yamlPath
	}

	if koolFile == "" {
		err = ErrKoolYmlNotFound
	} else {
		if !p.lookedUp[koolFile] {
			p.targetFiles = append(p.targetFiles, koolFile)
		}

		p.lookedUp[koolFile] = true
	}

	return
}

// Parse looks up for the given script name on all of the kool.yml files available
// on the configured lookup paths. If the script exists in more than one file
// this function will return the first occurrence and an ErrMultipleDefinedScript
// error just to let the user know and avoid confusing.
func (p *DefaultParser) Parse(script string) (commands []builder.Command, err error) {
	var (
		koolFile        string
		parsedFile      *KoolYaml
		found           bool
		previouslyFound bool
	)

	if len(p.targetFiles) == 0 {
		err = errors.New("kool.yml not found")
		return
	}

	for _, koolFile = range p.targetFiles {
		if parsedFile, err = ParseKoolYaml(koolFile); err != nil {
			return
		}

		if found = parsedFile.HasScript(script); found {
			if !previouslyFound {
				// this is the first time we find the script we want!
				previouslyFound = true

				if commands, err = parsedFile.ParseCommands(script); err != nil {
					return
				}
			} else {
				// so we already found once, and now found again the same script
				// in another file! let's warn about that
				err = ErrMultipleDefinedScript
			}
		}
	}
	return
}

// ParseAvailableScripts parse all available scripts
func (p *DefaultParser) ParseAvailableScripts(filter string) (scripts []string, err error) {
	var (
		koolFile     string
		parsedFile   *KoolYaml
		foundScripts map[string]bool
	)

	if len(p.targetFiles) == 0 {
		err = errors.New("kool.yml not found")
		return
	}

	foundScripts = make(map[string]bool)

	for _, koolFile = range p.targetFiles {
		if parsedFile, err = ParseKoolYaml(koolFile); err != nil {
			return
		}

		for script := range parsedFile.Scripts {
			if !foundScripts[script] && (filter == "" || strings.HasPrefix(script, filter)) {
				scripts = append(scripts, script)
			}

			foundScripts[script] = true
		}
	}

	sort.Strings(scripts)

	return
}

// LookUpVariables look for variables in the script
func (p *DefaultParser) LookUpVariables(script string) (variables []string) {
	var (
		koolFile        string
		parsedFile      *KoolYaml
		found           bool
		err             error
	)

	if len(p.targetFiles) == 0 {
		return
	}

	r := regexp.MustCompile("\\${.*?}")

	for _, koolFile = range p.targetFiles {
		if parsedFile, err = ParseKoolYaml(koolFile); err != nil {
			return
		}

		if found = parsedFile.HasScript(script); found {
			if line, isSingle := parsedFile.Scripts[script].(string); isSingle {
				variables = append(
					variables,
					r.FindAllString(line, -1)...,
				)
			} else if lines, isList := parsedFile.Scripts[script].([]interface{}); isList {
				for _, line := range lines {
					variables = append(
						variables,
						r.FindAllString(line.(string), -1)...,
					)
				}
			}
		}
	}

	for index := range variables {
		variable := variables[index]
		variables[index] = variable[2:len(variable) - 1]
	}

	return
}
