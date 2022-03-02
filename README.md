# go-oodle

Go wrapper for [Oodle Data Compression](http://www.radgametools.com/oodle.htm)

## Usage

Put `oo2core_9_win64.dll` in the directory with your application (or use the built-in downloader)

### Compress
```go
compressedData, err := oodle.Compress(data, oodle.AlgoKraken, oodle.CompressionLevelOptimal3)
```

### Decompress
```go
decompressedData, err := oodle.Decompress(compressedData, outputSize))
```

### Download
```go
if !oodle.IsDllExist() {
	err := oodle.Download()
	if err != nil {
		log.Fatalf("no oo2core_9_win64.dll")
	}
}
```


