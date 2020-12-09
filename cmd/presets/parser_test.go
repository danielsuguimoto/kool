package presets

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/afero"
)

func TestExistsParser(t *testing.T) {
	presets := make(map[string]map[string]string)
	laravelPreset := make(map[string]string)

	laravelPreset["kool.yml"] = ""
	presets["laravel"] = laravelPreset

	p := NewParser()
	p.LoadPresets(presets)

	exists := p.Exists("laravel")

	if !exists {
		t.Error("did not find existing preset laravel")
	}

	exists = p.Exists("javascript")

	if exists {
		t.Error("found not existing preset javascript")
	}
}

func TestGetAllParser(t *testing.T) {
	var allPresets []string
	p := NewParser()

	allPresets = p.GetPresets("")

	if len(allPresets) != 0 {
		t.Error("presets should be empty")
	}

	presets := make(map[string]map[string]string)
	laravelPreset := make(map[string]string)
	symfonyPreset := make(map[string]string)

	laravelPreset["kool.yml"] = ""
	symfonyPreset["kool.yml"] = ""

	presets["laravel"] = laravelPreset
	presets["symfony"] = symfonyPreset

	p.LoadPresets(presets)

	allPresets = p.GetPresets("")

	if len(allPresets) != 2 || allPresets[0] != "laravel" || allPresets[1] != "symfony" {
		t.Error("failed to get all presets")
	}
}

func TestGetLanguagesParser(t *testing.T) {
	var allLanguages []string
	p := NewParser()

	allLanguages = p.GetLanguages()

	if len(allLanguages) != 0 {
		t.Error("languages should be empty")
	}

	presets := make(map[string]map[string]string)

	laravelPreset := make(map[string]string)
	symfonyPreset := make(map[string]string)

	laravelPreset["kool.yml"] = ""
	symfonyPreset["kool.yml"] = ""

	presets["laravel"] = laravelPreset
	presets["symfony"] = symfonyPreset

	p.LoadPresets(presets)

	configs := make(map[string]string)

	phpConfig := "language: php"

	configs["laravel"] = phpConfig
	configs["symfony"] = phpConfig

	p.LoadConfigs(configs)

	allLanguages = p.GetLanguages()

	if len(allLanguages) != 1 || allLanguages[0] != "php" {
		t.Error("failed to get languages")
	}
}

func TestGetPresetByLanguageParser(t *testing.T) {
	presets := make(map[string]map[string]string)

	phpPreset := make(map[string]string)
	jsPreset := make(map[string]string)

	phpPreset["kool.yml"] = ""
	jsPreset["kool.yml"] = ""

	presets["php_language"] = phpPreset
	presets["javascript_language"] = jsPreset

	configs := make(map[string]string)

	configs["php_language"] = "language: php"
	configs["javascript_language"] = "language: javascript"

	p := NewParser()
	p.LoadPresets(presets)
	p.LoadConfigs(configs)

	phpPresets := p.GetPresets("php")

	if len(phpPresets) != 1 || phpPresets[0] != "php_language" {
		t.Error("failed to get preset by language")
	}
}

func TestIgnorePresetMetaKeysParser(t *testing.T) {
	presets := make(map[string]map[string]string)

	testingPreset := make(map[string]string)

	testingPreset["kool.yml"] = ""

	presets["preset"] = testingPreset

	configs := make(map[string]string)

	configs["preset"] = "language: php"

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "kool.yml", []byte("scripts"), os.ModePerm)

	p := NewParserFS(fs)
	p.LoadPresets(presets)

	foundFiles := p.LookUpFiles("preset")

	if len(foundFiles) != 1 || foundFiles[0] != "kool.yml" {
		t.Errorf("expecting to find only 'kool.yml', found %v", foundFiles)
	}
}

func TestGetPresetKeyContentParser(t *testing.T) {
	presets := make(map[string]map[string]string)

	preset := make(map[string]string)

	preset["key1"] = "value1"
	preset["key2"] = "value2"
	preset["key3"] = "value3"

	presets["preset"] = preset

	p := NewParser()
	p.LoadPresets(presets)

	content := p.GetPresetKeyContent("preset", "key2")

	if content != "value2" {
		t.Errorf("expecting to find value 'value2', found %s", content)
	}

	content = p.GetPresetKeyContent("invalid_preset", "key1")

	if content != "" {
		t.Errorf("expecting to find value 'value2', found %s", content)
	}

	content = p.GetPresetKeyContent("preset", "invalid_key1")

	if content != "" {
		t.Errorf("expecting to find value 'value2', found %s", content)
	}
}

