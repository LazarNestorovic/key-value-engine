package record

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"time"
)

const (
	CRCLen       int = 4
	TimestampLen int = 16
	TombstoneLen int = 1
	KeySizeLen   int = 8
	ValueSizeLen int = 8
	HeaderLen    int = CRCLen + TimestampLen + TombstoneLen + KeySizeLen + ValueSizeLen

	CRCOffset         int = 0
	TimestampOffset   int = CRCOffset + CRCLen
	TombstoneOffset   int = TimestampOffset + TimestampLen
	KeyLengthOffset   int = TombstoneOffset + TombstoneLen
	ValueLenghtOffset int = KeyLengthOffset + KeySizeLen
	KeyOffset         int = HeaderLen
)

var (
	ErrShortInput  = errors.New("record: input too short")
	ErrDamagedData = errors.New("record: checksum mismatch")
)

type Record struct {
	Timestamp time.Time
	Tombstone bool
	Key       string
	Value     []byte
}

func NewRecord(key string, value []byte) Record {
	return Record{
		Timestamp: time.Now(),
		Tombstone: false,
		Key:       key,
		Value:     value,
	}
}

func NewTombstone(key string) Record {
	return Record{
		Timestamp: time.Now(),
		Tombstone: true,
		Key:       key,
		Value:     nil,
	}
}

func (r Record) Serialize() []byte {
	keyLen := len(r.Key)
	recordLength := HeaderLen + keyLen
	var valueLen int
	if !r.Tombstone {
		valueLen = len(r.Value)
		recordLength += valueLen
	}

	buf := make([]byte, recordLength)

	//Timestamp
	binary.BigEndian.PutUint64(buf[TimestampOffset:], uint64(r.Timestamp.Unix()))
	binary.BigEndian.PutUint64(buf[TimestampOffset+8:], uint64(r.Timestamp.Nanosecond()))

	//Tombstone
	if r.Tombstone {
		buf[TombstoneOffset] = 1
	}

	//Key and Value Size
	binary.BigEndian.PutUint64(buf[KeyLengthOffset:], uint64(keyLen))
	binary.BigEndian.PutUint64(buf[ValueLenghtOffset:], uint64(valueLen))

	//Key and Value
	copy(buf[KeyOffset:], r.Key)
	if !r.Tombstone {
		copy(buf[KeyOffset+keyLen:], r.Value)
	}

	binary.BigEndian.PutUint32(buf[CRCOffset:], crc32.ChecksumIEEE(buf[TimestampOffset:]))

	return buf
}

func Deserialize(input []byte) (Record, int, error) {
	if len(input) < HeaderLen {
		return Record{}, 0, ErrShortInput
	}

	inputLen64 := uint64(len(input))

	keySize := binary.BigEndian.Uint64(input[KeyLengthOffset:])
	valueSize := binary.BigEndian.Uint64(input[ValueLenghtOffset:])

	if uint64(keySize) > inputLen64 || uint64(valueSize) > inputLen64 {
		return Record{}, 0, ErrShortInput
	}

	dataLen := uint64(HeaderLen) + keySize + valueSize

	if inputLen64 < dataLen {
		return Record{}, 0, ErrShortInput
	}

	inputCRC := binary.BigEndian.Uint32(input[CRCOffset:])
	crc := crc32.ChecksumIEEE(input[TimestampOffset:dataLen])
	if inputCRC != crc {
		return Record{}, 0, ErrDamagedData
	}

	sec := binary.BigEndian.Uint64(input[TimestampOffset:])
	nsec := binary.BigEndian.Uint64(input[TimestampOffset+8:])
	tombstone := input[TombstoneOffset]
	keyOffset64 := uint64(KeyOffset)
	key := string(input[keyOffset64 : keyOffset64+keySize])
	var value []byte = nil
	if tombstone == 0 {
		value = make([]byte, valueSize)
		copy(value, input[keyOffset64+keySize:dataLen])
	}
	return Record{Timestamp: time.Unix(int64(sec), int64(nsec)),
		Tombstone: byteToBool(tombstone),
		Key:       key,
		Value:     value}, int(dataLen), nil
}

func byteToBool(b byte) bool {
	return b != 0
}
