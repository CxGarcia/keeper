package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "keeper/internal/logger"

	"github.com/urfave/cli/v2"
)

const (
	pidFile  = "keeper.pid"
	lockFile = "keeper.lock"
)

func (h *Handler) startServer(c *cli.Context) error {
	addr := fmt.Sprintf(":%s", c.String("port"))
	detached := c.Bool("detached")

	// Check if the server is already running
	if pid, err := h.getRunningServerPID(); err == nil {
		return log.Errorf("server is already running with PID %d", pid)
	}

	if detached {
		return h.startDetached(addr)
	}

	if err := h.acquireLock(); err != nil {
		return log.Errorf("failed to acquire lock: %w", err)
	}

	defer h.releaseLock()

	if err := h.writePIDFile(); err != nil {
		return log.Errorf("failed to write PID file: %w", err)
	}

	if err := h.proxyService.Start(addr); err != nil {
		// Clean up PID file if server fails to start
		os.Remove("keeper.pid")
		return log.Errorf("error starting server: %v", err)
	}

	return nil
}

func (h *Handler) getRunningServerPID() (int, error) {
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return 0, err
	}

	if err := syscall.Kill(pid, 0); err != nil {
		return 0, err
	}

	return pid, nil
}

func (h *Handler) acquireLock() error {
	_, err := os.Create(lockFile)

	return err
}

func (h *Handler) releaseLock() error {
	return os.Remove(lockFile)
}

func (h *Handler) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func (h *Handler) startDetached(addr string) error {
	cmd := exec.Command(os.Args[0], "start", "--port", strings.TrimPrefix(addr, ":"))
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return log.Errorf("failed to start detached process: %w", err)
	}

	// Wait a short time to allow the child process to start and write its PID
	time.Sleep(100 * time.Millisecond)

	// Check if the PID file exists and read it
	if pid, err := h.getRunningServerPID(); err == nil {
		log.Infof("Server started in detached mode. PID: %d\n", pid)
	} else {
		log.Errorf("Server started in detached mode, but unable to read PID: %v", err)
	}

	return nil
}

func (h *Handler) stopServer(c *cli.Context) error {
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return log.Errorf("server is not running in detached mode")
		}
		return log.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return log.Errorf("invalid PID in file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return log.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return log.Errorf("failed to send termination signal: %w", err)
	}

	if err := os.Remove(pidFile); err != nil {
		return log.Errorf("failed to remove PID file: %w", err)
	}

	log.Infof("Server (PID: %d) stopped successfully\n", pid)

	return nil
}
