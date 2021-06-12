// Fast and dirty implementation for synchronized file server across replica
package fileserver

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var peerFileServerHost chan string

// PeerDiscovery for very specific Kubernetes environment
func (fs *fileServer) PeerDiscovery() {
	host, _ := os.Hostname()

	peerFileServerHost := make(chan string, 3)
	go func() {
		for g := range peerFileServerHost {
			fmt.Println(g)
			peerFileServerHost <- g
		}
	}()

	//todo: number could be unlimited but need stop checking further if
	// few % of host checking failed
	for {
		for i := 0; i <= 2; i++ {
			if strings.Contains(host, strconv.Itoa(i)) {
				continue
			}
			go func(i int) {
				h := fmt.Sprintf("fileserver-%d", i)
				addr := fmt.Sprintf("%s:8001", h)
				timeout := time.Duration(1 * time.Second)
				_, err := net.DialTimeout("tcp", addr, timeout)
				if err != nil {
					fmt.Printf("%s %s %s\n", h, "not responding", err.Error())
				} else {
					log.Infof("Host %s found, adding to peer...", h)
					peerFileServerHost <- h
				}
			}(i)
		}
		time.Sleep(5 * time.Second)
	}
}

func (fs *fileServer) StartSync() {
	log.Infof("Start synchronizing")

	//var foundPeer map[string]string

	for _, host := range fs.peerFileServerHost {
		host := host

		go func() {
			//if h, found := foundPeer[host]; found {
			//	log.Debug(h) // to skip multiple same synchronization
			//}
			hostAddr := strings.Split(host, ":")[0]

			log.Infof("Start synchronizing with %s", hostAddr)
			cmd := exec.Command("bash", "script/lsyncd.sh", hostAddr, fs.pathToServe)
			log.Debug(cmd.String())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Errorf("Synchronizing with %s failed to start: %s", hostAddr, err)
			} else {
				//foundPeer[hostAddr] = hostAddr // mark host as found
			}
		}()
	}
}
