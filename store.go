package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	FOLDER_NAME          = "chat"
	DB_NAME              = "chat.db"
	MAXIMUM_MESSAGE_SIZE = 4096
	HEADER_SIZE          = 16 // 4 + 4 + 4 + 4 = 16 bytes
	CONTENT_SIZE         = 22 + MAXIMUM_MESSAGE_SIZE
)

type Header struct {
	Magic   [4]byte // Identifier for CHAT ("CHAT")
	Version uint32
	Record  uint32
	Count   uint32
}

type Content struct {
	Id        uint32 // 4 bytes
	CreatedAt int64  // 8 bytes
	UpdatedAt int64  // 8 bytes
	Length    uint16 // 2 bytes
	Content   [MAXIMUM_MESSAGE_SIZE]byte
}

type Storage struct {
	stdOut chan string
	header Header
}

type Store interface {
	Check() error
	Initialize() error
	Store(id uint32, content Content) uint32
	Get(id uint32) string
	GetIds() []uint32
	GetOffset(id uint32) uint32
}

func (s *Storage) GetOffset(id uint32) uint32 {
	return HEADER_SIZE + (id * CONTENT_SIZE)
}

func (h *Header) GenerateId() uint32 {
	return h.Count + 1
}

func (s *Storage) Check() error {
	file := filepath.Join(FOLDER_NAME, DB_NAME)
	if _, error := os.OpenFile(file, os.O_RDONLY, 0644); error != nil {
		return error
	}
	return nil
}

func (s *Storage) Initialize() error {
	if err := os.MkdirAll(FOLDER_NAME, 0755); err != nil {
		fmt.Println("Error creating folder: ", err)
		return err
	}

	go func() {
		s.stdOut <- "Creating database..."
	}()

	path := filepath.Join(FOLDER_NAME, DB_NAME)
	file, error := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if os.IsExist(error) {
		s.loadHeader()
		go func() {
			s.stdOut <- "Database already exists"
		}()
		return nil
	}

	if error != nil {
		fmt.Println("Error initializing storage:", error)
		return error
	}

	defer file.Close()

	s.header = Header{
		Magic:   [4]byte{'C', 'H', 'A', 'T'},
		Version: 1,
		Record:  0,
		Count:   0,
	}
	s.saveHeader()

	go func() {
		s.stdOut <- "Database created successfully"
	}()

	return nil
}

func (s *Storage) loadHeader() error {
	path := filepath.Join(FOLDER_NAME, DB_NAME)
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, HEADER_SIZE)
	if _, err := file.Read(buf); err != nil {
		return err
	}

	copy(s.header.Magic[:], buf[:4])
	s.header.Version = binary.BigEndian.Uint32(buf[4:8])
	s.header.Record = binary.BigEndian.Uint32(buf[8:12])
	s.header.Count = binary.BigEndian.Uint32(buf[12:16])

	return nil
}

func (s *Storage) saveHeader() error {
	path := filepath.Join(FOLDER_NAME, DB_NAME)
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, HEADER_SIZE)
	copy(buf[:4], s.header.Magic[:])
	binary.BigEndian.PutUint32(buf[4:8], s.header.Version)
	binary.BigEndian.PutUint32(buf[8:12], s.header.Record)
	binary.BigEndian.PutUint32(buf[12:16], s.header.Count)

	file.Seek(0, io.SeekStart)
	if _, err := file.Write(buf); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Store(id uint32, content Content) (uint32, error) {
	if id == 0 {
		id = s.header.GenerateId()
	}
	offset := s.GetOffset(id)

	// Write content to file
	path := filepath.Join(FOLDER_NAME, DB_NAME)
	file, error := os.OpenFile(path, os.O_WRONLY, 0644)
	if error != nil {
		fmt.Println("Error opening file:", error)
		return 0, error
	}
	defer file.Close()

	buffer := make([]byte, CONTENT_SIZE)
	binary.BigEndian.PutUint32(buffer[:4], id)
	binary.BigEndian.PutUint64(buffer[4:12], uint64(content.CreatedAt))
	binary.BigEndian.PutUint64(buffer[12:20], uint64(content.UpdatedAt))
	binary.BigEndian.PutUint16(buffer[20:22], content.Length)
	copy(buffer[22:], content.Content[:content.Length])

	if _, error := file.WriteAt(buffer, int64(offset)); error != nil {
		fmt.Println("Error writing to file:", error)
		return 0, error
	}

	if id == 0 {
		s.header.Count++
		s.header.Record++
		s.saveHeader()
	}

	go func() {
		s.stdOut <- fmt.Sprintf("Stored message with ID %d", id)
	}()

	return id, nil
}

func (s *Storage) Get(id int64) string {
	return ""
}
