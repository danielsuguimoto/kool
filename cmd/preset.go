package cmd

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"kool-dev/kool/cmd/compose"
	"kool-dev/kool/cmd/parser"
	"kool-dev/kool/cmd/presets"
	"kool-dev/kool/cmd/shell"
	"kool-dev/kool/cmd/templates"

	"github.com/spf13/cobra"
)

// KoolPresetFlags holds the flags for the preset command
type KoolPresetFlags struct {
	Override bool
}

// KoolPreset holds handlers and functions to implement the preset command logic
type KoolPreset struct {
	DefaultKoolService
	Flags          *KoolPresetFlags
	presetsParser  presets.Parser
	composeParser  compose.Parser
	templateParser templates.Parser
	koolYamlParser parser.KoolYamlParser
	promptSelect   shell.PromptSelect
}

// ErrPresetFilesAlreadyExists error for existing presets files
var ErrPresetFilesAlreadyExists = errors.New("some preset files already exist")

func init() {
	var (
		preset    = NewKoolPreset()
		presetCmd = NewPresetCommand(preset)
	)

	rootCmd.AddCommand(presetCmd)
}

// NewKoolPreset creates a new handler for preset logic
func NewKoolPreset() *KoolPreset {
	return &KoolPreset{
		*newDefaultKoolService(),
		&KoolPresetFlags{false},
		presets.NewParser(),
		compose.NewParser(),
		templates.NewParser(),
		&parser.KoolYaml{},
		shell.NewPromptSelect(),
	}
}

// Execute runs the preset logic with incoming arguments.
func (p *KoolPreset) Execute(args []string) (err error) {
	var fileError, preset string

	p.loadParsers()

	if preset, err = p.getPresetArgOrAsk(args); err != nil {
		return
	}

	if !p.presetsParser.Exists(preset) {
		err = fmt.Errorf("Unknown preset %s", preset)
		return
	}

	p.Println("Preset", preset, "is initializing!")

	if !p.Flags.Override {
		if existingFiles := p.presetsParser.LookUpFiles(preset); len(existingFiles) > 0 {
			for _, fileName := range existingFiles {
				p.Warning("Preset file ", fileName, " already exists.")
			}

			err = ErrPresetFilesAlreadyExists
			return
		}
	}

	if err = p.customizePreset(preset); err != nil {
		return
	}

	if fileError, err = p.presetsParser.WriteFiles(preset); err != nil {
		err = fmt.Errorf("Failed to write preset file %s: %v", fileError, err)
		return
	}

	p.Success("Preset ", preset, " initialized!")
	return
}

// NewPresetCommand initializes new kool preset command
func NewPresetCommand(preset *KoolPreset) (presetCmd *cobra.Command) {
	presetCmd = &cobra.Command{
		Use:   "preset [PRESET]",
		Short: "Initialize kool preset in the current working directory. If no preset argument is specified you will be prompted to pick among the existing options.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			preset.SetOutStream(cmd.OutOrStdout())
			preset.SetInStream(cmd.InOrStdin())
			preset.SetErrStream(cmd.ErrOrStderr())

			if err := preset.Execute(args); err != nil {
				if err.Error() == ErrPresetFilesAlreadyExists.Error() {
					preset.Warning("Some preset files already exist. In case you wanna override them, use --override.")
					preset.Exit(2)
				} else if err.Error() == shell.ErrPromptSelectInterrupted.Error() {
					preset.Warning("Operation Cancelled")
					preset.Exit(0)
				} else {
					preset.Error(err)
					preset.Exit(1)
				}
			}
		},
	}

	presetCmd.Flags().BoolVarP(&preset.Flags.Override, "override", "", false, "Force replace local existing files with the preset files")
	return
}

func (p *KoolPreset) loadParsers() {
	p.presetsParser.LoadPresets(presets.GetAll())
	p.presetsParser.LoadTemplates(presets.GetTemplates())
	p.presetsParser.LoadConfigs(presets.GetConfigs())
}

func (p *KoolPreset) getPresetArgOrAsk(args []string) (preset string, err error) {
	if len(args) == 0 {
		if !p.IsTerminal() {
			err = fmt.Errorf("the input device is not a TTY; for non-tty environments, please specify a preset argument")
			return
		}

		var language string
		if language, err = p.promptSelect.Ask("What language do you want to use", p.presetsParser.GetLanguages()); err != nil {
			return
		}

		preset, err = p.promptSelect.Ask("What preset do you want to use", p.presetsParser.GetPresets(language))
	} else {
		preset = args[0]
	}

	return
}

