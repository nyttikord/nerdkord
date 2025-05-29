package latex2png

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Compile(latex string, opt *Options) (*os.File, error) {
	preambleFile, err := os.Open(opt.PreambleFilePath)
	if err != nil {
		panic(err)
		return nil, err
	}

	if opt.AddBeginDocument {
		latex = "\\begin{document}" + latex + "\\end{document}"
	}

	var tempDir string
	if opt.TempDir != "" {
		tempDir = opt.TempDir
	} else {
		tempDir = os.TempDir()
	}

	f, err := os.CreateTemp(tempDir, "*.tex")
	if err != nil {
		panic(err)
		return nil, err
	}

	_, err = preambleFile.WriteTo(f)
	if err != nil {
		panic(err)
		return nil, err
	}

	err = preambleFile.Close()
	if err != nil {
		panic(err)
		return nil, err
	}

	_, err = f.WriteString(latex)
	if err != nil {
		panic(err)
		return nil, err
	}

	cmd := exec.Command(
		opt.LatexBinary,
		"-output-format=dvi",
		"-output-directory="+tempDir,
		f.Name(),
	)
	err = cmd.Start()
	if err != nil {
		panic(err)
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		panic(err)
		return nil, err
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
		panic(err)
		return nil, err
	}
	// Ignore this error because it is triggered everytime
	_ = cmd.Wait()

	err = f.Close()
	if err != nil {
		panic(err)
		return nil, err
	}

	return os.Open(strings.Split(f.Name(), ".")[0] + ".png")
}
