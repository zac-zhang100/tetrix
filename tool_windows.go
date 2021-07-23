package main

import (
	"fmt"
	"math"
	"syscall"
	"unsafe"
)

// https://golang.org/pkg/syscall/?GOOS=windows#LazyDLL
// 加载一个dll动态链接库，用于对windows进行系统调用（Kernel32.dll file handles the memory usage in "Microsoft Windows"）
// NewLazySystemDLL is like NewLazyDLL, but will only search Windows System directory for the DLL if name is a base name (like "advapi32.dll").
var (
	kernel32                            = syscall.NewLazyDLL("kernel32.dll")
	moduser32                           = syscall.NewLazyDLL("user32.dll")
	info                                Console_Screen_Buffer_Info
	proc_create_event                   = kernel32.NewProc("CreateEventW")
	proc_get_console_mode               = kernel32.NewProc("GetConsoleMode")
	proc_set_console_mode               = kernel32.NewProc("SetConsoleMode")
	orig_mode                           dword
	get_system_metrics                  = moduser32.NewProc("GetSystemMetrics")
	proc_get_current_console_font       = kernel32.NewProc("GetCurrentConsoleFont")
	tmp_finfo                           console_font_info
	proc_set_console_screen_buffer_size = kernel32.NewProc("SetConsoleScreenBufferSize")
	proc_set_console_window_info        = kernel32.NewProc("SetConsoleWindowInfo")
	proc_get_console_cursor_info        = kernel32.NewProc("GetConsoleCursorInfo")
	orig_cursor_info                    console_cursor_info
	proc_set_console_cursor_info        = kernel32.NewProc("SetConsoleCursorInfo")
	in                                  syscall.Handle
	out                                 syscall.Handle
)

const (
	console_screen_buffer_info = "GetConsoleScreenBufferInfo"
	enable_window_input        = 0x8
	SM_CXMIN                   = 28
	SM_CYMIN                   = 29
)

// http://fusesource.github.io/jansi/documentation/native-api/index.html?org/fusesource/jansi/internal/Kernel32.CONSOLE_SCREEN_BUFFER_INFO.html
// type Console_Screen_Buffer_Info struct {
// 	attributes        uint16
// 	cursorPosition    coord
// 	maximumWindowSize coord
// 	size              coord
// 	sizeOf            int
// 	window            small_rect
// }

type console_font_info struct {
	font      uint32
	font_size coord
}

type Console_Screen_Buffer_Info struct {
	size                coord
	cursor_position     coord
	attributes          uint16
	window              small_rect
	maximum_window_size coord
}

type coord struct {
	x int16
	y int16
}

type dword uint32
type short int16

type small_rect struct {
	bottom int16
	left   int16
	right  int16
	sizeOf int
	top    int16
}

func Initial() error {
	var (
		err error
		// interrupt syscall.Handle
	)

	_, err = create_event()
	if err != nil {
		return err
	}

	in, err = syscall.Open("CONIN$", syscall.O_RDWR, 0)
	if err != nil {
		return err
	}
	out, err = syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return err
	}

	err = get_console_mode(in, &orig_mode)
	if err != nil {
		return err
	}

	err = set_console_mode(in, enable_window_input)
	if err != nil {
		return err
	}

	// orig_size, _ := get_term_size(out)
	win_size := get_win_size(out)

	err = set_console_screen_buffer_size(out, win_size)
	if err != nil {
		return err
	}

	err = fix_win_size(out, win_size)
	if err != nil {
		return err
	}

	err = get_console_cursor_info(out, &orig_cursor_info)
	if err != nil {
		return err
	}

	show_cursor(false)
	term_size, _ := get_term_size(out)
	fmt.Println("win size is:", term_size.x)
	// back_buffer.init(int(term_size.x), int(term_size.y))
	// front_buffer.init(int(term_size.x), int(term_size.y))
	// back_buffer.clear()
	// front_buffer.clear()
	// clear()

	// diffbuf = make([]diff_msg, 0, 32)

	// go input_event_producer()
	// IsInit = true
	return nil
}

func create_event() (out syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall6(proc_create_event.Addr(),
		4, 0, 0, 0, 0, 0, 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return syscall.Handle(r0), err
}

func Init() error {
	// in, err := syscall.Open("CONIN$", syscall.O_RDWR, 0)
	// if err != nil {
	// 	return err
	// }

	out, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return err
	}
	term_size, _ := get_term_size(out)
	fmt.Println("terminal size is:", int(term_size.x))
	fmt.Println("window size is:", int(term_size.y))
	return nil
}

