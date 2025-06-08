package latex2png

import (
	"bytes"
	"errors"
	"os"
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
	PreambleFile                string
}

var (
	ErrPreprocessor              = errors.New("Preprocessing error:")
	ErrCantRedefineDocumentClass = errors.New("cannot redefine documentclass")
	ErrForbiddenCommand          = errors.New("forbidden command")
	ErrCmdWithoutBeginDocument   = errors.New("command without `\\begin{document}`")
)

func Preprocess(input string, opt *PreprocessingOptions) (PreprocessingResult, error) {
	var err error = nil
	var debug error = nil

	res := new(bytes.Buffer)

	if strings.Contains(input, "\\documentclass") {
		err = errors.Join(err, ErrCantRedefineDocumentClass)
	}

	for _, cmd := range opt.ForbiddenCommands {
		if strings.Contains(input, "\\"+cmd) {
			err = errors.Join(err, ErrForbiddenCommand, errors.New("    command `\\"+cmd+"` is forbidden"))
		}
	}

	beginReg := regexp.MustCompile("\\\\begin\\s*{document}")
	if !beginReg.MatchString(input) {
		for _, cmd := range opt.CommandsBeforeBeginDocument {
			if strings.Contains(input, "\\"+cmd) {
				err = errors.Join(err, ErrCmdWithoutBeginDocument, errors.New("    can't use `\\"+cmd+"` without `\\begin{document}`"))
			}
		}

		debug = errors.Join(debug, errors.New("inserting `\\begin{document}\\begin{minipage}{16cm}` at input start"))
		debug = errors.Join(debug, errors.New("inserting `\\end{minipage}\\end{document}` at the end of input"))
		input = "\\begin{document}\n\\begin{minipage}{16cm}\n" + input + "\n\\end{minipage}\n\\end{document}"
	} else {
		endReg := regexp.MustCompile("\\\\end\\s*{document}")

		beginPos := beginReg.FindStringIndex(input)
		endPos := endReg.FindStringIndex(input)
		input = input[:beginPos[1]] +
			"\n\\begin{minipage}{16cm}\n" +
			input[beginPos[1]:endPos[0]] +
			"\n\\end{minipage}\n" +
			input[endPos[0]:]

		debug = errors.Join(debug, errors.New("inserting `\\begin{minipage}{16cm}` after begin document"))
		debug = errors.Join(debug, errors.New("inserting `\\end{minipage}` before end document"))
	}

	preambleFile, e := os.Open(opt.PreambleFile)
	defer func(f *os.File) {
		_ = f.Close()
	}(preambleFile)

	if e != nil {
		err = errors.Join(err, e)
	} else {
		debug = errors.Join(debug, errors.New("writing preamble content to buffer"))
		_, e := preambleFile.WriteTo(res)
		if e != nil {
			err = errors.Join(err, e)
		}
	}

	debug = errors.Join(debug, errors.New("writing input to buffer"))
	res.WriteString(input)

	if err != nil {
		err = errors.Join(ErrPreprocessor, err)
	}

	return PreprocessingResult{Value: res, Debug: debug}, err
}
