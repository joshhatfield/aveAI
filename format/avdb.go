package format

import (
	"encoding/gob"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"aveAI/store"
)

const (
	magic     = "AVE0"
	version   = uint16(1)
)

// header represents the .avdb file header structure.
type header struct {
	Magic       [4]byte
	Version     uint16
	EntryCount  uint32
	IndexOffset uint64
}

// Save serializes the store to a .avdb file.
func Save(path string, s *store.Store) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create .avdb: %w", err)
	}
	defer f.Close()

	entries := s.All()

	// Write header (placeholder, will rewrite at end)
	h := header{
		Version:     version,
		EntryCount:  uint32(len(entries)),
	}
	copy(h.Magic[:], magic)

	if err := binary.Write(f, binary.LittleEndian, h); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write entries
	enc := gob.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			return fmt.Errorf("encode entry %d: %w", e.ID, err)
		}
	}

	// Record index position and write index
	indexOffset, _ := f.Seek(0, io.SeekCurrent)
	h.IndexOffset = uint64(indexOffset)

	// Write index data (gob-encoded for now — will expand in Phase 2/3)
	type indexData struct {
		Entries []store.Entry
	}
	indexEnc := gob.NewEncoder(f)
	if err := indexEnc.Encode(indexData{Entries: entries}); err != nil {
		return fmt.Errorf("encode index: %w", err)
	}

	// Rewrite header with correct index offset
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("seek to rewrite header: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, h); err != nil {
		return fmt.Errorf("rewrite header: %w", err)
	}

	// Write entry count at entry start (for future use)
	// TODO: store entry offsets for random access

	return nil
}

// Load deserializes a .avdb file into a store.
func Load(path string) (*store.Store, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open .avdb: %w", err)
	}
	defer f.Close()

	// Read and validate header
	var h header
	if err := binary.Read(f, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	if string(h.Magic[:]) != magic {
		return nil, fmt.Errorf("invalid magic: expected %s, got %s", magic, string(h.Magic[:]))
	}
	if h.Version != version {
		return nil, fmt.Errorf("unsupported version: %d", h.Version)
	}

	// Read entries
	s := store.New()
	dec := gob.NewDecoder(f)
	for i := uint32(0); i < h.EntryCount; i++ {
		var e store.Entry
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("decode entry %d: %w", i, err)
		}
		s.Add(e)
	}

	return s, nil
}
