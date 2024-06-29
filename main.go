package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	yellow  func(a ...interface{}) string
	blue    func(a ...interface{}) string
	magenta func(a ...interface{}) string
	green   func(a ...interface{}) string

	targetDir string
	sourceDir string
	tempDir   string
	dryRun    bool
)

func init() {
	yellow = color.New(color.FgYellow).SprintFunc()
	blue = color.New(color.FgBlue).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "move-resource [terraform resource address]",
		Short: "Move a terraform resource from source directory to target directory",
		Args:  cobra.ExactArgs(1),
		Run:   runTerraformMove,
	}

	rootCmd.Flags().StringVar(&targetDir, "target-dir", "", "Target directory to move a terraform resource to")
	rootCmd.Flags().StringVar(&sourceDir, "source-dir", "", "Source directory to move a terraform resource from (optional, defaults to CWD)")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", true, "If true, do not actually move the resource")

	rootCmd.MarkFlagRequired("target-dir")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const SourceStateFileName string = "terraform.tfstate"
const TargetStateFileName string = "target-state.tfstate"

func resolvePath(path string) string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return absPath
}

func runTerraformMove(cmd *cobra.Command, args []string) {
	sourceDir = resolvePath(sourceDir)
	targetDir = resolvePath(targetDir)

	resourceAddress := args[0]
	if sourceDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			os.Exit(1)
		}
		sourceDir = cwd
	}

	fmt.Printf(
		"Moving resource %s: \n  From state %s \n  To state   %s \n",
		yellow(resourceAddress),
		blue(sourceDir),
		blue(targetDir),
	)

	if dryRun {
		fmt.Printf("%s enabled. Not actually making moves. \n", magenta("[dry-run]"))
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to continue? (yes/no): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if !strings.EqualFold(input, "yes") {
			fmt.Println("Exiting...")
			os.Exit(1)
		}
	}

	newTempDir, err := ioutil.TempDir("", "terraform-move")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		os.Exit(1)
	}
	tempDir = newTempDir
	defer os.RemoveAll(tempDir)

	sourceStatePath := filepath.Join(tempDir, SourceStateFileName)
	targetStatePath := filepath.Join(tempDir, TargetStateFileName)

	fmt.Printf("==> %s\n", green("Initializing source state"))

	if err := terraformInit(sourceDir); err != nil {
		fmt.Println("Error running terraform init in source directory:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Pulling source state"))

	if err := pullState(sourceDir, sourceStatePath); err != nil {
		fmt.Println("Error pulling state from source directory:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Initializing target state"))

	if err := terraformInit(targetDir); err != nil {
		fmt.Println("Error running terraform init in target directory:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Pulling target state"))

	if err := pullState(targetDir, targetStatePath); err != nil {
		fmt.Println("Error pulling state from target directory:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Moving resource from source to target in local state"))

	if err := moveResource(tempDir, targetStatePath, resourceAddress); err != nil {
		fmt.Println("Error moving resource:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Copying updated state to target directory"))

	if err := copyFile(targetStatePath, filepath.Join(targetDir, TargetStateFileName)); err != nil {
		fmt.Println("Error copying target state file:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Pushing updated target state."))

	if err := pushState(targetDir, TargetStateFileName); err != nil {
		fmt.Println("Error pushing state to target directory:", err)
		os.Exit(1)
	}

	fmt.Printf("==> %s\n", green("Successfully moved!"))

	if !dryRun {
		err = os.Remove(filepath.Join(targetDir, TargetStateFileName))
		if err != nil {
			fmt.Println("Error cleaning up filepath:", err)
			os.Exit(1)
		}
	}
}

func terraformInit(dir string) error {
	if dryRun {
		fmt.Printf(
			"%s Would run 'terraform init': \n  from %s\n",
			magenta("[dry-run]"),
			dir,
		)
		return nil
	} else {
		cmd := exec.Command("terraform", "init")
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		return cmd.Run()
	}
}

func pullState(dir, outPath string) error {
	if dryRun {
		prettyOutPath := strings.Replace(outPath, tempDir, "$temp", 1)

		fmt.Printf(
			"%s Would run 'terraform state pull': \n  from %s\n  write to %s\n",
			magenta("[dry-run]"),
			dir,
			prettyOutPath,
		)

		return nil
	} else {
		cmd := exec.Command("terraform", "state", "pull")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, out, 0644)
	}
}

func moveResource(tempDir, statePath, resourceAddress string) error {
	if dryRun {
		prettyStatePath := strings.Replace(statePath, tempDir, "$temp", 1)

		fmt.Printf(
			"%s Would run 'terraform state mv --state-out=%s'\n",
			magenta("[dry-run]"),
			prettyStatePath,
		)

		return nil
	} else {
		cmd := exec.Command(
			"terraform",
			"state",
			"mv",
			"--state-out="+statePath,
			resourceAddress,
			resourceAddress,
		)
		cmd.Dir = tempDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

func copyFile(src, dst string) error {
	if dryRun {
		prettySrcPath := strings.Replace(src, tempDir, "$temp", 1)

		fmt.Printf(
			"%s Would copy mutated state file over to target state directory\n  from %s\n  to %s\n",
			magenta("[dry-run]"),
			prettySrcPath,
			dst,
		)

		return nil
	} else {
		input, err := ioutil.ReadFile(src)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(dst, input, 0644)
	}
}

func pushState(dir, stateFile string) error {
	if dryRun {
		fmt.Printf(
			"%s Would run 'terraform state push %s' from %s\n",
			magenta("[dry-run]"),
			TargetStateFileName,
			dir,
		)

		return nil
	} else {
		cmd := exec.Command("terraform", "state", "push", stateFile)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}
