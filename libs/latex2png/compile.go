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
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
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
		_ = f.Close()
	}(outputFile)
	if err != nil {
		return err
	}

	_, err = outputFile.WriteTo(output)
	if err != nil {
		return nil
	}

	return os.RemoveAll(tempDir)
}
