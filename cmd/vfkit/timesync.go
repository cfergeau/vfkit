package main

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/Code-Hex/vz/v3"
	"github.com/crc-org/vfkit/pkg/config"
	"github.com/crc-org/vfkit/pkg/vf"
	sleepnotifier "github.com/prashantgupta24/mac-sleep-notifier/notifier"
	log "github.com/sirupsen/logrus"
)

type timeSyncer struct {
	vsockConn net.Conn
	vm        *vz.VirtualMachine
	vsockPort uint
}

func newTimeSyncer(vm *vz.VirtualMachine, vsockPort uint) *timeSyncer {
	return &timeSyncer{
		vm:        vm,
		vsockPort: vsockPort,
	}
}

func (ts *timeSyncer) conn() (net.Conn, error) {
	if ts.vm == nil || ts.vsockPort == 0 {
		return nil, fmt.Errorf("timeSyncer is in an invalid state")
	}

	if ts.vsockConn != nil {
		return ts.vsockConn, nil
	}
	vsockConn, err := vf.ConnectVsockSync(ts.vm, ts.vsockPort)
	if err != nil {
		return nil, fmt.Errorf("error connecting to vsock port %d: %v", ts.vsockPort, err)
	}
	ts.vsockConn = vsockConn

	return ts.vsockConn, nil
}

func (ts *timeSyncer) Close() error {
	if ts.vsockConn != nil {
		return ts.vsockConn.Close()
	}

	return nil
}

func (ts *timeSyncer) syncGuestTime() error {
	conn, err := ts.conn()
	if err != nil {
		return err
	}
	qemugaCmdTemplate := `{"execute": "guest-set-time", "arguments":{"time": %d}}` + "\n"
	qemugaCmd := fmt.Sprintf(qemugaCmdTemplate, time.Now().UnixNano())

	log.Debugf("sending %s to qemu-guest-agent", qemugaCmd)
	_, err = conn.Write([]byte(qemugaCmd))
	if err != nil {
		return err
	}
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}

	if response != `{"return": {}}`+"\n" {
		return fmt.Errorf("Unexpected response from qemu-guest-agent: %s", response)
	}

	return nil
}
func (ts *timeSyncer) watchWakeupNotifications() {
	sleepNotifierCh := sleepnotifier.GetInstance().Start()
	for {
		select {
		case activity := <-sleepNotifierCh:
			log.Debugf("Sleep notification: %s", activity)
			if activity.Type == sleepnotifier.Awake {
				log.Infof("machine awake")
				if err := ts.syncGuestTime(); err != nil {
					log.Debugf("error syncing guest time: %v", err)
				}
			}
		}
	}
}

func setupGuestTimeSync(vm *vz.VirtualMachine, timesync *config.TimeSync) error {
	if timesync == nil {
		return nil
	}

	log.Infof("Setting up host/guest time synchronization")

	timeSyncer := newTimeSyncer(vm, timesync.VsockPort())
	go timeSyncer.watchWakeupNotifications()

	return nil
}
