package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var mcLocation string
	var found bool
	var srcCount int
	var srcList, destList []string

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("[RITO SHADERS INIT]: Failed to determine the script directory:", err)
		return
	}

	placeholderFile := filepath.Join(dir, "materials", "putMaterialsHere")
	if _, err := os.Stat(placeholderFile); err == nil {
		os.Remove(placeholderFile)
	}

	displayIntro()

	if !confirm("[RITO SHADERS INIT]: Is IObit Unlocker installed on your system? (Y=Yes, N=No)") {
		fmt.Println("[RITO SHADERS INIT]: Redirecting you to the IObit Unlocker download page...")
		exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://www.iobit.com/en/iobit-unlocker.php").Start()
		time.Sleep(3 * time.Second)
		return
	}

	if !confirm("[RITO SHADERS INIT]: Have you unlocked the WindowsApps folder? (Y=Yes, N=No)") {
		if !confirm("[RITO SHADERS INIT]: Unlocking the WindowsApps folder might take a while depending on your system. Proceed? (Y=Yes, N=No)") {
			fmt.Println("[RITO SHADERS INIT]: Shader injection requires the WindowsApps folder to be unlocked.")
			return
		}

		unlockWindowsApps(dir)
	}

	mcLocation, found = findMinecraftLocation()
	if !found {
		fmt.Println("[RITO SHADERS INIT]: Minecraft installation not found in 'C:\\Program Files\\WindowsApps'. Please ensure Minecraft is installed.")
		return
	}

	if confirm("[RITO SHADERS INIT]: Would you like to back up the original materials? (Y=Yes, N=No)") {
		backupMaterials(mcLocation, dir)
	}

	srcList, destList, srcCount = findBinFiles(dir, mcLocation)
	if srcCount == 0 {
		fmt.Println("[RITO SHADERS INIT]: No .bin files detected. Please add .bin files to the /materials folder.")
		return
	}

	displayMaterialList(srcList, mcLocation)

	if !confirm("[RITO SHADERS INIT]: Ready to inject the new materials? (Y=Yes, N=No)") {
		fmt.Println("[RITO SHADERS INIT]: Operation canceled.")
		return
	}

	deleteVanillaMaterials(destList)

	moveSourceMaterials(srcList, mcLocation)

	fmt.Println("[RITO SHADERS INIT]: Injection completed successfully!")
}

func displayIntro() {
	fmt.Println("\n[RITO SHADERS INIT]:\nThis tool injects `.material.bin` files into Minecraft Bedrock Edition.")
	time.Sleep(3 * time.Second)
}

func confirm(prompt string) bool {
	fmt.Println(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y"
}

func unlockWindowsApps(dir string) {
	for {
		fmt.Println("[RITO SHADERS INIT]: Attempting to unlock WindowsApps folder...")
		cmd := exec.Command("powershell", "-command", "start-process", "-file", "takeOwnership.bat", "-verb", "runas", "-Wait")
		cmd.Dir = dir
		cmd.Run()

		if _, err := os.Stat(filepath.Join(dir, "claimedOwnership.txt")); err == nil {
			break
		} else {
			fmt.Println("[RITO SHADERS INIT]: UAC prompt not accepted. Retrying...")
		}
	}

	fmt.Println("[RITO SHADERS INIT]: WindowsApps folder unlocked successfully!")
	time.Sleep(2 * time.Second)
}

func findMinecraftLocation() (string, bool) {
	programFiles := os.Getenv("ProgramFiles")
	mcPattern := filepath.Join(programFiles, "WindowsApps", "Microsoft.MinecraftUWP_*")
	matches, err := filepath.Glob(mcPattern)
	if err != nil || len(matches) == 0 {
		return "", false
	}

	return matches[0], true
}

func backupMaterials(mcLocation, dir string) {
	src := filepath.Join(mcLocation, "data", "renderer", "materials")
	dest := filepath.Join(dir, "materials.backup")

	fmt.Println("[RITO SHADERS INIT]: Backing up original materials...")
	exec.Command("xcopy", src, dest, "/E", "/I", "/H", "/Y").Run()
	fmt.Println("[RITO SHADERS INIT]: Backup completed!")
	time.Sleep(2 * time.Second)
}

func findBinFiles(dir, mcLocation string) ([]string, []string, int) {
	var srcList, destList []string
	materialsDir := filepath.Join(dir, "materials")
	files, err := os.ReadDir(materialsDir)
	if err != nil {
		fmt.Println("[RITO SHADERS INIT]: Failed to read the materials directory:", err)
		return nil, nil, 0
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".bin") {
			srcFile := filepath.Join(materialsDir, file.Name())
			destFile := filepath.Join(mcLocation, "data", "renderer", file.Name())
			srcList = append(srcList, srcFile)
			destList = append(destList, destFile)
		}
	}

	return srcList, destList, len(srcList)
}

func displayMaterialList(srcList []string, mcLocation string) {
	fmt.Printf("[RITO SHADERS INIT]: %d .bin file(s) detected in the materials folder.\n", len(srcList))
	fmt.Println("[RITO SHADERS INIT]: Minecraft installation found at:", mcLocation)
	fmt.Println("[RITO SHADERS INIT]: -------- Materials to Inject --------")
	for _, src := range srcList {
		fmt.Println(filepath.Base(src))
	}
	fmt.Println("[RITO SHADERS INIT]: -------------------------------------")
}

func deleteVanillaMaterials(destList []string) {
	fmt.Println("[RITO SHADERS INIT]: Removing original materials...")
	for _, dest := range destList {
		exec.Command("IObitUnlocker", "/advanced", "/delete", dest).Run()
	}
}

func moveSourceMaterials(srcList []string, mcLocation string) {
	fmt.Println("[RITO SHADERS INIT]: Injecting new materials...")
	for _, src := range srcList {
		dest := filepath.Join(mcLocation, "data", "renderer", "materials", filepath.Base(src))
		exec.Command("IObitUnlocker", "/advanced", "/move", src, dest).Run()
	}
}
