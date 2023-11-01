package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	DATA_RECORD_PATH string = "./records/segment1.bin"
	KEY_INDEX        int64  = 36 + 1
)

func ErrorChecker(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

type RecordLength struct {
	bytePosition int64
	recordLength int64
}

func NewRecordLength(bytePosition, recordLength int64) RecordLength {
	return RecordLength{
		bytePosition: bytePosition,
		recordLength: recordLength,
	}
}

type Jumper struct {
	lastByte int64
	index    map[string]RecordLength
}

func (j *Jumper) Save(key string, data string) {
	rowLen := int64(len(data)) + KEY_INDEX + 1
	j.index[key] = NewRecordLength(j.lastByte, rowLen)
	j.lastByte += rowLen
}

func (j Jumper) GetByIndex(index string) RecordLength {
	return j.index[index]
}

func NewSaver() *Jumper {
	return &Jumper{
		index: make(map[string]RecordLength),
	}
}

func main() {
	keyIndex := NewSaver()

	inputCh := make(chan string)
	go Listener(inputCh)

	for input := range inputCh {
		keyIndex.Execute(input)
	}
}

func (j *Jumper) Execute(input string) {
	funcName, args := StringParser(input)

	switch funcName {
	case "read":
		j.Read(args)
		fmt.Printf("postgres=# ")
	case "write":
		j.CreateNewRecord(args)
	}
}

func StringParser(input string) (string, string) {
	index := strings.Index(input, "(")

	if index == -1 {
		return "", ""
	}

	return input[:index], input[index+1 : len(input)-1]
}

func (j Jumper) Read(key string) {
	start := time.Now()
	file, err := os.Open(DATA_RECORD_PATH)
	if err != nil {
		fmt.Println("Greška prilikom otvaranja datoteke:", err)
		return
	}
	defer file.Close()

	record := j.GetByIndex(key)

	_, err = file.Seek(record.bytePosition, 0) // Preskače određeni broj bajtova da bi došao do linije 8
	if err != nil {
		fmt.Println("Greška prilikom preskakanja bajtova:", err)
		return
	}

	var data []byte // Čita se linija određenog broja bajtova
	data = make([]byte, record.recordLength)
	_, err = file.Read(data)
	if err != nil {
		fmt.Println("Greška prilikom čitanja linije:", err)
		return
	}

	fmt.Println(time.Since(start))

	fmt.Printf("%s\n", data)
}

func RandomNewData(maxLength int) int {
	return rand.Intn(maxLength)
}

func (j *Jumper) CreateNewRecord(data string) {
	uuid, err := generateUUIDv4()
	ErrorChecker(err)

	row := fmt.Sprintf("%s %s\n", uuid, data)
	j.Save(uuid, data)

	AddNewRecord([]byte(row))
}

func WriteFile() (*os.File, error) {
	return os.OpenFile(DATA_RECORD_PATH, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}

func AddNewRecord(data []byte) {
	file, err := WriteFile()
	ErrorChecker(err)

	if err = binary.Write(file, binary.LittleEndian, &data); err != nil {
		panic(err)
	}
}

func generateUUIDv4() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
	return uuidStr, nil
}

func Listener(inputCh chan string) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case s := <-sig:
			fmt.Printf("Primljen signal: %v. Završavam program.\n", s)
			close(inputCh) // Zatvaramo kanal kada primimo signal
			return
		default:
			var input string
			fmt.Printf("postgres=# ")
			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Println("Greška prilikom čitanja komande:", err)
				continue
			}
			inputCh <- input
		}
	}
}
