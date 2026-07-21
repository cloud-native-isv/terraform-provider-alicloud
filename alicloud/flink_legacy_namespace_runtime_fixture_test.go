//go:build flink_legacy_namespace_runtime_fixture
// +build flink_legacy_namespace_runtime_fixture

package alicloud

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

const (
	flinkLegacyRuntimeLedgerHelperEnv        = "FLINK_LEGACY_RUNTIME_LEDGER_HELPER"
	flinkLegacyRuntimeLedgerHelperPathEnv    = "FLINK_LEGACY_RUNTIME_LEDGER_HELPER_PATH"
	flinkLegacyRuntimeLedgerHelperIterations = "FLINK_LEGACY_RUNTIME_LEDGER_HELPER_ITERATIONS"
)

func TestFlinkLegacyRuntimeFixtureLedgerCrossProcessSerialization(t *testing.T) {
	if os.Getenv(flinkLegacyRuntimeLedgerHelperEnv) == "1" {
		iterations, err := strconv.Atoi(os.Getenv(flinkLegacyRuntimeLedgerHelperIterations))
		if err != nil || iterations <= 0 {
			t.Fatalf("invalid helper iterations: %q", os.Getenv(flinkLegacyRuntimeLedgerHelperIterations))
		}
		path := os.Getenv(flinkLegacyRuntimeLedgerHelperPathEnv)
		for i := 0; i < iterations; i++ {
			if _, err := flinkLegacyRuntimeFixtureMutate(path, func(state *flinkLegacyRuntimeFixtureLedger) {
				state.WorkspacePolls++
			}); err != nil {
				t.Fatalf("helper mutate %d: %v", i, err)
			}
		}
		return
	}

	path := t.TempDir() + "/ledger.json"
	createFlinkLegacyTofuRuntimeLedgerLock(t, path)
	writeFlinkLegacyTofuRuntimeLedger(t, path, flinkLegacyTofuRuntimeLedger{})

	const processes = 6
	const iterations = 300
	type helperProcess struct {
		command *exec.Cmd
		output  *bytes.Buffer
	}
	commands := make([]helperProcess, 0, processes)
	for i := 0; i < processes; i++ {
		command := exec.Command(os.Args[0], "-test.run=^TestFlinkLegacyRuntimeFixtureLedgerCrossProcessSerialization$")
		command.Env = append(os.Environ(),
			flinkLegacyRuntimeLedgerHelperEnv+"=1",
			flinkLegacyRuntimeLedgerHelperPathEnv+"="+path,
			flinkLegacyRuntimeLedgerHelperIterations+"="+strconv.Itoa(iterations),
		)
		output := &bytes.Buffer{}
		command.Stdout = output
		command.Stderr = output
		if err := command.Start(); err != nil {
			t.Fatalf("start helper %d: %v", i, err)
		}
		commands = append(commands, helperProcess{command: command, output: output})
	}
	for i, helper := range commands {
		if err := helper.command.Wait(); err != nil {
			t.Fatalf("helper %d: %v\n%s", i, err, helper.output.Bytes())
		}
	}

	state, err := flinkLegacyRuntimeFixtureRead(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := processes * iterations; state.WorkspacePolls != want {
		t.Fatalf("cross-process mutations retained %d increments, want %d; stale load-modify-save snapshots lost updates", state.WorkspacePolls, want)
	}
}

func TestFlinkLegacyRuntimeFixtureLedgerLockFailsClosed(t *testing.T) {
	t.Run("missing lock", func(t *testing.T) {
		path := t.TempDir() + "/ledger.json"
		initial := []byte("{\"workspace_polls\":7}\n")
		if err := os.WriteFile(path, initial, 0o600); err != nil {
			t.Fatal(err)
		}

		if _, err := flinkLegacyRuntimeFixtureMutate(path, func(state *flinkLegacyRuntimeFixtureLedger) {
			state.WorkspacePolls++
		}); err == nil || !strings.Contains(err.Error(), "open legacy runtime fixture ledger lock") {
			t.Fatalf("mutate without lock error = %v, want missing-lock failure", err)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != string(initial) {
			t.Fatalf("failed mutation changed ledger to %q, want %q", body, initial)
		}
	})

	t.Run("insecure lock", func(t *testing.T) {
		path := t.TempDir() + "/ledger.json"
		if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path+flinkLegacyRuntimeFixtureLedgerLockSuffix, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path+flinkLegacyRuntimeFixtureLedgerLockSuffix, 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := flinkLegacyRuntimeFixtureRead(path); err == nil || !strings.Contains(err.Error(), "stable mode 0600 regular file") {
			t.Fatalf("read with insecure lock error = %v, want fail-closed mode rejection", err)
		}
	})
}