func get_term_size(out syscall.Handle) (coord, small_rect) {
	err := get_console_screen_buffer_info(out, &info)
	if err != nil {
		panic(err)
	}
	return info.size, info.window
}

func get_console_mode(h syscall.Handle, mode *dword) (err error) {
	r0, _, e1 := syscall.Syscall(proc_get_console_mode.Addr(),
		2, uintptr(h), uintptr(unsafe.Pointer(mode)), 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func get_current_console_font(h syscall.Handle, info *console_font_info) (err error) {
	r0, _, e1 := syscall.Syscall(proc_get_current_console_font.Addr(),
		3, uintptr(h), 0, uintptr(unsafe.Pointer(info)))
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func get_win_min_size(out syscall.Handle) coord {
	x, _, err := get_system_metrics.Call(SM_CXMIN)
	y, _, err := get_system_metrics.Call(SM_CYMIN)

	if x == 0 || y == 0 {
		if err != nil {
			panic(err)
		}
	}

	err1 := get_current_console_font(out, &tmp_finfo)
	if err1 != nil {
		panic(err1)
	}

	return coord{
		x: int16(math.Ceil(float64(x) / float64(tmp_finfo.font_size.x))),
		y: int16(math.Ceil(float64(y) / float64(tmp_finfo.font_size.y))),
	}
}

func set_console_mode(h syscall.Handle, mode dword) (err error) {
	r0, _, e1 := syscall.Syscall(proc_set_console_mode.Addr(),
		2, uintptr(h), uintptr(mode), 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func get_win_size(out syscall.Handle) coord {
	err := get_console_screen_buffer_info(out, &info)
	if err != nil {
		panic(err)
	}

	min_size := get_win_min_size(out)

	size := coord{
		x: info.window.right - info.window.left + 1,
		y: info.window.bottom - info.window.top + 1,
	}

	if size.x < min_size.x {
		size.x = min_size.x
	}

	if size.y < min_size.y {
		size.y = min_size.y
	}

	return size
}

// func Syscall(trap, nargs, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
func get_console_screen_buffer_info(h syscall.Handle, info *Console_Screen_Buffer_Info) (err error) {
	r, _, e := syscall.Syscall(kernel32.NewProc(console_screen_buffer_info).Addr(), 2, uintptr(h), uintptr(unsafe.Pointer(&info)), 0)
	if int(r) == 0 {
		if e != 0 {
			err = error(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func (this coord) uintptr() uintptr {
	return uintptr(*(*int32)(unsafe.Pointer(&this)))
}

func set_console_screen_buffer_size(h syscall.Handle, size coord) (err error) {
	r0, _, e1 := syscall.Syscall(proc_set_console_screen_buffer_size.Addr(),
		2, uintptr(h), size.uintptr(), 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func fix_win_size(out syscall.Handle, size coord) (err error) {
	window := small_rect{}
	window.top = 0
	window.bottom = size.y - 1
	window.left = 0
	window.right = size.x - 1
	return set_console_window_info(out, &window)
}

func (this *small_rect) uintptr() uintptr {
	return uintptr(unsafe.Pointer(this))
}

func set_console_window_info(h syscall.Handle, window *small_rect) (err error) {
	var absolute uint32
	absolute = 1
	r0, _, e1 := syscall.Syscall(proc_set_console_window_info.Addr(),
		3, uintptr(h), uintptr(absolute), window.uintptr())
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

type console_cursor_info struct {
	size    dword
	visible int32
}

func get_console_cursor_info(h syscall.Handle, info *console_cursor_info) (err error) {
	r0, _, e1 := syscall.Syscall(proc_get_console_cursor_info.Addr(),
		2, uintptr(h), uintptr(unsafe.Pointer(info)), 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func show_cursor(visible bool) {
	var v int32
	if visible {
		v = 1
	}

	var info console_cursor_info
	info.size = 100
	info.visible = v
	err := set_console_cursor_info(out, &info)
	if err != nil {
		panic(err)
	}
}

func set_console_cursor_info(h syscall.Handle, info *console_cursor_info) (err error) {
	r0, _, e1 := syscall.Syscall(proc_set_console_cursor_info.Addr(),
		2, uintptr(h), uintptr(unsafe.Pointer(info)), 0)
	if int(r0) == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// func clear() {
// 	var err error
// 	attr, char := cell_to_char_info(Cell{
// 		' ',
// 		foreground,
// 		background,
// 	})

// 	area := int(term_size.x) * int(term_size.y)
// 	err = fill_console_output_attribute(out, attr, area)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = fill_console_output_character(out, char[0], area)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if !is_cursor_hidden(cursor_x, cursor_y) {
// 		move_cursor(cursor_x, cursor_y)
// 	}
// }
