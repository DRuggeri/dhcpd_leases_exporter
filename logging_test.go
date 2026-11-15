package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestBuildInfoMetric(t *testing.T) {
	metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, metric := range metrics {
		if metric.GetName() == "build_info" {
			if len(metric.GetMetric()) != 1 {
				t.Fatalf("expected one build info metric, got %d", len(metric.GetMetric()))
			}
			labels := map[string]bool{}
			for _, label := range metric.GetMetric()[0].GetLabel() {
				labels[label.GetName()] = true
			}
			for _, name := range []string{"version", "revision", "branch", "goversion"} {
				if !labels[name] {
					t.Errorf("build info missing label %q", name)
				}
			}
			return
		}
	}
	t.Fatal("build info metric not found")
}

func TestLoggingFlags(t *testing.T) {
	if os.Getenv("DHCPD_EXPORTER_LOG_TEST") == "1" {
		os.Args = []string{
			"dhcpd_leases_exporter",
			"--dhcpd.leases=" + os.Getenv("DHCPD_EXPORTER_TEST_LEASES"),
			"--log.level=" + os.Getenv("DHCPD_EXPORTER_TEST_LEVEL"),
			"--log.format=" + os.Getenv("DHCPD_EXPORTER_TEST_FORMAT"),
		}
		main()
		return
	}

	for _, tc := range []struct {
		name, level, format, want string
	}{
		{"json error", "error", "json", ""},
		{"invalid level", "fatal", "logfmt", "unrecognized log level"},
		{"legacy format", "info", "logger:stderr", "unrecognized log format"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestLoggingFlags$")
			cmd.Env = append(os.Environ(),
				"DHCPD_EXPORTER_LOG_TEST=1",
				"DHCPD_EXPORTER_TEST_LEASES="+filepath.Join(t.TempDir(), "missing"),
				"DHCPD_EXPORTER_TEST_LEVEL="+tc.level,
				"DHCPD_EXPORTER_TEST_FORMAT="+tc.format,
			)
			output, err := cmd.CombinedOutput()
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %v: %s", err, output)
			}
			if tc.want != "" {
				if !strings.Contains(string(output), tc.want) {
					t.Fatalf("expected %q in %s", tc.want, output)
				}
				return
			}
			var entry struct {
				Level string `json:"level"`
				Msg   string `json:"msg"`
			}
			if err := json.Unmarshal(output, &entry); err != nil {
				t.Fatalf("invalid JSON log %q: %v", output, err)
			}
			if entry.Level != "ERROR" || entry.Msg != "Failed to read leases" {
				t.Fatalf("unexpected log entry: %+v", entry)
			}
		})
	}
}
