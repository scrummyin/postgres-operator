/* file: $GOPATH/src/godogs/godogs_test.go */
package pgo_godogs

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

type CmdContext struct {
	stdout       *string
	stderr       *string
	stdin        *string
	binaryToCall string
	cmdArgs      []string
}

var cmdContext = new(CmdContext)

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func (c *CmdContext) resetTerminal(i interface{}) {
	c.stdout = nil
	c.stderr = nil
	c.stdin = nil
	c.binaryToCall = ""
	c.cmdArgs = nil
}

func (c *CmdContext) runCliCommandAndCheckStderr() error {
	if c.binaryToCall == "" {
		return fmt.Errorf("The terminal command to call is undefined")
	}
	cmd := exec.Command(c.binaryToCall, c.cmdArgs...)
	cmd.Env = append(os.Environ())
	var bufOut, bufErr bytes.Buffer
	cmd.Stdout = &bufOut
	cmd.Stderr = &bufErr
	if c.stdin != nil {
		cmd.Stdin = bytes.NewBufferString(*(c.stdin))
	}
	err := cmd.Run()
	if err != nil {
		return err
	}
	c.stdout = bufferToPointerString(bufOut)
	c.stderr = bufferToPointerString(bufErr)
	return err
}

func (c *CmdContext) ensureNoStderr() error {
	var err error = nil
	if (c.stderr) == nil {
		return nil
	}
	if len(*(c.stderr)) > 0 {
		err = fmt.Errorf("Command errored with stderr of: %s",
			*(c.stderr),
		)
	}
	return err
}

func (c *CmdContext) runPgoCommand(command string) error {
	c.binaryToCall = "pgo"
	c.cmdArgs = strings.Split(command, " ")
	return c.runCliCommandAndCheckStderr()
}

func (c *CmdContext) createClusterWithPgbouncer(name string) error {
	c.runPgoCommand("create cluster " + name + " --pgbouncer")
	return c.ensureNoStderr()
}

func (c *CmdContext) createCluster(name string) error {
	c.runPgoCommand("create cluster " + name)
	return c.ensureNoStderr()
}

func (c *CmdContext) deleteAllClusters() error {
	c.cmdArgs = []string{"delete", "cluster", "all", "--no-prompt"}
	c.binaryToCall = "pgo"
	return c.runCliCommandAndCheckStderr()
}

func (c *CmdContext) compareClientAndServerVersions() error {
	if len(*(c.stdout)) <= 0 {
		return fmt.Errorf(
			"Command returned nothing to stdout",
		)
	}
	var lines []string = strings.Split(strings.TrimSuffix(*(c.stdout), "\n"), "\n")
	if lines == nil || len(lines) != 2 {
		return fmt.Errorf(
			"Expected command to return 2 lines it returned %d",
			len(lines),
		)
	}
	var pgoCliVersion = strings.Split(lines[0], " ")[3]
	var pgoServerVersion = strings.Split(lines[1], " ")[2]
	if strings.Compare(pgoCliVersion, pgoServerVersion) != 0 {
		return fmt.Errorf(
			"Client Server version mismatch cli (%s) and server (%s)",
			pgoCliVersion,
			pgoServerVersion,
		)
	}
	return nil
}

func (c *CmdContext) checkStdoutContains(text string) error {
	if !strings.Contains(*(c.stdout), text) {
		return fmt.Errorf(
			"Pgo out did not contain (%s) and was: \n %s",
			*(c.stdout),
			text,
		)
	}
	return nil
}

func checkForRunningPod(labels string) error {
	var labelsAsArgs = []string{"-l", labels}
	return waitForPod(labelsAsArgs)
}

func checkForRunningPrimaryPod(labels string) error {
	var labelsWithPrimary = "primary=true," + labels
	return checkForRunningPod(labelsWithPrimary)
}

func (c *CmdContext) runPgoWithStdin(text string, input string) error {
	c.stdin = &input
	return c.runPgoCommand(text)
}

