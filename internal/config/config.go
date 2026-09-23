package config

import (
	"encoding/json"
	"os"
)

type MemtableType string

const (
	MemtableTypeHashMap  MemtableType = "hashmap"
	MemtableTypeSkipList MemtableType = "skiplist"
)

type LSMCompactionType string

const (
	LSMCompactionTypeSizeTiered LSMCompactionType = "sizetiered"
	LSMCompactionTypeLeveled    LSMCompactionType = "leveled"
	LSMCompactionTypeHybrid     LSMCompactionType = "hybrid"
)

type MemtableConfig struct {
	Threshold   int          `json:"threshold"`
	InstanceNum int          `json:"instance_num"`
	ImplType    MemtableType `json:"impl_type"`
}
type WALConfig struct {
	SegmentSize int `json:"segment_size"`
}
type CacheConfig struct {
	Capacity int  `json:"capacity"`
	Enabled  bool `json:"enabled"`
}
type LSMConfig struct {
	MaxLevelNum       int               `json:"max_level_num"`
	CompactionType    LSMCompactionType `json:"compaction_type"`
	LevelGrowthFactor int               `json:"level_growth_factor"`
}
type RateLimitConfig struct {
	BucketCapacity int     `json:"bucket_capacity"`
	TokenRate      float64 `json:"token_rate"`
}
type BloomConfig struct {
	FalsePositiveRate float64 `json:"false_positive_rate"`
}
type StorageConfig struct {
	WALPath  string `json:"wal_path"`
	DataPath string `json:"data_path"`
}

type Config struct {
	Memtable  MemtableConfig  `json:"memtable"`
	WAL       WALConfig       `json:"wal"`
	Cache     CacheConfig     `json:"cache"`
	LSM       LSMConfig       `json:"lsm"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	Bloom     BloomConfig     `json:"bloom"`
	Storage   StorageConfig   `json:"storage"`
}

func DefaultConfig() *Config {
	return &Config{
		Memtable: MemtableConfig{
			Threshold:   100,
			InstanceNum: 1,
			ImplType:    MemtableTypeSkipList,
		},
		WAL: WALConfig{
			SegmentSize: 120,
		},
		Cache: CacheConfig{
			Capacity: 10,
			Enabled:  true,
		},
		LSM: LSMConfig{
			MaxLevelNum:       4,
			CompactionType:    LSMCompactionTypeLeveled,
			LevelGrowthFactor: 10,
		},
		RateLimit: RateLimitConfig{
			BucketCapacity: 5,
			TokenRate:      2,
		},
		Bloom: BloomConfig{
			FalsePositiveRate: 0.03,
		},
		Storage: StorageConfig{
			WALPath:  "data/wal",
			DataPath: "data/sstable",
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
