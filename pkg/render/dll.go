package render

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// findProjectRoot 从指定目录向上查找 go.mod 文件，确定项目根目录
func findProjectRoot(from string) (string, bool) {
	dir := from
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// collectRoots 收集所有可能的项目根目录候选
func collectRoots() []string {
	seen := map[string]bool{}
	var roots []string

	if wd, err := os.Getwd(); err == nil {
		if root, ok := findProjectRoot(wd); ok && !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}

	if exe, err := os.Executable(); err == nil {
		if root, ok := findProjectRoot(filepath.Dir(exe)); ok && !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}

	return roots
}

// findDLLDir 查找包含指定 DLL 文件的目录
func findDLLDir(dllName string, subDirs []string) string {
	if env := os.Getenv("SCRCPY_DLL_DIR"); env != "" {
		if info, err := os.Stat(filepath.Join(env, dllName)); err == nil && !info.IsDir() {
			return env
		}
	}

	for _, root := range collectRoots() {
		for _, sub := range subDirs {
			dir := filepath.Join(root, sub)
			if info, err := os.Stat(filepath.Join(dir, dllName)); err == nil && !info.IsDir() {
				return dir
			}
		}
	}

	return ""
}

var loadedDLLDirs []string

// addDllDirectory 将目录加入 Windows DLL 搜索路径
func addDllDirectory(dir string) {
	if dir == "" {
		return
	}
	// 避免重复添加
	for _, d := range loadedDLLDirs {
		if d == dir {
			return
		}
	}
	loadedDLLDirs = append(loadedDLLDirs, dir)

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("AddDllDirectory")
	ptr, _ := syscall.UTF16PtrFromString(dir)
	proc.Call(uintptr(unsafe.Pointer(ptr)))
}

// loadDLL 加载指定路径的 DLL，确保依赖可找到
func loadDLL(path string) *syscall.LazyDLL {
	// 先添加 DLL 所在目录到搜索路径
	dir := filepath.Dir(path)
	absDir, _ := filepath.Abs(dir)
	addDllDirectory(absDir)

	// 创建 LazyDLL，触发加载
	dll := syscall.NewLazyDLL(path)
	if err := dll.Load(); err != nil {
		// 如果失败，切换当前目录到 DLL 目录再试
		oldDir, _ := os.Getwd()
		os.Chdir(absDir)
		dll = syscall.NewLazyDLL(filepath.Base(path))
		dll.Load()
		os.Chdir(oldDir)
	}
	return dll
}

// findFFmpegDir 查找 FFmpeg DLL 目录
func findFFmpegDir() string {
	dir := findDLLDir("avcodec-62.dll", []string{
		"data/bin",
		"data/scrcpy-win64-v4.1",
	})
	if dir != "" {
		addDllDirectory(dir)
	}
	return dir
}

// findSDL3Dir 查找 SDL3 DLL 目录
func findSDL3Dir() string {
	dir := findDLLDir("SDL3.dll", []string{
		"data/scrcpy-win64-v4.1",
	})
	if dir != "" {
		addDllDirectory(dir)
	}
	return dir
}