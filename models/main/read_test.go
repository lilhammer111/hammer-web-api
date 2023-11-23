package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"
)

var filePath = "go.md"

var reciever string

// 分块读取
func BenchmarkChunkRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, _ := os.Open(filePath)
		buf := make([]byte, 1024)
		var buffer bytes.Buffer
		for {
			n, err := file.Read(buf)
			if err != nil {
				if err == io.EOF {
					buffer.Write(buf[:n])
					break
				}
				file.Close()
				b.Fatal(err)
			}
			buffer.Write(buf[:n])
		}
		reciever = buffer.String()
		file.Close()
	}
}

// 使用 io.Copy
func BenchmarkIOCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, _ := os.Open(filePath)

		var buffer bytes.Buffer
		_, err := io.Copy(&buffer, file)
		file.Close()
		if err != nil {
			b.Fatal(err)
		}
		reciever = buffer.String()
	}
}

// 使用 os.ReadFile
func BenchmarkReadFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, err := os.ReadFile(filePath)
		if err != nil {
			b.Fatal(err)
		}
		reciever = string(file)
	}
}

// 使用 bufio.Reader
func BenchmarkBufioReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, _ := os.Open(filePath)
		reader := bufio.NewReader(file)
		var buffer bytes.Buffer
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					buffer.Write(line) // 确保最后一行被写入
					break
				}
				file.Close()
				b.Fatal(err)
			}
			buffer.Write(line)
		}
		file.Close()

		reciever = buffer.String()
	}
}

//goos: linux
//goarch: amd64
//pkg: hammer-web-api/models/main
//cpu: 12th Gen Intel(R) Core(TM) i5-12400
//BenchmarkChunkRead-12              10000            127093 ns/op          179313 B/op         11 allocs/op
//BenchmarkIOCopy-12                 15879             87800 ns/op          179873 B/op         13 allocs/op
//BenchmarkReadFile-12               30106             50109 ns/op           98624 B/op          6 allocs/op
//BenchmarkBufioReader-12             4272            380694 ns/op          237801 B/op       2401 allocs/op
//PASS
//ok      hammer-web-api/models/main      9.197s
