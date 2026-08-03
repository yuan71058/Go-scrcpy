package render

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	sdl3 *syscall.LazyDLL

	procInit          *syscall.LazyProc
	procQuit          *syscall.LazyProc
	procCreateWindow  *syscall.LazyProc
	procDestroyWindow *syscall.LazyProc
	procCreateRenderer *syscall.LazyProc
	procDestroyRenderer *syscall.LazyProc
	procCreateTexture *syscall.LazyProc
	procDestroyTexture *syscall.LazyProc
	procUpdateTexture *syscall.LazyProc
	procUpdateYUVTexture *syscall.LazyProc
	procRenderClear   *syscall.LazyProc
	procRenderTexture *syscall.LazyProc
	procRenderPresent *syscall.LazyProc
	procPollEvent     *syscall.LazyProc
	procDelay         *syscall.LazyProc
	procGetTicks      *syscall.LazyProc
	procGetWindowSize *syscall.LazyProc
	procSetWindowTitle *syscall.LazyProc
	procSetWindowMouseGrab *syscall.LazyProc
	procGetKeyboardState *syscall.LazyProc
	procFree          *syscall.LazyProc
	procGetError      *syscall.LazyProc
)

const (
	SDL_INIT_VIDEO = 0x00000020
	SDL_TEXTUREACCESS_STREAMING = 1
	SDL_PIXELFORMAT_YV12 = 0x36315659
)

type SDLRenderer struct {
	window   uintptr
	renderer uintptr
	texture  uintptr
	width    int
	height   int
}

func NewSDLRenderer(title string, w, h int) (*SDLRenderer, error) {
	procInit.Call(SDL_INIT_VIDEO)

	titleBytes, _ := syscall.BytePtrFromString(title)
	win, _, _ := procCreateWindow.Call(
		uintptr(unsafe.Pointer(titleBytes)),
		uintptr(w), uintptr(h), 0,
	)
	if win == 0 {
		return nil, fmt.Errorf("创建窗口失败")
	}

	ren, _, _ := procCreateRenderer.Call(win, 0)
	if ren == 0 {
		procDestroyWindow.Call(win)
		return nil, fmt.Errorf("创建渲染器失败")
	}

	tex, _, _ := procCreateTexture.Call(
		ren,
		uintptr(SDL_PIXELFORMAT_YV12),
		uintptr(SDL_TEXTUREACCESS_STREAMING),
		uintptr(w), uintptr(h),
	)
	if tex == 0 {
		procDestroyRenderer.Call(ren)
		procDestroyWindow.Call(win)
		return nil, fmt.Errorf("创建纹理失败")
	}

	return &SDLRenderer{
		window:   win,
		renderer: ren,
		texture:  tex,
		width:    w,
		height:   h,
	}, nil
}

func (r *SDLRenderer) RenderYUV(yuv []byte) error {
	if len(yuv) == 0 {
		return nil
	}

	// SDL_UpdateYUVTexture: tex, rect, Y plane, Y stride, U plane, U stride, V plane, V stride
	ySize := r.width * r.height
	uvW := (r.width + 1) / 2
	uvH := (r.height + 1) / 2

	procUpdateYUVTexture.Call(
		r.texture,
		0,
		uintptr(unsafe.Pointer(&yuv[0])), uintptr(r.width),
		uintptr(unsafe.Pointer(&yuv[ySize])), uintptr(uvW),
		uintptr(unsafe.Pointer(&yuv[ySize+uvW*uvH])), uintptr(uvW),
	)

	procRenderClear.Call(r.renderer)
	procRenderTexture.Call(r.renderer, r.texture, 0, 0)
	procRenderPresent.Call(r.renderer)

	return nil
}

func (r *SDLRenderer) Destroy() {
	if r.texture != 0 {
		procDestroyTexture.Call(r.texture)
		r.texture = 0
	}
	if r.renderer != 0 {
		procDestroyRenderer.Call(r.renderer)
		r.renderer = 0
	}
	if r.window != 0 {
		procDestroyWindow.Call(r.window)
		r.window = 0
	}
	procQuit.Call()
}

func getLastError() string {
	ret, _, _ := procGetError.Call()
	if ret == 0 {
		return ""
	}
	return fmt.Sprintf("error code: %d", ret)
}

func init() {
	exe, _ := os.Executable()
	// mirror.exe 在 _examples/mirror/，向上三级到 project root
	root := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	dllDir := filepath.Join(root, "data", "scrcpy-win64-v4.1")
	sdl3 = syscall.NewLazyDLL(filepath.Join(dllDir, "SDL3.dll"))

	procInit = sdl3.NewProc("SDL_Init")
	procQuit = sdl3.NewProc("SDL_Quit")
	procCreateWindow = sdl3.NewProc("SDL_CreateWindow")
	procDestroyWindow = sdl3.NewProc("SDL_DestroyWindow")
	procCreateRenderer = sdl3.NewProc("SDL_CreateRenderer")
	procDestroyRenderer = sdl3.NewProc("SDL_DestroyRenderer")
	procCreateTexture = sdl3.NewProc("SDL_CreateTexture")
	procDestroyTexture = sdl3.NewProc("SDL_DestroyTexture")
	procUpdateTexture = sdl3.NewProc("SDL_UpdateTexture")
	procUpdateYUVTexture = sdl3.NewProc("SDL_UpdateYUVTexture")
	procRenderClear = sdl3.NewProc("SDL_RenderClear")
	procRenderTexture = sdl3.NewProc("SDL_RenderTexture")
	procRenderPresent = sdl3.NewProc("SDL_RenderPresent")
	procPollEvent = sdl3.NewProc("SDL_PollEvent")
	procDelay = sdl3.NewProc("SDL_Delay")
	procGetTicks = sdl3.NewProc("SDL_GetTicks")
	procGetWindowSize = sdl3.NewProc("SDL_GetWindowSize")
	procSetWindowTitle = sdl3.NewProc("SDL_SetWindowTitle")
	procSetWindowMouseGrab = sdl3.NewProc("SDL_SetWindowMouseGrab")
	procGetKeyboardState = sdl3.NewProc("SDL_GetKeyboardState")
	procFree = sdl3.NewProc("SDL_free")
	procGetError = sdl3.NewProc("SDL_GetError")
}
