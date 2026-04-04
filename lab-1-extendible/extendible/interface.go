package extendible

type MetaInterface interface {
	Index(hash uint64) uint64
	GetBucketID(idx uint64) uint64
	SetBucketID(idx uint64, bucketID uint64)
	DoubleDirectory()
	DirectorySize() uint64
	RepointAfterSplit(oldBucketID, newBucketID uint64, oldLocalDepth uint64)
	Save(path string) error
}

type BucketInterface interface {
	ID() uint64
	LocalDepth() uint64
	SetLocalDepth(depth uint64)
	Count() uint64
	Get(key int64) (int64, bool)
	Put(key int64, value int64) (inserted bool, updated bool)
	Delete(key int64) bool
	Entries() map[int64]int64
}
