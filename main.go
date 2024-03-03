package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

/* go run main.go run <cmd> <args> */
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("unexpected argument")
	}
}

func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	execute(cmd.Run())
}

func child() {
	fmt.Printf("Running %v \n", os.Args[2:])

	restrict_by_cgroup()
	
	/* exec new process */
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	execute(syscall.Sethostname([]byte("container")))

	/* change root using chroot system call */
	execute(syscall.Chroot("/home/dev/rootfs"))

	execute(os.Chdir("/"))
	execute(syscall.Mount("proc", "proc", "proc", 0, ""))

	execute(cmd.Run())

	execute(syscall.Unmount("proc", 0))
	execute(syscall.Unmount("thing", 0))
}

func restrict_by_cgroup() {
	cgroups := "/sys/fs/cgroup/"

	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "dev"), 0755)
	execute(ioutil.WriteFile(filepath.Join(pids, "dev/pids.max"), []byte("20"), 0700))

	/* Remove cgroup after container exit */
	execute(ioutil.WriteFile(filepath.Join(pids, "dev/notify_on_release"), []byte("1"), 0700))
	execute(ioutil.WriteFile(filepath.Join(pids, "dev/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func execute(err error) {
	if err != nil {
		panic(err)
	}
}