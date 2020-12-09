package cmd

import (
	"errors"
	"fmt"
	"kool-dev/kool/cmd/compose"
	"kool-dev/kool/cmd/presets"
	"kool-dev/kool/cmd/shell"
	"strings"

	"github.com/spf13/cobra"
)

// KoolPresetFlags holds the flags for the preset command
type KoolPresetFlags struct {
	Override bool
}

// KoolPreset holds handlers and functions to implement the preset command logic
type KoolPreset struct {
	DefaultKoolService
	Flags         *KoolPresetFlags
	presetsParser presets.Parser
	composeParser compose.Parser
	promptSelect  shell.PromptSelect
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
		&presets.DefaultParser{},
		compose.NewParser(),
		shell.NewPromptSelect(),
	}
}

// Execute runs the preset logic with incoming arguments.
func (p *KoolPreset) Execute(args []string) (err error) {
	var (
		fileError, preset, language string
		useDefaultCompose           bool
		servicesOptions             map[string]string
	)

	if len(args) == 0 {
		if !p.IsTerminal() {
			err = fmt.Errorf("the input device is not a TTY; for non-tty environments, please specify a preset argument")
			return
		}

		if language, err = p.promptSelect.Ask("What language do you want to use", p.presetsParser.GetLanguages()); err != nil {
			return
		}

		if preset, err = p.promptSelect.Ask("What preset do you want to use", p.presetsParser.GetPresets(language)); err != nil {
			return
		}
	} else {
		preset = args[0]
	}

	p.presetsParser.LoadPresets(presets.GetAll())

	if !p.presetsParser.Exists(preset) {
		err = fmt.Errorf("Unknown preset %s", preset)
		return
	}

	servicesOptions = make(map[string]string)
	useDefaultCompose = true

	if servicesToAskStr := p.presetsParser.GetPresetKeyContent(preset, "preset_ask_services"); servicesToAskStr != "" && p.IsTerminal() {
		servicesToAsk := strings.Split(servicesToAskStr, ",")

		for _, serviceName := range servicesToAsk {
			optionsKey := fmt.Sprintf("preset_%s_options", serviceName)
			question := fmt.Sprintf("What %s service do you want to use", serviceName)

			if optionsStr := p.presetsParser.GetPresetKeyContent(preset, optionsKey); optionsStr != "" {
				options := strings.Split(optionsStr, ",")

				if servicesOptions[serviceName], err = p.promptSelect.Ask(question, options); err != nil {
					return
				}
				useDefaultCompose = false
			}
		}
	}

	p.Println("Preset", preset, "is initializing!")

	if !p.Flags.Override {
		existingFiles := p.presetsParser.LookUpFiles(preset)
		for _, fileName := range existingFiles {
			p.Warning("Preset file ", fileName, " already exists.")
		}

		if len(existingFiles) > 0 {
			err = ErrPresetFilesAlreadyExists
			return
		}
	}

	presetKeys := p.presetsParser.GetPresetKeys(preset)

	templates := p.presetsParser.GetTemplates()

	for _, presetKey := range presetKeys {
		if strings.HasPrefix(presetKey, "preset_") {
			continue
		}

		var content string

		if presetKey == "docker-compose.yml" && !useDefaultCompose {
			defaultCompose := p.presetsParser.GetPresetKeyContent(preset, presetKey)

			if err = p.composeParser.Load(defaultCompose); err != nil {
				err = fmt.Errorf("Failed to write preset file %s: %v", presetKey, err)
				return
			}

			for serviceKey, serviceOption := range servicesOptions {
				if serviceOption == "none" {
					p.composeParser.RemoveService(serviceKey)
					p.composeParser.RemoveVolume(serviceKey)
				} else {
					key := formatTemplateKey(serviceOption)
					service := templates[serviceKey][key]

					if err = p.composeParser.SetService(serviceKey, service); err != nil {
						err = fmt.Errorf("Failed to write preset file %s: %v", presetKey, err)
						return
					}
				}
			}

			if content, err = p.composeParser.String(); err != nil {
				err = fmt.Errorf("Failed to write preset file %s: %v", presetKey, err)
				return
			}
		} else {
			content = p.presetsParser.GetPresetKeyContent(preset, presetKey)
		}

		if fileError, err = p.presetsParser.WriteFile(presetKey, content); err != nil {
			err = fmt.Errorf("Failed to write preset file %s: %v", fileError, err)
			return
		}
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

func formatTemplateKey(key string) (formattedKey string) {
	formattedKey = strings.ReplaceAll(key, " ", "")
	formattedKey = strings.ReplaceAll(formattedKey, ".", "")
	formattedKey = strings.ToLower(formattedKey) + ".yml"
	return
}
