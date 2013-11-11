package context

// so, basically, we're going to encode integer tags in base-16 on the stack

import (
	"runtime"
)

const (
	bitWidth = 4
)

func addStackTag(tag uint, context_call func()) {
	if context_call == nil {
		return
	}
	markS(tag, context_call)
}

func markS(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark0(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark1(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark2(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark3(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark4(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark5(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark6(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark7(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark8(tag uint, cb func()) uintptr { return _m(tag, cb) }
func mark9(tag uint, cb func()) uintptr { return _m(tag, cb) }
func markA(tag uint, cb func()) uintptr { return _m(tag, cb) }
func markB(tag uint, cb func()) uintptr { return _m(tag, cb) }
func markC(tag uint, cb func()) uintptr { return _m(tag, cb) }
func markD(tag uint, cb func()) uintptr { return _m(tag, cb) }
func markE(tag uint, cb func()) uintptr { return _m(tag, cb) }
func markF(tag uint, cb func()) uintptr { return _m(tag, cb) }

func _m(tag_remainder uint, cb func()) uintptr {
	if cb == nil {
		pc, _, _, ok := runtime.Caller(1)
		if !ok {
			panic("unable to find caller")
		}
		return runtime.FuncForPC(pc).Entry()
	}
	if tag_remainder == 0 {
		cb()
		return 0
	}
	current_octal_val := tag_remainder & 0xf
	tag_remainder >>= bitWidth
	switch current_octal_val {
	case 0x0:
		return mark0(tag_remainder, cb)
	case 0x1:
		return mark1(tag_remainder, cb)
	case 0x2:
		return mark2(tag_remainder, cb)
	case 0x3:
		return mark3(tag_remainder, cb)
	case 0x4:
		return mark4(tag_remainder, cb)
	case 0x5:
		return mark5(tag_remainder, cb)
	case 0x6:
		return mark6(tag_remainder, cb)
	case 0x7:
		return mark7(tag_remainder, cb)
	case 0x8:
		return mark8(tag_remainder, cb)
	case 0x9:
		return mark9(tag_remainder, cb)
	case 0xa:
		return markA(tag_remainder, cb)
	case 0xb:
		return markB(tag_remainder, cb)
	case 0xc:
		return markC(tag_remainder, cb)
	case 0xd:
		return markD(tag_remainder, cb)
	case 0xe:
		return markE(tag_remainder, cb)
	case 0xf:
		return markF(tag_remainder, cb)
	default:
		panic("programmer failed at base 16")
	}
}

var pc_lookup = map[uintptr]int{
	markS(0, nil): -1,
	mark0(0, nil): 0x0,
	mark1(0, nil): 0x1,
	mark2(0, nil): 0x2,
	mark3(0, nil): 0x3,
	mark4(0, nil): 0x4,
	mark5(0, nil): 0x5,
	mark6(0, nil): 0x6,
	mark7(0, nil): 0x7,
	mark8(0, nil): 0x8,
	mark9(0, nil): 0x9,
	markA(0, nil): 0xa,
	markB(0, nil): 0xb,
	markC(0, nil): 0xc,
	markD(0, nil): 0xd,
	markE(0, nil): 0xe,
	markF(0, nil): 0xf}

func readStackTags(stack []uintptr) (tags []uint) {
	var current_tag uint
	for _, pc := range stack {
		pc = runtime.FuncForPC(pc).Entry()
		val, ok := pc_lookup[pc]
		if !ok {
			continue
		}
		if val < 0 {
			tags = append(tags, current_tag)
			current_tag = 0
			continue
		}
		current_tag <<= bitWidth
		current_tag += uint(val)
	}
	return
}
