package git

import "os/exec"

func ExecuteCmd(name string, workdir string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	if workdir != "" {
		cmd.Dir = workdir
	}
	out, err := cmd.CombinedOutput()

	return string(out), err
}
