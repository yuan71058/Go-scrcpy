package input

// Android keycode 常量
// 完整列表参考: https://developer.android.com/reference/android/view/KeyEvent
const (
	KeycodeUnknown       int32 = 0
	KeycodeHome          int32 = 3   // HOME 键
	KeycodeBack          int32 = 4   // 返回键
	KeycodeCall          int32 = 5   // 拨号键
	KeycodeEndCall       int32 = 6   // 挂断键
	Keycode0             int32 = 7   // 数字 0
	Keycode1             int32 = 8   // 数字 1
	Keycode2             int32 = 9   // 数字 2
	Keycode3             int32 = 10  // 数字 3
	Keycode4             int32 = 11  // 数字 4
	Keycode5             int32 = 12  // 数字 5
	Keycode6             int32 = 13  // 数字 6
	Keycode7             int32 = 14  // 数字 7
	Keycode8             int32 = 15  // 数字 8
	Keycode9             int32 = 16  // 数字 9
	KeycodeStar          int32 = 17  // 星号键
	KeycodePound         int32 = 18  // 井号键
	KeycodeDPadUp        int32 = 19  // 方向键上
	KeycodeDPadDown      int32 = 20  // 方向键下
	KeycodeDPadLeft      int32 = 21  // 方向键左
	KeycodeDPadRight     int32 = 22  // 方向键右
	KeycodeDPadCenter    int32 = 23  // 方向键中
	KeycodeVolumeUp      int32 = 24  // 音量加
	KeycodeVolumeDown    int32 = 25  // 音量减
	KeycodePower         int32 = 26  // 电源键
	KeycodeCamera        int32 = 27  // 相机键
	KeycodeClear         int32 = 28  // 清除键
	KeycodeA             int32 = 29  // 字母 A
	KeycodeB             int32 = 30  // 字母 B
	KeycodeC             int32 = 31  // 字母 C
	KeycodeD             int32 = 32  // 字母 D
	KeycodeE             int32 = 33  // 字母 E
	KeycodeF             int32 = 34  // 字母 F
	KeycodeG             int32 = 35  // 字母 G
	KeycodeH             int32 = 36  // 字母 H
	KeycodeI             int32 = 37  // 字母 I
	KeycodeJ             int32 = 38  // 字母 J
	KeycodeK             int32 = 39  // 字母 K
	KeycodeL             int32 = 40  // 字母 L
	KeycodeM             int32 = 41  // 字母 M
	KeycodeN             int32 = 42  // 字母 N
	KeycodeO             int32 = 43  // 字母 O
	KeycodeP             int32 = 44  // 字母 P
	KeycodeQ             int32 = 45  // 字母 Q
	KeycodeR             int32 = 46  // 字母 R
	KeycodeS             int32 = 47  // 字母 S
	KeycodeT             int32 = 48  // 字母 T
	KeycodeU             int32 = 49  // 字母 U
	KeycodeV             int32 = 50  // 字母 V
	KeycodeW             int32 = 51  // 字母 W
	KeycodeX             int32 = 52  // 字母 X
	KeycodeY             int32 = 53  // 字母 Y
	KeycodeZ             int32 = 54  // 字母 Z
	KeycodeComma         int32 = 55  // 逗号
	KeycodePeriod        int32 = 56  // 句号
	KeycodeAltLeft       int32 = 57  // 左 Alt
	KeycodeAltRight      int32 = 58  // 右 Alt
	KeycodeShiftLeft     int32 = 59  // 左 Shift
	KeycodeShiftRight    int32 = 60  // 右 Shift
	KeycodeTab           int32 = 61  // Tab 键
	KeycodeSpace         int32 = 62  // 空格键
	KeycodeSym           int32 = 63  // 符号键
	KeycodeExplorer      int32 = 64  // 浏览器键
	KeycodeEnvelope      int32 = 65  // 邮件键
	KeycodeEnter         int32 = 66  // 回车键
	KeycodeDel           int32 = 67  // 删除键
	KeycodeGrave         int32 = 68  // 反引号
	KeycodeLeftParen     int32 = 69  // 左括号
	KeycodeRightParen    int32 = 70  // 右括号
	KeycodeSemicolon     int32 = 71  // 分号
	KeycodeApostrophe    int32 = 72  // 撇号
	KeycodeSlash         int32 = 73  // 斜杠
	KeycodeAt            int32 = 74  // @ 符号
	KeycodeNum           int32 = 75  // 数字键
	KeycodeHeadsethook   int32 = 76  // 耳机钩
	KeycodeFocus         int32 = 77  // 对焦键
	KeycodePlus          int32 = 78  // 加号
	KeycodeMenu          int32 = 82  // 菜单键
	KeycodeNotification  int32 = 83  // 通知键
	KeycodeSearch        int32 = 84  // 搜索键
	KeycodeMediaPlayPause int32 = 85  // 播放/暂停
	KeycodeMediaStop     int32 = 86  // 停止
	KeycodeMediaNext     int32 = 87  // 下一曲
	KeycodeMediaPrevious int32 = 88  // 上一曲
	KeycodeMediaRewind   int32 = 89  // 快退
	KeycodeMediaFastForward int32 = 90 // 快进
	KeycodeMute          int32 = 91  // 静音
	KeycodePageUp        int32 = 92  // Page Up
	KeycodePageDown      int32 = 93  // Page Down
	KeycodeEscape        int32 = 111 // ESC 键
	KeycodeForwardDel    int32 = 112 // 前向删除
	KeycodeCtrlLeft      int32 = 113 // 左 Ctrl
	KeycodeCtrlRight     int32 = 114 // 右 Ctrl
	KeycodeCapsLock      int32 = 115 // CapsLock
	KeycodeScrollLock    int32 = 116 // ScrollLock
	KeycodeMetaLeft      int32 = 117 // 左 Meta/Win
	KeycodeMetaRight     int32 = 118 // 右 Meta/Win
	KeycodeBreak         int32 = 121 // Break/Pause
	KeycodeInsert        int32 = 122 // Insert
	KeycodeMoveHome      int32 = 122 // Home
	KeycodeMoveEnd       int32 = 123 // End
	KeycodeF1            int32 = 131 // F1
	KeycodeF2            int32 = 132 // F2
	KeycodeF3            int32 = 133 // F3
	KeycodeF4            int32 = 134 // F4
	KeycodeF5            int32 = 135 // F5
	KeycodeF6            int32 = 136 // F6
	KeycodeF7            int32 = 137 // F7
	KeycodeF8            int32 = 138 // F8
	KeycodeF9            int32 = 139 // F9
	KeycodeF10           int32 = 140 // F10
	KeycodeF11           int32 = 141 // F11
	KeycodeF12           int32 = 142 // F12
	KeycodeNumLock       int32 = 143 // NumLock
	KeycodeNumpad0       int32 = 144 // 小键盘 0
	KeycodeNumpad1       int32 = 145 // 小键盘 1
	KeycodeNumpad2       int32 = 146 // 小键盘 2
	KeycodeNumpad3       int32 = 147 // 小键盘 3
	KeycodeNumpad4       int32 = 148 // 小键盘 4
	KeycodeNumpad5       int32 = 149 // 小键盘 5
	KeycodeNumpad6       int32 = 150 // 小键盘 6
	KeycodeNumpad7       int32 = 151 // 小键盘 7
	KeycodeNumpad8       int32 = 152 // 小键盘 8
	KeycodeNumpad9       int32 = 153 // 小键盘 9
	KeycodeNumpadDivide  int32 = 154 // 小键盘除号
	KeycodeNumpadMultiply int32 = 155 // 小键盘乘号
	KeycodeNumpadSubtract int32 = 156 // 小键盘减号
	KeycodeNumpadAdd     int32 = 157 // 小键盘加号
	KeycodeNumpadDot     int32 = 158 // 小键盘点号
	KeycodeNumpadEnter   int32 = 160 // 小键盘回车
	KeycodeNumpadEqual   int32 = 161 // 小键盘等号
	KeycodeAppSwitch     int32 = 187 // 应用切换
	KeycodePictureInPicture int32 = 171 // 画中画
	KeycodeAllApps       int32 = 284 // 所有应用
	KeycodeApp2          int32 = 302 // 应用 2
	KeycodeAssistant     int32 = 319 // 助手
)

