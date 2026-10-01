package record

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"time"
)

func checkRoundTrip(t *testing.T, want Record, buf []byte) {
	t.Helper()

	got, n, err := Deserialize(buf)
	if err != nil {
		t.Fatalf("Deserialize() error = %v, want nil", err)
	}
	if n != len(buf) {
		t.Errorf("Deserialize() n = %d, want %d (len of serialized buffer)", n, len(buf))
	}
	if got.Key != want.Key {
		t.Errorf("Key = %q, want %q", got.Key, want.Key)
	}
	if got.Tombstone != want.Tombstone {
		t.Errorf("Tombstone = %v, want %v", got.Tombstone, want.Tombstone)
	}
	// .Equal() upoređuje instant vremena, a == upoređuje interni prikaz
	// time.Time strukture (koji nosi i monotonic/location informaciju).
	// Serijalizacija/deserijalizacija ide kroz Unix()+Nanosecond(), pa se
	// monotonic deo gubi — isti trenutak, različita struktura, == bi lažno pukao.
	if !got.Timestamp.Equal(want.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, want.Timestamp)
	}
	// bytes.Equal, a ne == (slice se ne poredi sa ==) i ne reflect.DeepEqual
	// (nil vs. []byte{} se serijalizuje/deserijalizuje kao []byte{}, a
	// DeepEqual bi to smatrao različitim od nil; bytes.Equal ih smatra istim).
	if !bytes.Equal(got.Value, want.Value) {
		t.Errorf("Value = %v, want %v", got.Value, want.Value)
	}
}

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	longValue := bytes.Repeat([]byte("abcdefgh"), 500) // ~4KB

	cases := []struct {
		name   string
		record Record
	}{
		{
			name:   "obican zapis",
			record: NewRecord("key", []byte("value")),
		},
		{
			name:   "prazna vrednost",
			record: NewRecord("key", []byte{}),
		},
		{
			name:   "tombstone",
			record: NewTombstone("key"),
		},
		{
			name:   "kljuc sa nasim slovima",
			record: NewRecord("ključ", []byte("vrednost")),
		},
		{
			name:   "vrednost sa nula-bajtovima",
			record: NewRecord("key", []byte{0, 1, 0, 0, 2, 0}),
		},
		{
			name:   "duza vrednost (nekoliko KB)",
			record: NewRecord("key", longValue),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := c.record.Serialize()
			checkRoundTrip(t, c.record, buf)
		})
	}
}

func TestDeserializeCorruptedData(t *testing.T) {
	t.Run("korupcija u kljucu", func(t *testing.T) {
		r := NewRecord("testkey", []byte("testvalue"))
		buf := r.Serialize()
		buf[KeyOffset] ^= 0xFF

		_, _, err := Deserialize(buf)
		if !errors.Is(err, ErrDamagedData) {
			t.Errorf("err = %v, want ErrDamagedData", err)
		}
	})

	t.Run("korupcija u vrednosti", func(t *testing.T) {
		r := NewRecord("testkey", []byte("testvalue"))
		buf := r.Serialize()
		buf[KeyOffset+len(r.Key)] ^= 0xFF

		_, _, err := Deserialize(buf)
		if !errors.Is(err, ErrDamagedData) {
			t.Errorf("err = %v, want ErrDamagedData", err)
		}
	})

	// Pitanje: šta se dešava ako se korumpira bajt u polju DUŽINE ključa,
	// umesto u samom ključu?
	//
	// Odgovor: zavisi od smera promene vrednosti keySize.
	//
	//   - Ako korupcija UVEĆA keySize: Deserialize računa
	//     dataLen = HeaderLen + keySize + valueSize, što sada premašuje
	//     stvarnu dužinu ulaznog bafera. Provera "inputLen64 < dataLen" to
	//     uhvati PRE nego što se CRC i uopšte proveri -> ErrShortInput,
	//     ne ErrDamagedData, iako podaci nisu zaista "kratki", samo je
	//     dužina polja pokvarena.
	//
	//   - Ako korupcija UMANJI keySize: dataLen je i dalje <= dužina
	//     bafera, pa se prolazi kroz proveru dužine, ali se CRC računa
	//     nad pogrešnim (kraćim) opsegom bajtova nego što je korišćen
	//     prilikom serijalizacije -> checksum se ne poklapa -> ErrDamagedData.
	//
	// U oba slučaja Deserialize NIKAD ne panikuje i nikad ne vrati tiho
	// pogrešne podatke - dužinske provere se izvršavaju pre bilo kakvog
	// sečenja slice-a van granica, a KeyLengthOffset je deo regiona koji
	// je pokriven CRC-om, tako da se korupcija uvek detektuje, samo kroz
	// drugu grešku u zavisnosti od smera.

	t.Run("korupcija duzine kljuca - uvecanje (ErrShortInput)", func(t *testing.T) {
		r := NewRecord("test", []byte("val")) // len("test") == 4, LSB(4) == 0
		buf := r.Serialize()
		buf[KeyLengthOffset+KeySizeLen-1] ^= 0x01 // 4 -> 5, keySize uvecan

		_, _, err := Deserialize(buf)
		if !errors.Is(err, ErrShortInput) {
			t.Errorf("err = %v, want ErrShortInput", err)
		}
	})

	t.Run("korupcija duzine kljuca - umanjenje (ErrDamagedData)", func(t *testing.T) {
		r := NewRecord("tests", []byte("val")) // len("tests") == 5, LSB(5) == 1
		buf := r.Serialize()
		buf[KeyLengthOffset+KeySizeLen-1] ^= 0x01 // 5 -> 4, keySize umanjen

		_, _, err := Deserialize(buf)
		if !errors.Is(err, ErrDamagedData) {
			t.Errorf("err = %v, want ErrDamagedData", err)
		}
	})
}

