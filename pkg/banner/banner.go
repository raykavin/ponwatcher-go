package banner

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/google/goterm/term"
)

var fonts = []string{
	"lean",
	"larry3d",
	"nipples",
	"doom",
	"graffiti",
}

func PrintBanner(appName string) {
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomIndex := randSource.Intn(len(fonts))

	banner := figure.NewColorFigure(appName, fonts[randomIndex], "cyan", true)

	banner.Print()
	delayedPrint(strings.Repeat("-", 60))
}

func PrintText(text string) {
	t := term.Bold(term.Cyan(text))
	delayedPrint(t.String())
}

func PrintHeader(content string) {
	delayedPrint(strings.Repeat("-", 60))
	osName := runtime.GOOS
	arch := runtime.GOARCH
	numCPU := runtime.NumCPU()
	linuxDist := linuxDist()

	delayedPrint(term.Bold(term.Greenf("%s", content)).String())
	delayedPrint(term.Bold(term.Greenf("O.S Name: %s", osName)).String())
	delayedPrint(term.Bold(term.Greenf("Architecture: %s", arch)).String())
	delayedPrint(term.Bold(term.Greenf("CPU's Available: %d", numCPU)).String())

	if linuxDist != "" {
		delayedPrint(term.Bold(term.Greenf("O.S Dist. Name: %s", linuxDist)).String())
	}

	host, err := os.Hostname()
	if err == nil {
		delayedPrint(term.Bold(term.Greenf("Host Name: %s", host)).String())
	}

	kernelVersion, err := exec.Command("uname", "-r").Output()
	if err == nil {
		delayedPrint(term.Bold(term.Greenf("Kernel Version: %s", strings.Trim(string(kernelVersion), "\n"))).String())
	}

	delayedPrint(strings.Repeat("-", 60))

}

func linuxDist() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		fmt.Printf("Error reading /etc/os-release: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			distributionName := strings.Trim(line[len("NAME="):], "\"")
			return distributionName
		}
	}

	return ""

}

func delayedPrint(text string, delay ...time.Duration) {
	t := 2 * time.Millisecond

	if len(delay) > 0 {
		t = delay[0]
	}

	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(t)
	}

	fmt.Println()
}
