package etcdmgmt

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gluster/glusterd2/tests"

	log "github.com/Sirupsen/logrus"
	heketitests "github.com/heketi/tests"
)

func formETCDCommand() *exec.Cmd {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not able to get hostname")
		return nil
	}

	listenClientUrls := "http://" + hostname + ":2379"

	advClientUrls := "http://" + hostname + ":2379"

	listenPeerUrls := "http://" + hostname + ":2380"

	initialAdvPeerUrls := "http://" + hostname + ":2380"

	etcdCmd := exec.Command("etcd",
		"-listen-client-urls", listenClientUrls,
		"-advertise-client-urls", advClientUrls,
		"-listen-peer-urls", listenPeerUrls,
		"-initial-advertise-peer-urls", initialAdvPeerUrls,
		"--initial-cluster", "default="+listenPeerUrls)

	return etcdCmd
}

func TestStartETCDWithInvalidExecName(t *testing.T) {
	// Mock the executable name such that it fails
	defer heketitests.Patch(&ExecName, "abc").Restore()
	_, err := StartETCD()
	tests.Assert(t, err != nil)
}

func TestStartETCD(t *testing.T) {
	etcdCtx, err := StartETCD()
	tests.Assert(t, err == nil)
	etcdCtx.Kill()
	etcdCtx.Wait()
}

func TestWriteETCDPidFile(t *testing.T) {
	cmd := formETCDCommand()
	_ = cmd.Start()
	tests.Assert(t, writeETCDPidFile(cmd.Process.Pid) == nil)
	os.Remove(etcdPidFile)

	// change etcdPidFile location such that its an invalid path and
	// writeETCDPidFile should fail
	defer heketitests.Patch(&etcdPidFile, "/a/b/c/d/etcd.pid").Restore()
	tests.Assert(t, writeETCDPidFile(cmd.Process.Pid) != nil)
	cmd.Process.Kill()
	cmd.Process.Wait()
}

func TestIsETCDStartNeeded(t *testing.T) {
	// check once etcd process is running isETCDStartNeeded returns false
	os.Remove(etcdPidFile)
	cmd := formETCDCommand()
	err := cmd.Start()
	tests.Assert(t, err == nil)
	err = writeETCDPidFile(cmd.Process.Pid)
	tests.Assert(t, err == nil)
	time.Sleep(15 * time.Second)
	start, _ := isETCDStartNeeded()
	tests.Assert(t, start == false)

	//check once etcd process is killed isETCDStartNeeded returns true
	err = cmd.Process.Kill()
	tests.Assert(t, err == nil)
	_, err = cmd.Process.Wait()
	tests.Assert(t, err == nil)
	start, _ = isETCDStartNeeded()
	tests.Assert(t, start == true)

	// check if the pid file is missing then isETCDStartNeeded returns true
	os.Remove(etcdPidFile)
	start, _ = isETCDStartNeeded()
	tests.Assert(t, start == true)
	cmd.Process.Kill()
	cmd.Process.Wait()
}
