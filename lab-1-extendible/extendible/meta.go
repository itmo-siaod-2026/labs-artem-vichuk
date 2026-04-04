package extendible

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	startDepth = 1

	metaMagic      uint64 = 0x4558544d45544131
	metaVersion    uint64 = 1
	metaHeaderSize        = 6 * 8 // magic, version, globalDepth, bucketLimit, nextBucketID, dirCount
)

type Meta struct {
	GlobalDepth  uint64
	BucketLimit  uint64
	NextBucketID uint64
	Directory    []uint64
}

func NewMeta(bucketLimit uint64) *Meta {
	size := uint64(1 << startDepth)
	return &Meta{
		GlobalDepth:  startDepth,
		BucketLimit:  bucketLimit,
		NextBucketID: 3,
		Directory:    make([]uint64, size),
	}
}

func metaFilePath(baseDir string) string {
	return filepath.Join(baseDir, "meta.dat")
}

func LoadMeta(baseDir string) (*Meta, error) {
	path := metaFilePath(baseDir)

	f, data, err := mmapFileRead(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = closeMappedFile(f, data, false)
	}()

	if len(data) < metaHeaderSize {
		return nil, fmt.Errorf("meta file too small")
	}

	magic := binary.LittleEndian.Uint64(data[0:8])
	version := binary.LittleEndian.Uint64(data[8:16])
	if magic != metaMagic {
		return nil, fmt.Errorf("invalid meta magic")
	}
	if version != metaVersion {
		return nil, fmt.Errorf("unsupported meta version: %d", version)
	}

	globalDepth := binary.LittleEndian.Uint64(data[16:24])
	bucketLimit := binary.LittleEndian.Uint64(data[24:32])
	nextBucketID := binary.LittleEndian.Uint64(data[32:40])
	dirCount := binary.LittleEndian.Uint64(data[40:48])

	wantSize := metaHeaderSize + int(dirCount)*8
	if len(data) != wantSize {
		return nil, fmt.Errorf("invalid meta size: got=%d want=%d", len(data), wantSize)
	}
	if dirCount == 0 {
		return nil, fmt.Errorf("empty directory in meta")
	}

	dir := make([]uint64, dirCount)
	offset := metaHeaderSize
	for i := uint64(0); i < dirCount; i++ {
		dir[i] = binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8
	}

	return &Meta{
		GlobalDepth:  globalDepth,
		BucketLimit:  bucketLimit,
		NextBucketID: nextBucketID,
		Directory:    dir,
	}, nil
}

func (m *Meta) Save(baseDir string) error {
	if len(m.Directory) == 0 {
		return fmt.Errorf("cannot save empty directory")
	}

	path := metaFilePath(baseDir)
	size := metaHeaderSize + len(m.Directory)*8

	f, data, err := mmapFileWrite(path, size)
	if err != nil {
		return err
	}
	defer func() {
		_ = closeMappedFile(f, data, true)
	}()

	binary.LittleEndian.PutUint64(data[0:8], metaMagic)
	binary.LittleEndian.PutUint64(data[8:16], metaVersion)
	binary.LittleEndian.PutUint64(data[16:24], m.GlobalDepth)
	binary.LittleEndian.PutUint64(data[24:32], m.BucketLimit)
	binary.LittleEndian.PutUint64(data[32:40], m.NextBucketID)
	binary.LittleEndian.PutUint64(data[40:48], uint64(len(m.Directory)))

	offset := metaHeaderSize
	for _, bucketID := range m.Directory {
		binary.LittleEndian.PutUint64(data[offset:offset+8], bucketID)
		offset += 8
	}

	return nil
}

func (m *Meta) DirectorySize() uint64 {
	return uint64(len(m.Directory))
}

func (m *Meta) GetBucketID(idx uint64) uint64 {
	if idx >= uint64(len(m.Directory)) {
		return 0
	}
	return m.Directory[idx]
}

func (m *Meta) SetBucketID(idx uint64, bucketID uint64) {
	if idx >= uint64(len(m.Directory)) {
		return
	}
	m.Directory[idx] = bucketID
}

func (m *Meta) DoubleDirectory() {
	oldSize := len(m.Directory)
	newDirectory := make([]uint64, oldSize*2)

	for i := 0; i < oldSize; i++ {
		newDirectory[i] = m.Directory[i]
		newDirectory[i+oldSize] = m.Directory[i]
	}

	m.Directory = newDirectory
	m.GlobalDepth++
}

func (m *Meta) RepointAfterSplit(oldBucketID, newBucketID uint64, oldLocalDepth uint64) {
	for idx := uint64(0); idx < uint64(len(m.Directory)); idx++ {
		if m.Directory[idx] != oldBucketID {
			continue
		}

		bit := (idx >> oldLocalDepth) & 1
		if bit == 1 {
			m.Directory[idx] = newBucketID
		}
	}
}

func (m *Meta) CanShrink() bool {
	if m.GlobalDepth <= startDepth || len(m.Directory)%2 != 0 {
		return false
	}

	half := len(m.Directory) / 2
	for i := 0; i < half; i++ {
		if m.Directory[i] != m.Directory[i+half] {
			return false
		}
	}
	return true
}

func (m *Meta) ShrinkDirectory() {
	for m.CanShrink() {
		half := len(m.Directory) / 2
		m.Directory = append([]uint64(nil), m.Directory[:half]...)
		m.GlobalDepth--
	}
}

func EnsureBaseDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func (m *Meta) Index(hash uint64) uint64 {
	mask := uint64((1 << m.GlobalDepth) - 1)
	return hash & mask
}
