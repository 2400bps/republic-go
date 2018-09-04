package testutils

import (
	"os"
	"strconv"

	"github.com/onsi/ginkgo"
)

const reset = "\x1b[0m"
const green = "\x1b[32;1m"
const yellow = "\x1b[33;1m"

// GetCIEnv returns true if the CI environment variable is set
func GetCIEnv() bool {
	ciEnv := os.Getenv("CI")
	ci, err := strconv.ParseBool(ciEnv)
	if err != nil {
		ci = false
	}
	return ci
}

// GanacheContext can be used instead of Context to skip tests when they are
// being run in a CI environment (to avoid getting flagged for running Bitcoin
// mining software, and Ganache software).
func GanacheContext(description string, f func()) bool {
	if GetCIEnv() {
		return ginkgo.PContext(description, func() {
			ginkgo.It("Skipping ganache tests...", func() {})
		})
	}
	return ginkgo.Context(description, f)
}

// SkipCIBeforeSuite skips the BeforeSuite, which runs even if there are no tests
func SkipCIBeforeSuite(f func()) bool {
	if !GetCIEnv() {
		return ginkgo.BeforeSuite(f)
	}
	return false
}

// SkipCIAfterSuite skips the AfterSuite, which runs even if there are no tests
func SkipCIAfterSuite(f func()) bool {
	if !GetCIEnv() {
		return ginkgo.AfterSuite(f)
	}
	return false
}

func SkipCIDescribe(d string, f func()) bool {
	if !GetCIEnv() {
		return ginkgo.Describe(d, f)
	}
	return false
}

func GanacheAfterEach(body interface{}, timeout ...float64) bool {
	return ginkgo.AfterEach(body, timeout...)
}
