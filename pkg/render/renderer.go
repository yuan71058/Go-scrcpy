package render

import (
	"fmt"
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
	procWaitEvent     *syscall.LazyProc
	procDelay         *syscall.LazyProc
	procGetTicks      *syscall.LazyProc
	procGetWindowSize *syscall.LazyProc
	procSetWindowTitle *syscall.LazyProc
	procSetWindowMouseGrab *syscall.LazyProc
	procGetKeyboardState *syscall.LazyProc
	procFree          *syscall.LazyProc
	procGetError      *syscall.LazyProc
	procPushEvent     *syscall.LazyProc
)

const (
	SDL_INIT_VIDEO = 0x00000020
	SDL_TEXTUREACCESS_STREAMING = 1
	SDL_PIXELFORMAT_YV12 = 0x32315659 // FOURCC('Y','V','1','2')
	SDL_PIXELFORMAT_IYUV = 0x56555949 // FOURCC('I','Y','U','V')
	SDL_EVENT_USER = 0x8000

	// SDL 事件类型
	SDL_EVENT_QUIT = 0x100
	SDL_EVENT_KEY_DOWN = 0x300
	SDL_EVENT_KEY_UP = 0x301
	SDL_EVENT_MOUSE_MOTION = 0x400
	SDL_EVENT_MOUSE_BUTTON_DOWN = 0x401
	SDL_EVENT_MOUSE_BUTTON_UP = 0x402
	SDL_EVENT_MOUSE_WHEEL = 0x403
	SDL_EVENT_FINGER_DOWN = 0x700
	SDL_EVENT_FINGER_UP = 0x701
	SDL_EVENT_FINGER_MOTION = 0x702
	SDL_EVENT_WINDOW_RESIZED = 0x210
	SDL_EVENT_WINDOW_CLOSE_REQUESTED = 0x21C

	// 自定义事件
	SC_EVENT_NEW_FRAME = SDL_EVENT_USER
)

type SDLRenderer struct {
	window   uintptr
	renderer uintptr
	texture  uintptr
	width    int
	height   int
}