func TestSetPresetKeyContentParser(t *testing.T) {
	presets := make(map[string]map[string]string)

	preset := make(map[string]string)

	preset["key1"] = "value1"
	preset["key2"] = "value2"
	preset["key3"] = "value3"

	presets["preset"] = preset

	p := NewParser()
	p.LoadPresets(presets)

	p.SetPresetKeyContent("preset", "key2", "value2Changed")

	content := p.GetPresetKeyContent("preset", "key2")

	if content != "value2Changed" {
		t.Errorf("expecting to find value 'value2Changed', found %s", content)
	}

	p.SetPresetKeyContent("invalid_preset", "key1", "value1Changed")

	content = p.GetPresetKeyContent("preset", "key1")

	if content != "value1" {
		t.Errorf("expecting to find value 'value1', found %s", content)
	}

	p.SetPresetKeyContent("preset", "invalid_key", "value1Changed")

	content = p.GetPresetKeyContent("preset", "key1")

	if content != "value1" {
		t.Errorf("expecting to find value 'value1', found %s", content)
	}
}

func TestGetTemplatesParser(t *testing.T) {
	var allTemplates map[string]map[string]string
	p := NewParser()

	allTemplates = p.GetTemplates()

	if allTemplates != nil {
		t.Error("templates should be empty")
	}

	templates := make(map[string]map[string]string)

	template := make(map[string]string)

	template["key1"] = "value1"
	template["key2"] = "value2"
	template["key3"] = "value3"

	templates["template"] = template

	p.LoadTemplates(templates)

	allTemplates = p.GetTemplates()

	if !reflect.DeepEqual(allTemplates, templates) {
		t.Error("failed to get all presets")
	}
}

func TestWriteFilesParser(t *testing.T) {
	fs := afero.NewMemMapFs()

	p := NewParserFS(fs)

	presets := make(map[string]map[string]string)
	preset := make(map[string]string)

	preset["kool.yml"] = "value1"
	presets["preset"] = preset

	p.LoadPresets(presets)

	if _, err := p.WriteFiles("preset"); err != nil {
		t.Errorf("unexpected error writing file, err: %v", err)
	}

	if _, err := fs.Stat("kool.yml"); os.IsNotExist(err) {
		t.Error("could not write the file 'kool.yml'")
	}
}

func TestLoadPresetsParser(t *testing.T) {
	presets := map[string]map[string]string{
		"laravel": {"file": "content"},
	}

	p := &DefaultParser{}
	p.LoadPresets(presets)

	if ok := reflect.DeepEqual(p.Presets, presets); !ok {
		t.Error("did not load the presets correctly")
	}
}

func TestLoadTemplatesParser(t *testing.T) {
	templates := map[string]map[string]string{
		"service": {"serviceKey": "content"},
	}

	p := &DefaultParser{}
	p.LoadTemplates(templates)

	if ok := reflect.DeepEqual(p.Templates, templates); !ok {
		t.Error("did not load the templates correctly")
	}
}

func TestLoadConfigsParser(t *testing.T) {
	configs := map[string]string{
		"preset": "preset_config",
	}

	p := &DefaultParser{}
	p.LoadConfigs(configs)

	if ok := reflect.DeepEqual(p.Configs, configs); !ok {
		t.Error("did not load the configs correctly")
	}
}

func TestGetConfigParser(t *testing.T) {
	configs := map[string]string{
		"preset": `language: php
commands:
  create:
    - command
questions:
  question1:
    message: message?
    options:
      option1Key: option1Value
`,
	}

	p := &DefaultParser{}
	p.LoadConfigs(configs)

	cfg, err := p.GetConfig("preset")

	if err != nil {
		t.Errorf("unexpected error getting config, err: %v", err)
	}

	if cfg == nil {
		t.Error("unexpected empty preset configuration")
		return
	}

	if cfg.Language != "php" {
		t.Error("failed getting language preset configuration")
	}

	createCmds, createCmdsExists := cfg.Commands["create"]
	if !createCmdsExists || len(createCmds) != 1 || createCmds[0] != "command" {
		t.Error("failed getting create commands preset configuration")
	}

	questions1, question1Exists := cfg.Questions["question1"]
	if !question1Exists || questions1.Message != "message?" || len(questions1.Options) != 1 || questions1.Options[0].Key != "option1Key" || questions1.Options[0].Value != "option1Value" {
		t.Error("failed getting questions preset configuration")
	}
}
