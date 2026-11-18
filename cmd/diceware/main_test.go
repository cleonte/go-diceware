package main

import (
	"bytes"
	"strings"
	"testing"
)

func execute(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestDefault(t *testing.T) {
	out, errOut, err := execute(t)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) == "" || strings.Contains(strings.TrimSpace(out), "\n") {
		t.Errorf("want one passphrase line, got %q", out)
	}
	if !strings.Contains(errOut, "Entropy: 77.5 bits (6 words, English wordlist)") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestSeparatorAndLanguage(t *testing.T) {
	for _, lang := range []string{"en", "ro", "mixed"} {
		out, _, err := execute(t, "-w", "4", "-s", "|", "-l", lang)
		if err != nil {
			t.Fatalf("-l %s: %v", lang, err)
		}
		if n := len(strings.Split(strings.TrimSpace(out), "|")); n != 4 {
			t.Errorf("-l %s: %d words in %q, want 4", lang, n, out)
		}
	}
}

func TestRolls(t *testing.T) {
	out, _, err := execute(t, "-r", "-w", "3", "-l", "mixed", "-s", " ")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "Dice rolls: [") || !strings.HasPrefix(lines[1], "Passphrase: ") {
		t.Fatalf("output = %q", out)
	}
	if len(strings.Fields(lines[0])) != 5 { // "Dice" "rolls:" and three rolls
		t.Errorf("want 3 rolls in %q", lines[0])
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"too few words", []string{"-w", "0"}, "word count must be between 1 and 24"},
		{"too many words", []string{"-w", "25"}, "word count must be between 1 and 24"},
		{"bad language", []string{"-l", "xx"}, "unsupported language"},
		{"positional argument", []string{"extra"}, "unknown command"},
		{"unknown flag", []string{"--nope"}, "unknown flag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := execute(t, tt.args...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want it to contain %q", err, tt.want)
			}
			if out != "" {
				t.Errorf("stdout = %q on error, want nothing", out)
			}
		})
	}
}

func TestLimits(t *testing.T) {
	for _, w := range []string{"1", "24"} {
		if _, _, err := execute(t, "-w", w); err != nil {
			t.Errorf("-w %s: %v", w, err)
		}
	}
}

func TestVersion(t *testing.T) {
	out, _, err := execute(t, "--version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "diceware version ") {
		t.Errorf("version output = %q", out)
	}
}

func TestCompletion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		out, _, err := execute(t, "completion", shell)
		if err != nil || !strings.Contains(out, "diceware") {
			t.Errorf("completion %s: err=%v, %d bytes", shell, err, len(out))
		}
	}
	if _, _, err := execute(t, "completion", "tcsh"); err == nil {
		t.Error("completion tcsh should fail")
	}
	if _, _, err := execute(t, "completion"); err == nil {
		t.Error("completion without a shell should fail")
	}
}
