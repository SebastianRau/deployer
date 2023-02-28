package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sebastianRau/deployer/pkg/templating"
)

type StepTypes string

const (
	CreateFolderType StepTypes = "mkdir"
	CopyType         StepTypes = "cp"
	DeleteType       StepTypes = "rm"
	ReplaceTextType  StepTypes = "replaceText"
	FileWriterType   StepTypes = "fileWriter"
	ExecType         StepTypes = "exec"
	EcecEachFileType StepTypes = "execEachFile"
	GitUpdateType    StepTypes = "gitUpdate"
	SubStepsType     StepTypes = "subSteps"
)

type DeployerStep struct {
	// mandatory
	Type      string      `json:"type"`
	Parameter interface{} `json:"parameter"`
	// optinal
	Description string `json:"description"`
	IgnoreError bool   `json:"ignoreError"`
}

type JsonConfig struct {
	Steps []DeployerStep `json:"steps"`
}

type ExceutableStep interface {
	Exec(v io.Writer) error
}

func UnmarshalConfigTemplate(templ string, data string) (*JsonConfig, error) {
	jsonResult := JsonConfig{}

	jsonConfigFile, err := templating.ParseTemplateJsonData(templ, data)
	if err != nil {
		return &jsonResult, err
	}

	if err := json.Unmarshal([]byte(jsonConfigFile), &jsonResult); err != nil {
		return &jsonResult, err
	}

	if jsonResult.Steps == nil {
		return &jsonResult, fmt.Errorf("No Steps defined")
	}

	for i, s := range jsonResult.Steps {
		if s.Type == "" {
			return &jsonResult, fmt.Errorf("No Type for Step #%d defined", i)
		}
		if s.Parameter == nil {
			return &jsonResult, fmt.Errorf("No Parameter for Step #%d (type: %s) defined", i, s.Type)
		}
	}

	return &jsonResult, err
}

func (c *JsonConfig) Marshal() (string, error) {
	result, err := json.Marshal(&c)
	return string(result), err
}

func (c *JsonConfig) MarshalIndent(ident string) (string, error) {
	result, err := json.MarshalIndent(&c, "", ident)
	return string(result), err
}

func (c *JsonConfig) Exceute(out io.Writer, verboseFlag bool) error {

	for _, s := range c.Steps {

		var ex ExceutableStep
		var err error

		switch StepTypes(s.Type) {

		case CreateFolderType:
			ex, err = UnmarschalCreatefolder(s)
		case CopyType:
			ex, err = UnmarschalCopy(s)
		case DeleteType:
			ex, err = UnmarschalDelete(s)
		case FileWriterType:
			ex, err = UnmarschalFileWriter(s)
		case ReplaceTextType:
			ex, err = UnmarschalReplaceText(s)
		case ExecType:
			ex, err = UnmarschalExec(s)
		case EcecEachFileType:
			ex, err = UnmarschalExecEach(s)
		case GitUpdateType:
			ex, err = UnmarschalGitUpdate(s)
		case SubStepsType:
			ex, err = UnmarschalSubSteps(s, out, verboseFlag)

		default:
			err = fmt.Errorf("Cant Parse type: %s\n", s.Type)
			ex = nil
		}

		// Exec Step
		if ex != nil {

			var write io.Writer
			v := bytes.NewBufferString("")

			if verboseFlag {
				write = v
			} else {
				write = nil
			}

			fmt.Fprint(out, formatStep(s.Description, s.Type, (err != nil), 80))
			err = ex.Exec(write)
			if err != nil && !s.IgnoreError {
				fmt.Fprintf(out, "%v\n", err.Error())
				return err
			}

			if verboseFlag {
				_, err = out.Write(v.Bytes())
				if err != nil {
					panic(err.Error())
				}
			}
		}

	}
	return nil
}

func formatStep(sName string, sType string, err bool, len int) string {
	result := "OK"
	if err {
		result = "ERROR"
	}

	return fmt.Sprintf("%-*s %s\n", len, sName, result)
}
