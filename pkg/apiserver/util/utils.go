package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

func GenerateID() string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)
}

func GetVersionEndpoint(router *chi.Mux, gitCommit string) {
	router.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		version := struct {
			Revision string
		}{
			Revision: gitCommit,
		}
		resp, err := json.Marshal(version)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			log.Error(err)
		}
	})
}

// CommandExecLogger wrap exec.Command to log to stdoutPipe and stderrPipe
func CommandExecLogger(cmd *exec.Cmd) error {
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	quitChan := make(chan struct{})
	if err := cmd.Start(); err != nil {
		log.Error(err)
		return err
	}

	go func() {
		for {
			reader := bufio.NewReader(stdoutPipe)
			line, err := reader.ReadString('\n')

			select {
			case <-quitChan:
				fmt.Println("asu exit out")
				log.Infoln("Closing logger for stdout")
				return
			default:
				if err == nil {
					fmt.Println(line)
					line, err = reader.ReadString('\n')
				} else {
					log.Error(err)
				}

			}
		}

	}()
	go func() {
		for {
			reader := bufio.NewReader(stderrPipe)
			line, err := reader.ReadString('\n')
			select {
			case <-quitChan:
				fmt.Println("asu exit err")
				log.Infoln("Closing logger for stderr")
				return
			default:
				if err == nil {
					fmt.Println(line)
					line, err = reader.ReadString('\n')
				} else {
					log.Error(err)
				}
			}
		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Error("error wait", err)
	}
	close(quitChan)
	return err
}

func GetCaller() string {
	if _, file, no, ok := runtime.Caller(1); ok {
		return fmt.Sprintf("called from %s:%d", file, no)
	}
	return ""
}
