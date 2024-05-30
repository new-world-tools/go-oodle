package oodle

import (
	"github.com/ebitengine/purego"
	"golang.org/x/sys/windows"
	"sync"
)

const (
	libUrl  = "https://github.com/new-world-tools/go-oodle/releases/download/v0.2.3-files/oo2core_9_win64.dll"
	libName = "oo2core_9_win64.dll"
)

var libOnce struct {
	sync.Once
	handle uintptr
	err    error
}

func loadLib() (uintptr, error) {
	libOnce.Do(func() {
		libPath, err := resolveLibPath()
		if err != nil {
			libOnce.err = err
			return
		}

		handle, err := windows.LoadLibrary(libPath)
		if err != nil {
			libOnce.err = err
			return
		}

		libOnce.handle = uintptr(handle)

		purego.RegisterLibFunc(&compress, libOnce.handle, "OodleLZ_Compress")
		purego.RegisterLibFunc(&decompress, libOnce.handle, "OodleLZ_Decompress")
		purego.RegisterLibFunc(&getCompressionLevelName, libOnce.handle, "OodleLZ_CompressionLevel_GetName")
		purego.RegisterLibFunc(&getCompressorName, libOnce.handle, "OodleLZ_Compressor_GetName")
	})

	return libOnce.handle, libOnce.err
}
