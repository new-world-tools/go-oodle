package oodle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"unsafe"
)

const (
	CompressionLevelNone      = 0
	CompressionLevelSuperFast = 1
	CompressionLevelVeryFast  = 2
	CompressionLevelFast      = 3
	CompressionLevelNormal    = 4
	CompressionLevelOptimal1  = 5
	CompressionLevelOptimal2  = 6
	CompressionLevelOptimal3  = 7
	CompressionLevelCount     = 8
)

const (
	AlgoLZH         = 0
	AlgoLZHLW       = 1
	AlgoLZNIB       = 2
	AlgoNone        = 3
	AlgoLZB16       = 4
	AlgoLZBLW       = 5
	AlgoLZA         = 6
	AlgoLZNA        = 7
	AlgoKraken      = 8
	AlgoLZQ1        = 8
	AlgoMermaid     = 9
	AlgoLZNIB2      = 9
	AlgoBitKnit     = 10
	AlgoSelkie      = 11
	AlgoHydra       = 12
	AlgoAkkorokamui = 12
	AlgoLeviathan   = 13
	AlgoCount       = 14
)

var (
	dllName = "oo2core_9_win64.dll"
	paths   = []string{
		dllName,
		filepath.Join(os.TempDir(), "go-oodle", dllName),
	}
)

var dllRe = regexp.MustCompile(dllName)

func Compress(input []byte, algo int, compressionLevel int) ([]byte, error) {
	dll, err := getDll()
	if err != nil {
		return nil, err
	}

	proc, err := dll.FindProc("OodleLZ_Compress")
	if err != nil {
		return nil, err
	}

	inputSize := len(input)
	outputSize := inputSize * 2
	output := make([]byte, outputSize)

	r1, _, err := proc.Call(
		uintptr(algo),
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(inputSize),
		uintptr(unsafe.Pointer(&output[0])),
		uintptr(compressionLevel),
		0,
		0,
		0,
	)

	if err.(syscall.Errno) != 0 {
		return nil, err
	}

	data := make([]byte, r1)
	copy(data, output)

	return data, nil
}

func Decompress(input []byte, outputSize int64) ([]byte, error) {
	dll, err := getDll()
	if err != nil {
		return nil, err
	}

	proc, err := dll.FindProc("OodleLZ_Decompress")
	if err != nil {
		return nil, err
	}

	inputSize := len(input)
	output := make([]byte, outputSize)

	_, _, err = proc.Call(
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(inputSize),
		uintptr(unsafe.Pointer(&output[0])),
		uintptr(outputSize),
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		3,
	)

	if err.(syscall.Errno) != 0 {
		return nil, err
	}

	return output, nil
}

var dllOnce struct {
	sync.Once
	dll *syscall.DLL
	err error
}

func getDll() (*syscall.DLL, error) {
	dllOnce.Do(func() {
		dllPath, err := resolveDllPath()
		if err != nil {
			dllOnce.err = err
			return
		}
		dll, err := syscall.LoadDLL(dllPath)

		dllOnce.dll = dll
		dllOnce.err = err
	})

	return dllOnce.dll, dllOnce.err
}

func IsDllExist() bool {
	_, err := resolveDllPath()
	return err == nil
}

func resolveDllPath() (string, error) {
	for _, dllPath := range paths {
		_, err := os.Stat(dllPath)
		if !os.IsNotExist(err) {
			return dllPath, nil
		}
	}

	return "", fmt.Errorf("`%s` is not resolve", dllName)
}
