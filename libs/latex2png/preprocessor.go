package latex2png

import (
	"bytes"
	"errors"
	"html/template"
	"regexp"
	"strings"
)

type PreprocessingResult struct {
	Value *bytes.Buffer
	Debug error
}

type PreprocessingOptions struct {
	// do not put a \ in front of the commands
	// `documentclass` is forbidden by default
	ForbiddenCommands           []string
	CommandsBeforeBeginDocument []string
	TemplateFile                string
	UserPreamble                string
}

var (
	ErrPreprocessor              = errors.New("preprocessing error")
	ErrCantRedefineDocumentClass = errors.New("cannot redefine documentclass")
	ErrForbiddenCommand          = errors.New("forbidden command")
	ErrCmdWithoutBeginDocument   = errors.New("command without `\\begin{document}`")
)

type templateData struct {
	Preamble string
	Document string
	After    string
}

func Preprocess(input string, opt *PreprocessingOptions) (*PreprocessingResult, error) {
	var err error = nil
	var debug error = nil

	res := new(bytes.Buffer)

	if strings.Contains(input, `\documentclass`) {
		err = errors.Join(err, ErrCantRedefineDocumentClass)
	}

	for _, cmd := range opt.ForbiddenCommands {
		if strings.Contains(input, `\`+cmd) {
			err = errors.Join(err, ErrForbiddenCommand, errors.New("    command `\\"+cmd+"` is forbidden"))
		}
	}

	data := templateData{}

	beginReg := regexp.MustCompile(`\\begin\s*{document}`)
	beginPos := beginReg.FindStringIndex(input)
	if beginPos == nil {
		for _, cmd := range opt.CommandsBeforeBeginDocument {
			if strings.Contains(input, `\`+cmd) {
				err = errors.Join(err, ErrCmdWithoutBeginDocument, errors.New("    can't use `\\"+cmd+"` without `\\begin{document}`"))
			}
		}
		data.Preamble = opt.UserPreamble
		data.Document = input
	} else {
		endReg := regexp.MustCompile(`\\end\s*{document}`)

		endPos := endReg.FindStringIndex(input)

		data.Preamble = input[:beginPos[1]]
		data.Document = input[beginPos[1]:endPos[0]]
		data.After = input[endPos[0]:]
	}

	t, e := template.ParseFiles(opt.TemplateFile)
	if e != nil {
		return nil, errors.Join(err, e)
	}
	e = t.Execute(res, data)
	if e != nil {
		return nil, errors.Join(err, e)
	}

	if err != nil {
		err = errors.Join(ErrPreprocessor, err)
	}

	return &PreprocessingResult{Value: res, Debug: debug}, err
}
