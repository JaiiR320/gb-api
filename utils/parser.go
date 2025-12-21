package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Parser is a utility for reading primitive typed data from binary data
type Parser struct {
	Reader    io.Reader
	ByteOrder binary.ByteOrder
}

// NewParser creates a new binary parser from an io.Reader
func NewParser(Reader io.Reader, ByteOrder binary.ByteOrder) *Parser {
	return &Parser{
		Reader:    Reader,
		ByteOrder: ByteOrder,
	}
}

func (p *Parser) SetPosition(offset int64, whence int) (int64, error) {
	if seeker, ok := p.Reader.(io.Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	return 0, fmt.Errorf("reader does not support seeking")
}

// GetByte reads a single byte (uint8)
func (p *Parser) GetUInt8() (byte, error) {
	var val uint8
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	return val, err
}

// GetShort reads an int16
func (p *Parser) GetInt16() (int16, error) {
	var val int16
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	return val, err
}

// GetUShort reads a uint16
func (p *Parser) GetUInt16() (uint16, error) {
	var val uint16
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	return val, err
}

// GetInt reads an int32
func (p *Parser) GetInt32() (int32, error) {
	var val int32
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	return val, err
}

// GetUInt reads a uint32
func (p *Parser) GetUInt32() (uint32, error) {
	var val uint32
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	return val, err
}

// GetFloat reads a float32
func (p *Parser) GetFloat32() (float32, error) {
	var val uint32
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(val), nil
}

// GetDouble reads a float64
func (p *Parser) GetFloat64() (float64, error) {
	var val uint64
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(val), nil
}

// GetLong reads a uint64
func (p *Parser) GetUInt64() (uint64, error) {
	var val uint64
	err := binary.Read(p.Reader, p.ByteOrder, &val)
	return val, err
}

// GetString reads a null-terminated string with optional length limit
func (p *Parser) GetString(maxLen int) (string, error) {
	s := ""
	var c byte
	var err error
	count := 0

	for {
		c, err = p.GetUInt8()
		if err != nil {
			return s, err
		}
		if c == 0 {
			break
		}
		s += string(c)
		count++
		if maxLen > 0 && count >= maxLen {
			break
		}
	}
	return s, nil
}

// GetFixedLengthString reads a fixed-length string, excluding null bytes
func (p *Parser) GetFixedLengthString(length int) (string, error) {
	s := ""
	for i := 0; i < length; i++ {
		c, err := p.GetUInt8()
		if err != nil {
			return s, err
		}
		if c > 0 {
			s += string(c)
		}
	}
	return s, nil
}

// GetFixedLengthTrimmedString reads a fixed-length string, excluding control characters and spaces
func (p *Parser) GetFixedLengthTrimmedString(length int) (string, error) {
	s := ""
	for i := 0; i < length; i++ {
		c, err := p.GetUInt8()
		if err != nil {
			return s, err
		}
		if c > 32 {
			s += string(c)
		}
	}
	return s, nil
}

// ReadStruct reads binary data directly into a struct
// The struct fields must be exported and of compatible types
func (p *Parser) ReadStruct(data interface{}) error {
	return binary.Read(p.Reader, p.ByteOrder, data)
}

// ReadMultiple reads a sequence of values in bulk
// Returns error if any read fails
func (p *Parser) ReadMultiple(values ...interface{}) error {
	for _, val := range values {
		if err := binary.Read(p.Reader, p.ByteOrder, val); err != nil {
			return err
		}
	}
	return nil
}