//Yes sadly go returns a "<nil>" instead of nil if there is a nil
func bufferToPointerString(buffer bytes.Buffer) *string {
	var result = buffer.String()
	if result == "<nil>" {
		return nil
	}
	return &result
}

func noPodsWithLabelShouldExist(labels string) error {
	var labelsAsArgs = []string{"-l", labels}
	return waitForPodsToDie(labelsAsArgs)
}

func waitForPodsToDie(labels []string) error {
	c := CmdContext{binaryToCall: "kubectl", cmdArgs: []string{"get", "pod"}}
	c.cmdArgs = append(c.cmdArgs, labels...)
	var sleepDuration, err = time.ParseDuration("3s")
	if err != nil {
		return err
	}
	for tries := 0; tries <= 30; tries++ {
		time.Sleep(sleepDuration)
		err := c.runCliCommandAndCheckStderr()
		if err != nil {
			return err
		}
		if strings.Contains(*(c.stdout), "No resources found.") {
			break
		}
	}
	return nil
}

func waitForPod(labels []string) error {
	c := CmdContext{binaryToCall: "kubectl", cmdArgs: []string{"get", "pod"}}
	c.cmdArgs = append(c.cmdArgs, labels...)
	var sleepDuration, err = time.ParseDuration("3s")
	if err != nil {
		return err
	}
	for tries := 0; tries <= 30; tries++ {
		time.Sleep(sleepDuration)
		err := c.runCliCommandAndCheckStderr()
		if err != nil {
			return err
		}
		if strings.Contains(*(c.stdout), "Running") && strings.Contains(*(c.stdout), "1/1") {
			break
		}
		if strings.Contains(*(c.stdout), "Terminating") || strings.Contains(*(c.stdout), "Error") {
			return fmt.Errorf("Pod has entered an unexpected state of : %s",
				*(c.stdout),
			)
		}
	}
	if !(strings.Contains(*(c.stdout), "Running") || strings.Contains(*(c.stdout), "1/1")) {
		return fmt.Errorf("Pod has entered an unexpected state of : %s",
			*(c.stdout),
		)
	}
	return nil
}

func FeatureContext(s *godog.Suite) {
	c := &CmdContext{}

	s.BeforeScenario(c.resetTerminal)

	s.Step(`^I run "pgo ([^"]+)"$`, c.runPgoCommand)
	s.Step(`^I create a cluster named (\w+)$`, c.createCluster)
	s.Step(`^I create a cluster named (\w+) with pgbouncer$`, c.createClusterWithPgbouncer)
	s.Step(`^There should be matching version info for both client and server$`, c.compareClientAndServerVersions)
	s.Step(`^A primary pod labeled with "([^"]+)" should be up$`, checkForRunningPrimaryPod)
	s.Step(`^No pods with label "([^"]*)" should exist$`, noPodsWithLabelShouldExist)
	s.Step(`^A pod labeled with "([^"]+)" should be up$`, checkForRunningPod)
	s.Step(`^An existing cluster named (\w+)$`, c.createCluster)
	s.Step(`^Then pgo should have stdout containing "([^"]+)"$`, c.checkStdoutContains)
	s.Step(`^No clusters are currently running$`, c.deleteAllClusters)
	s.Step(`^I run "pgo ([^"]+)" and type "([^"]+)"$`, c.runPgoWithStdin)

	s.AfterSuite(func() {
		deletePodContext := CmdContext{binaryToCall: "pgo",
			cmdArgs: []string{"delete", "cluster", "all", "--no-prompt"}}
		err := deletePodContext.runCliCommandAndCheckStderr()
		if err != nil {
			deletePVCContext := CmdContext{binaryToCall: "kubectl",
				cmdArgs: []string{"delete", "pvc", "-l", "pg-cluster"}}
			err = deletePVCContext.runCliCommandAndCheckStderr()
		}
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
