package extendible

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
)

const (
	bucketMagic      uint64 = 0x4558544255434b31 // "EXTBUCK1"
	bucketVersion    uint64 = 1
	bucketHeaderSize        = 5 * 8 // magic, version, id, localDepth, count
	bucketEntrySize         = 16    // int64 key + int64 value
)

type Bucket struct {
	id         uint64
	localDepth uint64
	slots      map[int64]int64
}

func NewBucket(id uint64, localDepth uint64) *Bucket {
	return &Bucket{
		id:         id,
		localDepth: localDepth,
		slots:      make(map[int64]int64),
	}
}

func bucketFilePath(baseDir string, id uint64) string {
	return filepath.Join(baseDir, fmt.Sprintf("bucket_%020d.dat", id))
}

func LoadBucket(baseDir string, id uint64) (*Bucket, error) {
	path := bucketFilePath(baseDir, id)

	f, data, err := mmapFileRead(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = closeMappedFile(f, data, false)
	}()

	if len(data) < bucketHeaderSize {
		return nil, fmt.Errorf("bucket file too small")
	}

	magic := binary.LittleEndian.Uint64(data[0:8])
	version := binary.LittleEndian.Uint64(data[8:16])
	storedID := binary.LittleEndian.Uint64(data[16:24])
	localDepth := binary.LittleEndian.Uint64(data[24:32])
	count := binary.LittleEndian.Uint64(data[32:40])

	if magic != bucketMagic {
		return nil, fmt.Errorf("invalid bucket magic")
	}
	if version != bucketVersion {
		return nil, fmt.Errorf("unsupported bucket version: %d", version)
	}
	if storedID != id {
		return nil, fmt.Errorf("bucket id mismatch: got %d want %d", storedID, id)
	}

	wantSize := bucketHeaderSize + int(count)*bucketEntrySize
	if len(data) != wantSize {
		return nil, fmt.Errorf("invalid bucket size: got=%d want=%d", len(data), wantSize)
	}

	b := NewBucket(id, localDepth)
	offset := bucketHeaderSize
	for i := uint64(0); i < count; i++ {
		key := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		value := int64(binary.LittleEndian.Uint64(data[offset+8 : offset+16]))
		b.slots[key] = value
		offset += bucketEntrySize
	}

	return b, nil
}

func (b *Bucket) Save(baseDir string) error {
	path := bucketFilePath(baseDir, b.id)
	size := bucketHeaderSize + len(b.slots)*bucketEntrySize

	f, data, err := mmapFileWrite(path, size)
	if err != nil {
		return err
	}
	defer func() {
		_ = closeMappedFile(f, data, true)
	}()

	binary.LittleEndian.PutUint64(data[0:8], bucketMagic)
	binary.LittleEndian.PutUint64(data[8:16], bucketVersion)
	binary.LittleEndian.PutUint64(data[16:24], b.id)
	binary.LittleEndian.PutUint64(data[24:32], b.localDepth)
	binary.LittleEndian.PutUint64(data[32:40], uint64(len(b.slots)))

	offset := bucketHeaderSize
	for key, value := range b.slots {
		binary.LittleEndian.PutUint64(data[offset:offset+8], uint64(key))
		binary.LittleEndian.PutUint64(data[offset+8:offset+16], uint64(value))
		offset += bucketEntrySize
	}

	return nil
}

func DeleteBucketFile(baseDir string, id uint64) error {
	err := osRemove(bucketFilePath(baseDir, id))
	if err == nil {
		return nil
	}
	if isNotExist(err) {
		return nil
	}
	return err
}

func osRemove(path string) error {
	return removeFile(path)
}

func isNotExist(err error) bool {
	return fileNotExist(err)
}

func (b *Bucket) ID() uint64 {
	return b.id
}

func (b *Bucket) LocalDepth() uint64 {
	return b.localDepth
}

func (b *Bucket) SetLocalDepth(depth uint64) {
	b.localDepth = depth
}

func (b *Bucket) Count() uint64 {
	return uint64(len(b.slots))
}

func (b *Bucket) Get(key int64) (int64, bool) {
	v, ok := b.slots[key]
	return v, ok
}

func (b *Bucket) Put(key int64, value int64) (inserted bool, updated bool) {
	_, exists := b.slots[key]
	if exists {
		b.slots[key] = value
		return false, true
	}

	b.slots[key] = value
	return true, false
}

func (b *Bucket) Delete(key int64) bool {
	_, exists := b.slots[key]
	if !exists {
		return false
	}

	delete(b.slots, key)
	return true
}

func (b *Bucket) Entries() map[int64]int64 {
	cp := make(map[int64]int64, len(b.slots))
	for k, v := range b.slots {
		cp[k] = v
	}
	return cp
}

func (b *Bucket) ReplaceEntries(entries map[int64]int64) {
	b.slots = make(map[int64]int64, len(entries))
	for k, v := range entries {
		b.slots[k] = v
	}
}

func (b *Bucket) IsFull(limit uint64) bool {
	return uint64(len(b.slots)) >= limit
}

func (b *Bucket) Clear() {
	b.slots = make(map[int64]int64)
}

func (b *Bucket) EntriesCount() int {
	return len(b.slots)
}
