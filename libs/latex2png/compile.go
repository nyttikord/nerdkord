package latex2png

import (
	"github.com/anhgelus/gokord/utils"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Compile(output io.Writer, latex string, opt *Options) error {
	preambleFile, err := os.Open(opt.PreambleFilePath)
	defer func(f *os.File) {
		name := f.Name()
		_ = f.Close()
		err = os.Remove(name)
		if err != nil {
			utils.SendAlert("latex2png/compile.go - Removing preamble file", err.Error())
		}
	}(preambleFile)

	if err != nil {
		return err
	}

	if opt.AddBeginDocument {
		// Minipage environment forces a maximal width
		latex = "\\begin{document}\n\\begin{minipage}{16cm}\n" + latex + "\\end{minipage}\n\\end{document}"
	}

	var tempDir string
	if opt.TempDir != "" {
		tempDir = opt.TempDir
	} else {
		tempDir = os.TempDir()
	}

	f, err := os.CreateTemp(tempDir, "nerdkord_*.tex")
	defer func(f *os.File) {
		name := f.Name()
		_ = f.Close()
		err = os.Remove(name)
		if err != nil {
			utils.SendAlert("latex2png/compile.go - Removing temporary file", err.Error())
		}
	}(f)
	if err != nil {
		return err
	}

	_, err = preambleFile.WriteTo(f)
	if err != nil {
		return err
	}

	_, err = f.WriteString(latex)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		opt.LatexBinary,
		"-output-format=dvi",
		"-output-directory="+tempDir,
		f.Name(),
	)
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
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

	err = cmd.Start()
	if err != nil {
		return err
	}
	// Ignore this error because it is triggered everytime
	_ = cmd.Wait()

	outputFile, err := os.Open(strings.Split(f.Name(), ".")[0] + ".png")
	defer func(f *os.File) {
		name := f.Name()
		_ = f.Close()
		err = os.Remove(name)
		if err != nil {
			utils.SendAlert("latex2png/compile.go - Removing temporary file", err.Error())
		}
	}(outputFile)
	if err != nil {
		return err
	}

	_, err = outputFile.WriteTo(output)
	return nil
}
