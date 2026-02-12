package gobake

import (
	"os"
	"os/exec"
	"testing"
)

func TestReservedTaskName(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		e := NewEngine()
		e.Task("init", "fail", func(ctx *Context) error { return nil })
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestReservedTaskName")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