func (p *KoolPreset) customizePreset(preset string) (err error) {
	var presetConfig *presets.PresetConfig

	if presetConfig, err = p.presetsParser.GetConfig(preset); err != nil || presetConfig == nil {
		err = fmt.Errorf("error parsing preset config; err: %v", err)
		return
	}

	if err = p.setDefaultTemplates(presetConfig); err != nil {
		return
	}

	if err = p.customizeCompose(preset, presetConfig); err != nil {
		return
	}

	err = p.customizeKoolYaml(preset, presetConfig)
	return
}

func (p *KoolPreset) setDefaultTemplates(config *presets.PresetConfig) (err error) {
	allTemplates := p.presetsParser.GetTemplates()

	for _, template := range config.Templates {
		if err = p.templateParser.Parse(allTemplates[template.Key][template.Template]); err != nil {
			err = fmt.Errorf("Failed to write preset file docker-compose.yml: %v", err)
			return
		}

		for _, service := range p.templateParser.GetServices() {
			p.composeParser.SetService(template.Key, service.Value.(yaml.MapSlice))
		}

		for _, volume := range p.templateParser.GetVolumes() {
			p.composeParser.SetVolume(volume.Key.(string))
		}

		for scriptName, scripts := range p.templateParser.GetScripts() {
			p.koolYamlParser.SetScript(scriptName, scripts)
		}
	}

	return
}

func (p *KoolPreset) customizeCompose(preset string, config *presets.PresetConfig) (err error) {
	var newCompose string
	allTemplates := p.presetsParser.GetTemplates()

	if servicesToAsk := config.Questions["compose"]; len(servicesToAsk) > 0 {
		for _, question := range servicesToAsk {
			var (
				options        []string
				selectedOption string = question.DefaultAnswer
				serviceName           = question.Key
			)

			optionTemplate := make(map[string]string)

			for _, option := range question.Options {
				options = append(options, option.Name)
				optionTemplate[option.Name] = allTemplates[serviceName][option.Template]
			}

			if p.IsTerminal() {
				if selectedOption, err = p.promptSelect.Ask(question.Message, options); err != nil {
					return
				}
			}

			if selectedOption != "none" {
				if err = p.templateParser.Parse(optionTemplate[selectedOption]); err != nil {
					err = fmt.Errorf("Failed to write preset file docker-compose.yml: %v", err)
					return
				}

				for _, service := range p.templateParser.GetServices() {
					p.composeParser.SetService(serviceName, service.Value.(yaml.MapSlice))
				}

				for _, volume := range p.templateParser.GetVolumes() {
					p.composeParser.SetVolume(volume.Key.(string))
				}

				for scriptName, scripts := range p.templateParser.GetScripts() {
					p.koolYamlParser.SetScript(scriptName, scripts)
				}
			}
		}

		if newCompose, err = p.composeParser.String(); err != nil {
			err = fmt.Errorf("Failed to write preset file docker-compose.yml: %v", err)
			return
		}

		p.presetsParser.SetPresetKeyContent(preset, "docker-compose.yml", newCompose)
	}

	return
}

func (p *KoolPreset) customizeKoolYaml(preset string, config *presets.PresetConfig) (err error) {
	var newKoolYaml string
	allTemplates := p.presetsParser.GetTemplates()

	if scriptsToAsk := config.Questions["kool"]; len(scriptsToAsk) > 0 {
		for _, question := range scriptsToAsk {
			var (
				options        []string
				selectedOption string = question.DefaultAnswer
			)

			optionTemplate := make(map[string]string)

			for _, option := range question.Options {
				options = append(options, option.Name)
				optionTemplate[option.Name] = allTemplates["scripts"][option.Template]
			}

			if p.IsTerminal() {
				if selectedOption, err = p.promptSelect.Ask(question.Message, options); err != nil {
					return
				}
			}

			if selectedOption != "none" {
				if err = p.templateParser.Parse(optionTemplate[selectedOption]); err != nil {
					err = fmt.Errorf("Failed to write preset file kool.yml: %v", err)
					return
				}

				for scriptName, scripts := range p.templateParser.GetScripts() {
					p.koolYamlParser.SetScript(scriptName, scripts)
				}
			}
		}

		if newKoolYaml, err = p.koolYamlParser.String(); err != nil {
			err = fmt.Errorf("Failed to write preset file kool.yml: %v", err)
			return
		}

		p.presetsParser.SetPresetKeyContent(preset, "kool.yml", newKoolYaml)
	}

	return
}
