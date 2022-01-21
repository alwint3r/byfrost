package main

import (
	"encoding/binary"
	"errors"
	"log"
)

type State int64

const (
	WaitHeader State = iota
	ParsingFileNameSize
	ParsingFileName
	ParsingFileContentSize
	ParsingContent
	Finished
	Unknown
)

const HeaderSize = 3
const FileNameSize = 2
const FileContentSize = 4

type ByfrostServerContext struct {
	State          State
	FileName       string
	FileNameLength uint16
	FileContent    []byte
	FileSize       uint32
	Buffer         []byte
	BufferSize     uint32
}

func InitByfrostServerContext() *ByfrostServerContext {
	return &ByfrostServerContext{
		State:          WaitHeader,
		Buffer:         make([]byte, 0),
		BufferSize:     0,
		FileName:       "",
		FileNameLength: 0,
		FileContent:    make([]byte, 0),
		FileSize:       0,
	}
}

func (f *ByfrostServerContext) ResetState() {
	f.State = WaitHeader
	f.FileContent = make([]byte, 0)
	f.FileSize = 0
	f.Buffer = make([]byte, 0)
	f.BufferSize = 0
	f.FileName = ""
	f.FileNameLength = 0
}

func (f *ByfrostServerContext) Process(b byte) (State, error) {
	if f.State == WaitHeader {
		if f.BufferSize < HeaderSize {
			f.Buffer = append(f.Buffer, b)
			f.BufferSize++
		}

		if f.BufferSize == HeaderSize {
			if f.Buffer[0] != 0x21 && f.Buffer[1] != 0x12 {
				log.Println("Invalid header")

				f.ResetState()
				return WaitHeader, errors.New("Invalid header")
			}

			if f.Buffer[2] != 0x01 {
				log.Printf("Invalid command: %d", f.Buffer[2])

				f.ResetState()
				return WaitHeader, errors.New("Invalid command")
			}

			log.Println("Header received")

			f.State = ParsingFileNameSize
			f.FileNameLength = 0
			f.FileName = ""
			f.Buffer = make([]byte, 0)
			f.BufferSize = 0

			return f.State, nil
		}
	}

	if f.State == ParsingFileNameSize {
		if f.BufferSize < FileNameSize {
			f.Buffer = append(f.Buffer, b)
			f.BufferSize++
		}

		if f.BufferSize == FileNameSize {
			f.FileNameLength = binary.LittleEndian.Uint16(f.Buffer)

			log.Printf("File name length: %d", f.FileNameLength)
			if f.FileNameLength == 0 {
				log.Println("Invalid file name length")

				f.ResetState()

				return ParsingFileNameSize, errors.New("Invalid file name length")
			}

			f.State = ParsingFileName
			f.Buffer = make([]byte, 0)
			f.BufferSize = 0

			return f.State, nil
		}
	}

	if f.State == ParsingFileName {
		if f.BufferSize < uint32(f.FileNameLength) {
			f.Buffer = append(f.Buffer, b)
			f.BufferSize++
		}

		if f.BufferSize == uint32(f.FileNameLength) {
			f.FileName = string(f.Buffer)

			log.Println("File name:", f.FileName)

			f.State = ParsingFileContentSize
			f.Buffer = make([]byte, 0)
			f.BufferSize = 0

			return ParsingFileContentSize, nil
		}
	}

	if f.State == ParsingFileContentSize {
		if f.BufferSize < FileContentSize {
			f.Buffer = append(f.Buffer, b)
			f.BufferSize++
		}

		if f.BufferSize == FileContentSize {
			f.FileSize = binary.LittleEndian.Uint32(f.Buffer)
			f.Buffer = make([]byte, 0)
			f.BufferSize = 0

			log.Printf("File size: %d", f.FileSize)

			f.State = ParsingContent
			f.Buffer = make([]byte, 0)
			f.BufferSize = 0

			return f.State, nil
		}
	}

	if f.State == ParsingContent {
		if f.BufferSize < uint32(f.FileSize) {
			f.Buffer = append(f.Buffer, b)
			f.BufferSize++
		}

		if f.BufferSize == uint32(f.FileSize) {
			f.FileContent = append(f.FileContent, f.Buffer...)
			f.Buffer = make([]byte, 0)
			f.BufferSize = 0

			log.Println("Finish")

			f.State = Finished

			return f.State, nil
		}
	}

	return f.State, nil
}
