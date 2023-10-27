package oodle

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"unsafe"
)

const (
	CompressorInvalid = -1
	CompressorNone    = 3

	// new
	CompressorKraken    = 8
	CompressorLeviathan = 13
	CompressorMermaid   = 9
	CompressorSelkie    = 11
	CompressorHydra     = 12

	// deprecated
	CompressorBitKnit = 10
	CompressorLZB16   = 4
	CompressorLZNA    = 7
	CompressorLZH     = 0
	CompressoLZHLW    = 1
	CompressorLZNIB   = 2
	CompressorLZBLW   = 5
	CompressorLZA     = 6

	CompressorCount   = 14
	CompressorForce32 = 0x40000000
)

const (
	CompressionLevelNone      = 0
	CompressionLevelSuperFast = 1
	CompressionLevelVeryFast  = 2
	CompressionLevelFast      = 3
	CompressionLevelNormal    = 4

	CompressionLevelOptimal1 = 5
	CompressionLevelOptimal2 = 6
	CompressionLevelOptimal3 = 7
	CompressionLevelOptimal4 = 8
	CompressionLevelOptimal5 = 9

	CompressionLevelHyperFast1 = -1
	CompressionLevelHyperFast2 = -2
	CompressionLevelHyperFast3 = -3
	CompressionLevelHyperFast4 = -4

	// aliases
	CompressionLevelHyperFast = CompressionLevelHyperFast1
	CompressionLevelOptimal   = CompressionLevelOptimal2
	CompressionLevelMax       = CompressionLevelOptimal5
	CompressionLevelMin       = CompressionLevelHyperFast4

	CompressionLevelForce32 = 0x40000000
	CompressionLevelInvalid = CompressionLevelForce32
)

const (
	FuzzSafeNo  = 0
	FuzzSafeYes = 1
)

const (
	CheckCRCNo      = 0
	CheckCRCYes     = 1
	CheckCRCForce32 = 0x40000000
)

const (
	VerbosityNone    = 0
	VerbosityMinimal = 1
	VerbositySome    = 2
	VerbosityLots    = 3
	VerbosityForce32 = 0x40000000
)

const (
	DecodeThreadPhase1   = 1
	DecodeThreadPhase2   = 2
	DecodeThreadPhaseAll = 3
	DecodeUnthreaded     = DecodeThreadPhaseAll
)

var (
	dllName = "oo2core_9_win64.dll"
	paths   = []string{
		dllName,
		getTempDllPath(),
	}
)

var dllRe = regexp.MustCompile(dllName)

func Compress(input []byte, compressor int, compressionLevel int) ([]byte, error) {
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

	var options uintptr = 0
	var dictionaryBase uintptr = 0
	var lrm uintptr = 0
	var scratchMem uintptr = 0
	var scratchSize uintptr = 0

	r1, _, err := proc.Call(
		uintptr(compressor),
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(inputSize),
		uintptr(unsafe.Pointer(&output[0])),
		uintptr(compressionLevel),
		options,
		dictionaryBase,
		lrm,
		scratchMem,
		scratchSize,
	)

	if err.(syscall.Errno) != 0 {
		return nil, err
	}

	if r1 == 0 {
		return nil, errors.New("compress failure")
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

	var decBufBase uintptr = 0
	var decBufSize uintptr = 0
	var fpCallback uintptr = 0
	var callbackUserData uintptr = 0
	var decoderMemory uintptr = 0
	var decoderMemorySize uintptr = 0

	r1, _, err := proc.Call(
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(inputSize),
		uintptr(unsafe.Pointer(&output[0])),
		uintptr(outputSize),
		FuzzSafeNo,
		CheckCRCNo,
		VerbosityNone,
		decBufBase,
		decBufSize,
		fpCallback,
		callbackUserData,
		decoderMemory,
		decoderMemorySize,
		DecodeThreadPhaseAll,
	)

	if err.(syscall.Errno) != 0 {
		return nil, err
	}

	if r1 == 0 {
		return nil, errors.New("decompress failure")
	}

	return output, nil
}

//func DecompressChunk(input []byte, outputSize int64) ([]byte, error) {
//	dll, err := getDll()
//	if err != nil {
//		return nil, err
//	}
//
//	decoderCreateProc, err := dll.FindProc("OodleLZDecoder_Create")
//	if err != nil {
//		return nil, err
//	}
//
//	//var compressor uintptr = CompressorInvalid
//	var memory uintptr = 0
//	var memorySize uintptr = 0
//
//	r1, _, err := decoderCreateProc.Call(
//		-1,
//		uintptr(outputSize),
//		memory,
//		memorySize,
//	)
//
//	if err.(syscall.Errno) != 0 {
//		return nil, err
//	}
//
//	decodeSomeProc, err := dll.FindProc("OodleLZDecoder_DecodeSome")
//	if err != nil {
//		return nil, err
//	}
//
//	inputSize := len(input)
//	output := make([]byte, 20000)
//	decBuf := make([]byte, 20000)
//
//	var decBufPos uintptr = 0
//	var decBufferSize uintptr = uintptr(inputSize)
//	var decBufAvail = decBufferSize - decBufPos
//
//	r1, _, err = decodeSomeProc.Call(
//		r1,
//		uintptr(unsafe.Pointer(&output[0])),
//		uintptr(unsafe.Pointer(&decBuf[0])),
//		decBufPos,
//		decBufferSize,
//		decBufAvail,
//
//		uintptr(unsafe.Pointer(&input[0])),
//		uintptr(inputSize),
//		FuzzSafeNo,
//		CheckCRCNo,
//		VerbosityNone,
//		DecodeThreadPhaseAll,
//	)
//
//	if err.(syscall.Errno) != 0 {
//		return nil, err
//	}
//
//	if r1 == 0 {
//		return nil, errors.New("decompress failure")
//	}
//
//	return output, nil
//}

//func GetCompressionLevelName(compressionLevel int) (string, error) {
//	dll, err := getDll()
//	if err != nil {
//		return "", err
//	}
//
//	proc, err := dll.FindProc("OodleLZ_CompressionLevel_GetName")
//	if err != nil {
//		return "", err
//	}
//
//	r1, _, err := proc.Call(
//		uintptr(compressionLevel),
//	)
//
//	if err.(syscall.Errno) != 0 {
//		return "", err
//	}
//
//	return C.GoString((*C.char)(unsafe.Pointer(r1))), nil
//}
//
//func GetCompressorName(compressor int) (string, error) {
//	dll, err := getDll()
//	if err != nil {
//		return "", err
//	}
//
//	proc, err := dll.FindProc("OodleLZ_Compressor_GetName")
//	if err != nil {
//		return "", err
//	}
//
//	r1, _, err := proc.Call(
//		uintptr(compressor),
//	)
//
//	if err.(syscall.Errno) != 0 {
//		return "", err
//	}
//
//	return C.GoString((*C.char)(unsafe.Pointer(r1))), nil
//}

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

func getTempDllPath() string {
	return filepath.Join(os.TempDir(), "go-oodle", dllName)
}

const dllUrl = "https://github.com/new-world-tools/go-oodle/releases/download/v0.2.1-file/oo2core_9_win64.dll"

func Download() error {
	req, err := http.NewRequest("GET", dllUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filePath := getTempDllPath()
	err = os.MkdirAll(filepath.Dir(filePath), 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