// InitSDL 初始化 SDL 视频子系统（必须在创建窗口前调用）
func InitSDL() {
	procInit.Call(SDL_INIT_VIDEO)
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
		errMsg := getLastError()
		procDestroyWindow.Call(win)
		return nil, fmt.Errorf("创建渲染器失败: %s", errMsg)
	}

	tex, _, _ := procCreateTexture.Call(
		ren,
		uintptr(SDL_PIXELFORMAT_IYUV),
		uintptr(SDL_TEXTUREACCESS_STREAMING),
		uintptr(w), uintptr(h),
	)
	if tex == 0 {
		errMsg := getLastError()
		procDestroyRenderer.Call(ren)
		procDestroyWindow.Call(win)
		return nil, fmt.Errorf("创建纹理失败: %s", errMsg)
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

	// YUV420P: Y plane + U plane + V plane
	ySize := r.width * r.height
	uvW := (r.width + 1) / 2
	uvH := (r.height + 1) / 2
	uvSize := uvW * uvH

	// IYUV 格式直接对应 Y/U/V 平面顺序
	procUpdateYUVTexture.Call(
		r.texture,
		0,
		uintptr(unsafe.Pointer(&yuv[0])), uintptr(r.width),           // Y plane
		uintptr(unsafe.Pointer(&yuv[ySize])), uintptr(uvW),           // U plane
		uintptr(unsafe.Pointer(&yuv[ySize+uvSize])), uintptr(uvW),    // V plane
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

// PushEvent 推送自定义事件到 SDL 事件队列
func PushEvent(eventType uint32, data1 uintptr) bool {
	var event [128]byte
	// SDL_UserEvent (SDL3):
	// offset 0: type (4 bytes)
	// offset 4: reserved (4 bytes)
	// offset 8: timestamp (8 bytes)
	// offset 16: windowID (4 bytes)
	// offset 20: code (4 bytes)
	// offset 24: data1 (8 bytes)
	// offset 32: data2 (8 bytes)
	*(*uint32)(unsafe.Pointer(&event[0])) = eventType
	*(*uintptr)(unsafe.Pointer(&event[24])) = data1
	ret, _, _ := procPushEvent.Call(uintptr(unsafe.Pointer(&event[0])))
	return ret != 0
}

// PollEvent 轮询 SDL 事件，返回事件类型和数据
func PollEvent() (uint32, []byte) {
	event := make([]byte, 128)
	ret, _, _ := procPollEvent.Call(uintptr(unsafe.Pointer(&event[0])))
	if ret == 0 {
		return 0, nil
	}
	return *(*uint32)(unsafe.Pointer(&event[0])), event
}

// WaitEvent 阻塞等待 SDL 事件，返回事件类型和数据
// 参考原版 scrcpy：主线程使用 SDL_WaitEvent 阻塞等待，避免忙等导致 Windows 消息泵饥饿
func WaitEvent() (uint32, []byte) {
	event := make([]byte, 128)
	ret, _, _ := procWaitEvent.Call(uintptr(unsafe.Pointer(&event[0])))
	if ret == 0 {
		return 0, nil
	}
	return *(*uint32)(unsafe.Pointer(&event[0])), event
}

// Delay 等待指定毫秒数
func Delay(ms uint32) {
	procDelay.Call(uintptr(ms))
}

// GetEventMouseButton 从 SDL 事件中获取鼠标按钮（1=左, 2=中, 3=右）
func GetEventMouseButton(event []byte) int {
	if len(event) < 25 {
		return 0
	}
	return int(event[24])
}

// GetEventMouseCoords 从 SDL 事件中获取鼠标坐标 (SDL3: x=offset 28, y=offset 32)
func GetEventMouseCoords(event []byte) (x, y float32) {
	if len(event) < 36 {
		return 0, 0
	}
	x = *(*float32)(unsafe.Pointer(&event[28]))
	y = *(*float32)(unsafe.Pointer(&event[32]))
	return
}

// GetEventMouseScroll 从 SDL 事件中获取鼠标滚轮 (SDL3: x=offset 28, y=offset 32)
func GetEventMouseScroll(event []byte) (hscroll, vscroll float32) {
	if len(event) < 36 {
		return 0, 0
	}
	hscroll = *(*float32)(unsafe.Pointer(&event[28]))
	vscroll = *(*float32)(unsafe.Pointer(&event[32]))
	return
}

// GetEventKeycode 从 SDL 事件中获取按键码
func GetEventKeycode(event []byte) int32 {
	if len(event) < 32 {
		return 0
	}
	return *(*int32)(unsafe.Pointer(&event[28]))
}

// GetEventKeyMod 从 SDL 事件中获取修饰键
func GetEventKeyMod(event []byte) int32 {
	if len(event) < 34 {
		return 0
	}
	return int32(*(*uint16)(unsafe.Pointer(&event[32])))
}

func getLastError() string {
	ret, _, _ := procGetError.Call()
	if ret == 0 {
		return ""
	}
	// SDL_GetError returns const char*, read the string
	var buf [256]byte
	for i := 0; i < 255; i++ {
		b := *(*byte)(unsafe.Pointer(ret + uintptr(i)))
		buf[i] = b
		if b == 0 {
			break
		}
	}
	return string(buf[:])
}

func init() {
	dllDir := findSDL3Dir()
	if dllDir == "" {
		dllDir = "."
	}
	sdl3 = loadDLL(filepath.Join(dllDir, "SDL3.dll"))

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
		procWaitEvent = sdl3.NewProc("SDL_WaitEvent")
	procDelay = sdl3.NewProc("SDL_Delay")
	procGetTicks = sdl3.NewProc("SDL_GetTicks")
	procGetWindowSize = sdl3.NewProc("SDL_GetWindowSize")
	procSetWindowTitle = sdl3.NewProc("SDL_SetWindowTitle")
	procSetWindowMouseGrab = sdl3.NewProc("SDL_SetWindowMouseGrab")
	procGetKeyboardState = sdl3.NewProc("SDL_GetKeyboardState")
	procFree = sdl3.NewProc("SDL_free")
	procGetError = sdl3.NewProc("SDL_GetError")
	procPushEvent = sdl3.NewProc("SDL_PushEvent")
}