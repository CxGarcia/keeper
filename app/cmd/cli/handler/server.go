package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "keeper/internal/logger"

	"github.com/urfave/cli/v2"
)

const (
	lockFile    = "keeper.lock"
	processFile = "process.json"
)

type ProcessInfo struct {
	PID        int       `json:"pid"`
	StartTime  time.Time `json:"start_time"`
	IsDetached bool      `json:"is_detached"`
	Port       string    `json:"port"`
}

func (h *Handler) startServer(c *cli.Context) error {
	addr := fmt.Sprintf(":%s", c.String("port"))
	detached := c.Bool("detached")

	if info, err := h.getRunningServerInfo(); err == nil {
		return log.Errorf("server is already running with PID %d on port %s", info.PID, info.Port)
	}

	if detached {
		return h.startDetached(addr)
	}

	if err := h.acquireLock(); err != nil {
		return log.Errorf("failed to acquire lock: %w", err)
	}

	defer h.releaseLock()

	info := ProcessInfo{
		PID:        os.Getpid(),
		StartTime:  time.Now(),
		IsDetached: false,
		Port:       c.String("port"),
	}

	if err := h.writeProcessInfo(info); err != nil {
		return log.Errorf("failed to write process info: %w", err)
	}

	if err := h.proxyService.Start(addr); err != nil {
		os.Remove(processFile)
		return log.Errorf("error starting server: %v", err)
	}

	return nil
}

func (h *Handler) stopServer(c *cli.Context) error {
	info, err := h.getRunningServerInfo()
	if err != nil {
		if os.IsNotExist(err) {
			return log.Errorf("server is not running")
		}
		return log.Errorf("failed to read process info: %w", err)
	}

	process, err := os.FindProcess(info.PID)
	if err != nil {
		return log.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return log.Errorf("failed to send termination signal: %w", err)
	}

	if err := os.Remove(processFile); err != nil {
		return log.Errorf("failed to remove process info file: %w", err)
	}

	log.Infof("Server (PID: %d, Port: %s) stopped successfully\n", info.PID, info.Port)

	return nil
}

func (h *Handler) statusServer(c *cli.Context) error {
	info, err := h.getRunningServerInfo()
	if err != nil {
		log.Infof("server is not running")
		return nil
	}

	log.Infof("Server status:")
	log.Infof("  PID: %d", info.PID)
	log.Infof("  Start Time: %s", info.StartTime.Format(time.RFC3339))
	log.Infof("  Detached: %t", info.IsDetached)
	log.Infof("  Port: %s", info.Port)

	return nil
}

func (h *Handler) getRunningServerInfo() (*ProcessInfo, error) {
	data, err := os.ReadFile(processFile)
	if err != nil {
		return nil, err
	}

	var info ProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process info: %w", err)
	}

	if err := syscall.Kill(info.PID, 0); err != nil {
		return nil, err
	}

	return &info, nil
}

func (h *Handler) acquireLock() error {
	_, err := os.Create(lockFile)

	return err
}

func (h *Handler) releaseLock() error {
	return os.Remove(lockFile)
}

func (h *Handler) writeProcessInfo(info ProcessInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal process info: %w", err)
	}
	return os.WriteFile(processFile, data, 0644)
}

func (h *Handler) startDetached(addr string) error {
	cmd := exec.Command(os.Args[0], "start", "--port", strings.TrimPrefix(addr, ":"))
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return log.Errorf("failed to start detached process: %w", err)
	}

	// wait a short time to allow the child process to start and write its info
	time.Sleep(100 * time.Millisecond)

	info, err := h.getRunningServerInfo()
	if err == nil {
		info.IsDetached = true
		if err := h.writeProcessInfo(*info); err != nil {
			log.Errorf("failed to update process info: %v", err)
		}
		log.Infof("Server started in detached mode. PID: %d, Port: %s\n", info.PID, info.Port)
	} else {
		log.Errorf("Server started in detached mode, but unable to read process info: %v", err)
	}

	return nil
}
