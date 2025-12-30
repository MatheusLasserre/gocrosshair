// Package main implements a lightweight, static crosshair overlay for X11/XWayland.
// It creates a transparent, click-through window that displays a crosshair at screen center.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"

	"gocrosshair/config"
	"gocrosshair/overlay"
	"gocrosshair/wizard"
)

// Version information (set during build with ldflags)
var (
	version = "dev"
)

const daemonEnvVar = "GOCROSSHAIR_DAEMON"

// daemonize forks the process to run in the background.
// Returns true if this is the parent process (should exit after forking).
func daemonize() bool {
	if os.Getenv(daemonEnvVar) == "1" {
		return false
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), daemonEnvVar+"=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start background process: %v", err)
	}

	fmt.Printf("✓ Crosshair started (PID %d)\n", cmd.Process.Pid)
	fmt.Println("  To stop: gocrosshair -stop")
	return true
}

// getPIDFilePath returns the path to the PID file.
func getPIDFilePath() string {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "gocrosshair.pid")
	}
	return "/tmp/gocrosshair.pid"
}

// writePIDFile writes the current process ID to the PID file.
func writePIDFile() error {
	pidPath := getPIDFilePath()
	return os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0644)
}

// removePIDFile removes the PID file.
func removePIDFile() {
	os.Remove(getPIDFilePath())
}

// readPIDFile reads the PID from the PID file.
func readPIDFile() (int, error) {
	pidPath := getPIDFilePath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

// isProcessRunning checks if a process with the given PID is running.
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds, so we need to send signal 0 to check
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// stopRunningInstance sends SIGTERM to any running gocrosshair instance.
func stopRunningInstance() error {
	pid, err := readPIDFile()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no running instance found (PID file does not exist)")
		}
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	if !isProcessRunning(pid) {
		removePIDFile()
		return fmt.Errorf("no running instance found (stale PID file removed)")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}

	fmt.Printf("Sent stop signal to gocrosshair (PID %d)\n", pid)
	return nil
}

// runSetupWizard runs the interactive configuration wizard.
// Returns true if user wants to start the crosshair after setup.
func runSetupWizard(cfgPath string) bool {
	monitors, err := getMonitorsForWizard()
	if err != nil {
		log.Printf("Warning: could not detect monitors: %v", err)
	}

	model := wizard.NewModel(monitors, cfgPath)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running setup wizard: %v", err)
	}

	if m, ok := finalModel.(wizard.Model); ok && m.WantsToStart() {
		return true
	}
	return false
}

// getMonitorsForWizard connects to X and retrieves monitor info for the wizard.
func getMonitorsForWizard() ([]wizard.Monitor, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)

	monitors, err := overlay.GetMonitors(conn, screen)
	if err != nil {
		return nil, err
	}

	wizardMonitors := make([]wizard.Monitor, len(monitors))
	for i, m := range monitors {
		wizardMonitors[i] = wizard.Monitor{
			Index:   i,
			Name:    m.Name,
			Width:   m.Width,
			Height:  m.Height,
			Primary: m.Primary,
		}
	}

	return wizardMonitors, nil
}

func main() {
	configPath := flag.String("config", "", "Path to configuration file (default: ~/.config/gocrosshair/config.toml)")
	listMonitors := flag.Bool("list-monitors", false, "List available monitors and exit")
	showVersion := flag.Bool("version", false, "Show version and exit")
	stopInstance := flag.Bool("stop", false, "Stop any running gocrosshair instance")
	runSetup := flag.Bool("setup", false, "Run interactive setup wizard to create configuration")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gocrosshair - Lightweight crosshair overlay for X11/XWayland\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfiguration file location:\n")
		fmt.Fprintf(os.Stderr, "  Default: ~/.config/gocrosshair/config.toml\n")
		fmt.Fprintf(os.Stderr, "  Override with -config flag or XDG_CONFIG_HOME environment variable\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("gocrosshair version %s\n", version)
		os.Exit(0)
	}

	if *stopInstance {
		if err := stopRunningInstance(); err != nil {
			log.Fatalf("Error: %v", err)
		}
		os.Exit(0)
	}

	if *listMonitors {
		if err := overlay.ListMonitors(); err != nil {
			log.Fatalf("Error: %v", err)
		}
		os.Exit(0)
	}

	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = config.GetConfigPath()
	}

	if *runSetup {
		if !runSetupWizard(cfgPath) {
			// User cancelled, exit
			os.Exit(0)
		}
		// Continue to start the crosshair after successful setup
	}

	if pid, err := readPIDFile(); err == nil && isProcessRunning(pid) {
		fmt.Printf("Crosshair is already running (PID %d).\n", pid)
		fmt.Println("To stop: gocrosshair -stop")
		os.Exit(1)
	}

	if daemonize() {
		os.Exit(0)
	}

	if err := writePIDFile(); err != nil {
		log.Printf("Warning: failed to write PID file: %v", err)
	}
	defer removePIDFile()

	cfg, created, err := config.LoadOrCreate(cfgPath)
	if err != nil {
		cfg, err = config.HandleInvalidConfig(cfgPath, err)
		if err != nil {
			log.Fatalf("Configuration error: %v", err)
		}
	}

	// Validate even if newly created, in case defaults change
	if !created {
		if err := cfg.Validate(); err != nil {
			cfg, err = config.HandleInvalidConfig(cfgPath, err)
			if err != nil {
				log.Fatalf("Configuration error: %v", err)
			}
		}
	}

	o, err := overlay.NewOverlay(cfg)
	if err != nil {
		log.Fatalf("Failed to create overlay: %v", err)
	}
	defer o.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		o.Close()
		os.Exit(0)
	}()

	if err := o.Run(); err != nil {
		log.Fatalf("Overlay error: %v", err)
	}
}
