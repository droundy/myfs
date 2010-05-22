package main

import (
	"os"
	"os/signal"
	"fmt"
	"bitbucket.org/taruti/go-extra/fuse"
)

func main() {
	mountpoint := "/tmp/m"
	fuseFd,err := fuse.MountFuse(mountpoint, []string{})
  fmt.Printf("f = %v, err = %v\n", fuseFd, err) 
	if err != nil {
		fmt.Println("I see an error!\n", err)
		return
	}
	defer UnmountFuse(mountpoint)
	go func() {
		err = fuse.FuseLoop(&MyFS{}, fuseFd)
		fmt.Printf("err = %v\n", err) 
	}()
	for {
		switch sig := <- signal.Incoming; sig.(signal.UnixSignal) {
		case 17,18,20:
			fmt.Println("See signal", sig)
		default:
			fmt.Println("Exiting due to signal", sig)
			return
		}
	}
}

var paths = [...]string{"/bin/fusermount", "/usr/bin/fusermount"}

func UnmountFuse(mountpoint string) (err os.Error) {
	var pid int
	fmt.Println("Unmounting", mountpoint,"...")
	for i := 0; i < len(paths); i++ {
		var args []string = []string{paths[i], "-u", mountpoint}
		pid, err = os.ForkExec(paths[i], args, []string{}, "",
			[]*os.File{os.Stdin, os.Stdout, os.Stderr})
		if err == nil {
			os.Wait(pid, 0)
			return err
		}
	}
	return err
}