// KeyCodeToName keycode 到名称的映射
var KeyCodeToName = map[int32]string{
	KeycodeHome:          "HOME",
	KeycodeBack:          "BACK",
	KeycodeDPadUp:        "DPAD_UP",
	KeycodeDPadDown:      "DPAD_DOWN",
	KeycodeDPadLeft:      "DPAD_LEFT",
	KeycodeDPadRight:     "DPAD_RIGHT",
	KeycodeDPadCenter:    "DPAD_CENTER",
	KeycodeVolumeUp:      "VOLUME_UP",
	KeycodeVolumeDown:    "VOLUME_DOWN",
	KeycodePower:         "POWER",
	KeycodeEnter:         "ENTER",
	KeycodeDel:           "DELETE",
	KeycodeSpace:         "SPACE",
	KeycodeTab:           "TAB",
	KeycodeEscape:        "ESCAPE",
	KeycodeMenu:          "MENU",
	KeycodeSearch:        "SEARCH",
	KeycodeCtrlLeft:      "CTRL_LEFT",
	KeycodeCtrlRight:     "CTRL_RIGHT",
	KeycodeShiftLeft:     "SHIFT_LEFT",
	KeycodeShiftRight:    "SHIFT_RIGHT",
	KeycodeAltLeft:       "ALT_LEFT",
	KeycodeAltRight:      "ALT_RIGHT",
	KeycodeMetaLeft:      "META_LEFT",
	KeycodeMetaRight:     "META_RIGHT",
	KeycodeCapsLock:      "CAPS_LOCK",
	KeycodeNumLock:       "NUM_LOCK",
	KeycodeScrollLock:    "SCROLL_LOCK",
	KeycodeF1:            "F1",
	KeycodeF2:            "F2",
	KeycodeF3:            "F3",
	KeycodeF4:            "F4",
	KeycodeF5:            "F5",
	KeycodeF6:            "F6",
	KeycodeF7:            "F7",
	KeycodeF8:            "F8",
	KeycodeF9:            "F9",
	KeycodeF10:           "F10",
	KeycodeF11:           "F11",
	KeycodeF12:           "F12",
	KeycodeA:             "A",
	KeycodeB:             "B",
	KeycodeC:             "C",
	KeycodeD:             "D",
	KeycodeE:             "E",
	KeycodeF:             "F",
	KeycodeG:             "G",
	KeycodeH:             "H",
	KeycodeI:             "I",
	KeycodeJ:             "J",
	KeycodeK:             "K",
	KeycodeL:             "L",
	KeycodeM:             "M",
	KeycodeN:             "N",
	KeycodeO:             "O",
	KeycodeP:             "P",
	KeycodeQ:             "Q",
	KeycodeR:             "R",
	KeycodeS:             "S",
	KeycodeT:             "T",
	KeycodeU:             "U",
	KeycodeV:             "V",
	KeycodeW:             "W",
	KeycodeX:             "X",
	KeycodeY:             "Y",
	KeycodeZ:             "Z",
	Keycode0:             "0",
	Keycode1:             "1",
	Keycode2:             "2",
	Keycode3:             "3",
	Keycode4:             "4",
	Keycode5:             "5",
	Keycode6:             "6",
	Keycode7:             "7",
	Keycode8:             "8",
	Keycode9:             "9",
}

// GetNameByKeycode 获取 keycode 的名称
func GetNameByKeycode(keycode int32) string {
	if name, ok := KeyCodeToName[keycode]; ok {
		return name
	}
	return "UNKNOWN"
}
