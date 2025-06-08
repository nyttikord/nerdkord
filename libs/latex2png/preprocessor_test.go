package latex2png

import (
	"errors"
	"testing"
)

func TestPreprocess(t *testing.T) {
	t.Log("testing empty string")
	res, err := Preprocess("", &PreprocessingOptions{PreambleFile: "../../config/default.tex"})
	expected := "\\documentclass{standalone}\n\n\\begin{document}\n\\begin{minipage}{16cm}\n\n\\end{minipage}\n\\end{document}\n\n"
	if err != nil {
		t.Errorf("got error %s", err.Error())
	} else if res.Value.String() != expected {
		t.Errorf("got %s, want %s", res.Value.String(), expected)
	}

	t.Log("testing redefining documentclass")
	_, err = Preprocess("\\documentclass {article}", &PreprocessingOptions{PreambleFile: "../../config/default.tex"})
	if !errors.Is(err, ErrCantRedefineDocumentClass) {
		t.Errorf("should raise a CantRedefineDocumentclass error")
	}

	t.Log("testing forbidden command")
	_, err = Preprocess("\\include{aaa.pdf}", &PreprocessingOptions{
		ForbiddenCommands: []string{"include"},
		PreambleFile:      "../../config/default.tex"},
	)
	if !errors.Is(err, ErrForbiddenCommand) {
		t.Error("should raise a ForbiddenCommand error")
	}

	t.Log("testing inserting begin document")
	_, err = Preprocess("\\usepackage{amsmath}\n\\usepackage[margins = 1in]{geometry}\nCoucou", &PreprocessingOptions{
		CommandsBeforeBeginDocument: []string{"usepackage"},
		PreambleFile:                "../../config/default.tex",
	})
	if !errors.Is(err, ErrCmdWithoutBeginDocument) {
		t.Error("should raise a CmdWithoutBeginDocument error")
	}
}
