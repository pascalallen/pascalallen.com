package main

import "testing"

func TestExecuteCommandLogicLs(t *testing.T) {
	result := executeCommandLogic("ls")
	if result != "bio.txt" {
		t.Fatalf("expected 'bio.txt', got '%s'", result)
	}
}

func TestExecuteCommandLogicCatBioTxt(t *testing.T) {
	result := executeCommandLogic("cat bio.txt")
	if result != bioText {
		t.Fatalf("expected bio text, got '%s'", result)
	}
}

func TestExecuteCommandLogicCatMissingArg(t *testing.T) {
	result := executeCommandLogic("cat")
	if result != "cat: missing file operand" {
		t.Fatalf("expected missing file operand error, got '%s'", result)
	}
}

func TestExecuteCommandLogicCatUnknownFile(t *testing.T) {
	result := executeCommandLogic("cat unknown.txt")
	if result != "cat: unknown.txt: No such file or directory" {
		t.Fatalf("expected not found error, got '%s'", result)
	}
}

func TestExecuteCommandLogicWhoami(t *testing.T) {
	result := executeCommandLogic("whoami")
	if result != "pascal" {
		t.Fatalf("expected 'pascal', got '%s'", result)
	}
}

func TestExecuteCommandLogicHelp(t *testing.T) {
	result := executeCommandLogic("help")
	if result != "available commands: cat, clear, help, ls, whoami" {
		t.Fatalf("expected help text, got '%s'", result)
	}
}

func TestExecuteCommandLogicClear(t *testing.T) {
	result := executeCommandLogic("clear")
	if result != "__CLEAR__" {
		t.Fatalf("expected '__CLEAR__', got '%s'", result)
	}
}

func TestExecuteCommandLogicUnknownCommand(t *testing.T) {
	result := executeCommandLogic("foo")
	if result != "bash: foo: command not found" {
		t.Fatalf("expected command not found, got '%s'", result)
	}
}

func TestExecuteCommandLogicEmpty(t *testing.T) {
	result := executeCommandLogic("")
	if result != "" {
		t.Fatalf("expected empty string, got '%s'", result)
	}
}
