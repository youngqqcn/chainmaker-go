package wasm

import (
	"encoding/binary"
)

var defaultMax = uint32(65535)  // 4G

func memoryBase(vm *VirtualMachine) uint64 {
	vm.ActiveContext.PC++
	_ = vm.FetchUint32() // ignore align
	vm.ActiveContext.PC++
	return uint64(vm.FetchUint32()) + vm.OperandStack.Pop()
}

func i32Load(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(uint64(binary.LittleEndian.Uint32(vm.Memory[base:])))
}

func i64Load(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(binary.LittleEndian.Uint64(vm.Memory[base:]))
}

func f32Load(vm *VirtualMachine) {
	i32Load(vm)
}

func f64Load(vm *VirtualMachine) {
	i64Load(vm)
}

func i32Load8s(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(uint64(vm.Memory[base]))
}

func i32Load8u(vm *VirtualMachine) {
	i32Load8s(vm)
}

func i32Load16s(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(uint64(binary.LittleEndian.Uint16(vm.Memory[base:])))
}

func i32Load16u(vm *VirtualMachine) {
	i32Load16s(vm)
}

func i64Load8s(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(uint64(vm.Memory[base]))
}

func i64Load8u(vm *VirtualMachine) {
	i64Load8s(vm)
}

func i64Load16s(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(uint64(binary.LittleEndian.Uint16(vm.Memory[base:])))
}

func i64Load16u(vm *VirtualMachine) {
	i64Load16s(vm)
}

func i64Load32s(vm *VirtualMachine) {
	base := memoryBase(vm)
	safeStore(vm, base, false)
	vm.OperandStack.Push(uint64(binary.LittleEndian.Uint32(vm.Memory[base:])))
}

func i64Load32u(vm *VirtualMachine) {
	i64Load32s(vm)
}

func i32Store(vm *VirtualMachine) {
	val := vm.OperandStack.Pop()
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint32(vm.Memory[base:], uint32(val))
}

func i64Store(vm *VirtualMachine) {
	val := vm.OperandStack.Pop()
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint64(vm.Memory[base:], val)
}

func f32Store(vm *VirtualMachine) {
	val := vm.OperandStack.Pop()
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint32(vm.Memory[base:], uint32(val))
}

func f64Store(vm *VirtualMachine) {
	v := vm.OperandStack.Pop()
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint64(vm.Memory[base:], v)
}

func i32Store8(vm *VirtualMachine) {
	v := byte(vm.OperandStack.Pop())
	base := memoryBase(vm)
	safeStore(vm, base, true)
	vm.Memory[base] = v
}

func i32Store16(vm *VirtualMachine) {
	v := uint16(vm.OperandStack.Pop())
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint16(vm.Memory[base:], v)
}

func i64Store8(vm *VirtualMachine) {
	v := byte(vm.OperandStack.Pop())
	base := memoryBase(vm)
	safeStore(vm, base, true)
	vm.Memory[base] = v
}

func i64Store16(vm *VirtualMachine) {
	v := uint16(vm.OperandStack.Pop())
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint16(vm.Memory[base:], v)
}

func i64Store32(vm *VirtualMachine) {
	v := uint32(vm.OperandStack.Pop())
	base := memoryBase(vm)
	safeStore(vm, base, true)
	binary.LittleEndian.PutUint32(vm.Memory[base:], v)
}

func memorySize(vm *VirtualMachine) {
	vm.ActiveContext.PC++
	vm.OperandStack.Push(uint64(int32(len(vm.Memory) / vmPageSize)))
}

func memoryGrow(vm *VirtualMachine) {
	vm.ActiveContext.PC++
	n := uint32(vm.OperandStack.Pop())

	if vm.InnerModule.SecMemory[0].Max != nil &&
		uint64(n+uint32(len(vm.Memory)/vmPageSize)) > uint64(*(vm.InnerModule.SecMemory[0].Max)) {
		v := int32(-1)
		vm.OperandStack.Push(uint64(v))
		return
	}

	vm.OperandStack.Push(uint64(len(vm.Memory)) / vmPageSize)
	vm.Memory = append(vm.Memory, make([]byte, n*vmPageSize)...)
	vm.Gas += tableGas[OptCodeMemoryGrow] * uint64(n)
}

func safeStore(vm *VirtualMachine, base uint64, isStore bool) {
	if base >= uint64(len(vm.Memory)) {
		if !isStore {
			panic("Illegal memory operation ")
			return
		}

		if vm.InnerModule.SecMemory[0].Max == nil {
			vm.InnerModule.SecMemory[0].Max = &defaultMax
		}

		diff := base - uint64(len(vm.Memory)) + 1
		preMemory := len(vm.Memory) + int(diff)
		if uint64(uint32(preMemory/vmPageSize)) > uint64(*(vm.InnerModule.SecMemory[0].Max)) {
			panic("out of vm max memory limit ")
			return
		}

		if isStore {
			vm.Memory = append(vm.Memory, make([]byte, diff)...)
		}
	}
}
