package main

import (
	_ "bufio"
	"encoding/csv"
	"fmt"
	"os" // 파일 읽고 쓰기는 기본적으로 os에서 제공
)

func errCheck1(e error) {
	if e != nil {
		panic(e)
	}
}

func errCheck2(e error) {
	if e != nil {
		fmt.Println(e)
		return
	}
}

func main() {
	//FIle Write
	//패키지 저장소를 통해서 Excel 등 다양한 파일 형식 쓰기, 읽기 가능
	//cvs file write example
	//package github 주서: https://github.com/tealeg/xlsx
	//bufio : 파일이 용량이 클 경우 버퍼 사용 권장

	//파일 생성
	file, err := os.Create("test_write.csv")
	errCheck1(err)
	//리소스 해제
	defer file.Close()

	//csv write 생성
	wr := csv.NewWriter(file)
	//wr := csv.NewWriter(bufio.NewWriter(file))
	//csv 내용 쓰기
	wr.Write([]string{"kim", "4.8"})
	wr.Write([]string{"Lee", "4.2"})
	wr.Write([]string{"Park", "4.3"})
	wr.Write([]string{"Cho", "4.1"})
	wr.Write([]string{"Hong", "4.7"})

	wr.Flush() // 버퍼 -> 파일로 쓰기

	fi, err := file.Stat()
	errCheck1(err)

	fmt.Printf("CSV 쓰기 작업 후 파일 크기(%d byte)\n", fi.Size())
	fmt.Println("CSV 파일 명", fi.Name())
	fmt.Println("파일 권한", fi.Mode())

}
