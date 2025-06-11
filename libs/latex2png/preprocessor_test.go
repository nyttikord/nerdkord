package latex2png

import (
	"errors"
	"testing"
)

func TestPreprocess(t *testing.T) {
	t.Log("testing empty string")
	res, err := Preprocess("", &PreprocessingOptions{TemplateFile: "../../config/template.tex"})
	expected := `
\documentclass{standalone}

\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{amsmath, amssymb}
\usepackage{lipsum}

\begin{document}
\begin{minipage}{16cm}

\end{minipage}
\end{document}
`
	if err != nil {
		t.Errorf("got error %s", err.Error())
	} else if res.Value.String() != expected {
		t.Errorf("got %s, want %s", res.Value.String(), expected)
	}

	t.Log("testing redefining documentclass")
	_, err = Preprocess("\\documentclass {article}", &PreprocessingOptions{TemplateFile: "../../config/template.tex"})
	if !errors.Is(err, ErrCantRedefineDocumentClass) {
		t.Errorf("should raise a CantRedefineDocumentclass error")
	}

	t.Log("testing forbidden command")
	_, err = Preprocess("\\include{aaa.pdf}", &PreprocessingOptions{
		ForbiddenCommands: []string{"include"},
		TemplateFile:      "../../config/template.tex"},
	)
	if !errors.Is(err, ErrForbiddenCommand) {
		t.Error("should raise a ForbiddenCommand error")
	}

	t.Log("testing inserting begin document")
	_, err = Preprocess("\\usepackage{amsmath}\n\\usepackage[margins = 1in]{geometry}\nCoucou", &PreprocessingOptions{
		CommandsBeforeBeginDocument: []string{"usepackage"},
		TemplateFile:                "../../config/template.tex",
	})
	if !errors.Is(err, ErrCmdWithoutBeginDocument) {
		t.Error("should raise a CmdWithoutBeginDocument error")
	}
}
