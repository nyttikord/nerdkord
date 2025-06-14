package latex2png

import (
	"bytes"
	"errors"
	"github.com/anhgelus/gokord/utils"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ErrLatexCompilation struct {
	rawErr *bytes.Buffer
}

func (e ErrLatexCompilation) Error() string {
	s := e.rawErr.String()

	l := strings.Split(s, "!")
	s = strings.Join(l[1:], "!")

	l = strings.Split(s, "Transcript written on ")
	return "!" + strings.Join(l[:len(l)-1], "Transcript written on ")
}

func Compile(output io.Writer, latex string, opt *Options) error {
	res, err := Preprocess(latex, opt.PreprocessingOptions)
	if err != nil {
		return err
	}
	if res.Debug != nil {
		utils.SendDebug("Latex preprocessing debug:\n" + res.Debug.Error())
	}

	var tempDir string
	if opt.TempDir != "" {
		tempDir = opt.TempDir
	} else {
		tempDir = os.TempDir()
	}
	tempDir, err = os.MkdirTemp(tempDir, "nerdkord_*_latex_compilation")
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(tempDir, "nerdkord_*.tex")
	if err != nil {
		return err
	}

	_, err = res.Value.WriteTo(f)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		opt.LatexBinary,
		"-output-format=dvi",
		"-output-directory="+tempDir,
		f.Name(),
	)
	var outErr bytes.Buffer
	cmd.Stdout = &outErr

	err = cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return ErrLatexCompilation{rawErr: &outErr}
		}
		return err
	}

	bg := formatColor(opt.BackgroundColor)
	fg := formatColor(opt.ForegroundColor)
	cmd = exec.Command(
		opt.DvipngBinary,
		"-D",
		strconv.FormatInt(int64(opt.ImageDPI), 10),
		"-bg",
		bg,
		"-fg",
		fg,
		"-T",
		"bbox",
		opt.OutputFormat,
		"-o",
		strings.Split(f.Name(), ".")[0]+".png",
		strings.Split(f.Name(), ".")[0]+".dvi",
	)

	// Ignore this error because it is triggered everytime
	_ = cmd.Run()

	outputFile, err := os.Open(strings.Split(f.Name(), ".")[0] + ".png")
	if err != nil {
		return err
	}

	_, err = outputFile.WriteTo(output)
	if err != nil {
		return err
	}

	_ = outputFile.Close()
	_ = f.Close()

	err = os.RemoveAll(tempDir)
	if err != nil {
		utils.SendAlert("commands/latex.go - Removing temporary files", err.Error())
	}

	return nil
}
