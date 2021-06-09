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

// CommandExecWrapper wrap exec.Command to log to stdoutPipe and stderrPipe
func CommandExecWrapper(cmd *exec.Cmd) error {
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	go func() {
		for {
			reader := bufio.NewReader(stdoutPipe)
			line, err := reader.ReadString('\n')
			for err == nil {
				fmt.Println(line)
				line, err = reader.ReadString('\n')
			}
		}
	}()
	go func() {
		for {
			reader := bufio.NewReader(stderrPipe)
			line, err := reader.ReadString('\n')
			for err == nil {
				fmt.Println(line)
				line, err = reader.ReadString('\n')
			}
		}
	}()

	if err := cmd.Run(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
