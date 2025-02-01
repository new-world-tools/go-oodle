package oodle

import "C"
import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

var possibleLibPaths = []string{
	libName,
	getTempDllPath(),
}

func Compress(input []byte, compressor int, compressionLevel int) ([]byte, error) {
	_, err := loadLib()
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

	r1 := compress(
		compressor,
		unsafe.Pointer(&input[0]),
		inputSize,
		unsafe.Pointer(&output[0]),
		compressionLevel,
		options,
		dictionaryBase,
		lrm,
		scratchMem,
		scratchSize,
	)

	if r1 == 0 {
		return nil, errors.New("compress failure")
	}

	data := make([]byte, r1)
	copy(data, output)

	return data, nil
}

func Decompress(input []byte, outputSize int64) ([]byte, error) {
	_, err := loadLib()
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

	r1 := decompress(
		unsafe.Pointer(&input[0]),
		inputSize,
		unsafe.Pointer(&output[0]),
		outputSize,
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

	if r1 == 0 {
		return nil, errors.New("decompress failure")
	}

	return output, nil
}

func GetCompressionLevelName(compressionLevel int) (string, error) {
	_, err := loadLib()
	if err != nil {
		return "", err
	}

	r1 := getCompressionLevelName(compressionLevel)
	if r1 == 0 {
		return "", errors.New("error getting compression level name")
	}

	return C.GoString((*C.char)(unsafe.Pointer(r1))), nil
}

func GetCompressorName(compressor int) (string, error) {
	_, err := loadLib()
	if err != nil {
		return "", err
	}

	r1 := getCompressorName(compressor)
	if r1 == 0 {
		return "", errors.New("error getting compressor name")
	}

	return C.GoString((*C.char)(unsafe.Pointer(r1))), nil
}

var compress func(int, unsafe.Pointer, int, unsafe.Pointer, int, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr
var decompress func(unsafe.Pointer, int, unsafe.Pointer, int64, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr
var getCompressionLevelName func(int) uintptr
var getCompressorName func(int) uintptr

func IsLibExists() bool {
	_, err := resolveLibPath()
	return err == nil
}

func resolveLibPath() (string, error) {
	for _, libPath := range possibleLibPaths {
		_, err := os.Stat(libPath)
		if !os.IsNotExist(err) {
			return libPath, nil
		}
	}

	return "", fmt.Errorf("`%s` could not be resolved", libName)
}

func getTempDllPath() string {
	return filepath.Join(os.TempDir(), "go-oodle", libName)
}

func Download() error {
	req, err := http.NewRequest("GET", libUrl, nil)
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