func TestDeserializeShortInput(t *testing.T) {
	r := NewRecord("testkey", []byte("testvalue"))
	buf := r.Serialize()

	t.Run("odsecen poslednji bajt", func(t *testing.T) {
		truncated := buf[:len(buf)-1]
		_, _, err := Deserialize(truncated)
		if !errors.Is(err, ErrShortInput) {
			t.Errorf("err = %v, want ErrShortInput", err)
		}
	})

	t.Run("odsecen na 10 bajtova", func(t *testing.T) {
		truncated := buf[:10]
		_, _, err := Deserialize(truncated)
		if !errors.Is(err, ErrShortInput) {
			t.Errorf("err = %v, want ErrShortInput", err)
		}
	})
}

func TestDeserializeTwoRecordsInSequence(t *testing.T) {
	rec1 := NewRecord("abc", []byte("hello"))
	rec2 := NewRecord("xyz", []byte("world!!"))

	buf1 := rec1.Serialize()
	buf2 := rec2.Serialize()

	combined := make([]byte, 0, len(buf1)+len(buf2))
	combined = append(combined, buf1...)
	combined = append(combined, buf2...)

	got1, n1, err := Deserialize(combined)
	if err != nil {
		t.Fatalf("Deserialize(first) error = %v, want nil", err)
	}
	if n1 != len(buf1) {
		t.Fatalf("Deserialize(first) n = %d, want %d", n1, len(buf1))
	}
	if got1.Key != rec1.Key || !bytes.Equal(got1.Value, rec1.Value) {
		t.Errorf("first record = %+v, want %+v", got1, rec1)
	}

	got2, n2, err := Deserialize(combined[n1:])
	if err != nil {
		t.Fatalf("Deserialize(second) error = %v, want nil", err)
	}
	if n2 != len(buf2) {
		t.Fatalf("Deserialize(second) n = %d, want %d", n2, len(buf2))
	}
	if got2.Key != rec2.Key || !bytes.Equal(got2.Value, rec2.Value) {
		t.Errorf("second record = %+v, want %+v", got2, rec2)
	}
}

// Ovaj test bi uhvatio bag sa offsetima od pre par dana: offset poslednjeg
// polja plus njegova veličina mora biti jednako HeaderLen, inače Key/Value
// počinju da se čitaju sa pogrešnog mesta u baferu.
func TestLastFieldOffsetPlusSizeEqualsHeaderLen(t *testing.T) {
	got := ValueLenghtOffset + ValueSizeLen
	if got != HeaderLen {
		t.Errorf("ValueLenghtOffset + ValueSizeLen = %d, want HeaderLen = %d", got, HeaderLen)
	}
	if KeyOffset != HeaderLen {
		t.Errorf("KeyOffset = %d, want HeaderLen = %d", KeyOffset, HeaderLen)
	}
}

func TestTombstoneIgnoresValue(t *testing.T) {
	r := Record{Timestamp: time.Now(), Tombstone: true, Key: "key", Value: []byte("ignored")}
	buf := r.Serialize()

	if len(buf) != HeaderLen+len("key") {
		t.Errorf("len(buf) = %d, want %d", len(buf), HeaderLen+len("key"))
	}
	if valueSize := binary.BigEndian.Uint64(buf[ValueLenghtOffset:]); valueSize != 0 {
		t.Errorf("ValueLenghtOffset field = %d, want 0", valueSize)
	}

	got, _, err := Deserialize(buf)
	if err != nil {
		t.Fatalf("Deserialize() error = %v, want nil", err)
	}
	if !got.Tombstone {
		t.Errorf("Tombstone = %v, want true", got.Tombstone)
	}
	if len(got.Value) != 0 {
		t.Errorf("len(Value) = %d, want 0", len(got.Value))
	}
}
