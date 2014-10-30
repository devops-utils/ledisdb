package config

import (
	"bytes"
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/siddontang/go/ioutil2"
	"io"
	"io/ioutil"
)

var (
	ErrNoConfigFile = errors.New("Running without a config file")
)

const (
	DefaultAddr string = "127.0.0.1:6380"

	DefaultDBName string = "goleveldb"

	DefaultDataDir string = "./var"

	KB int = 1024
	MB int = KB * 1024
	GB int = MB * 1024
)

type LevelDBConfig struct {
	Compression     bool `toml:"compression"`
	BlockSize       int  `toml:"block_size"`
	WriteBufferSize int  `toml:"write_buffer_size"`
	CacheSize       int  `toml:"cache_size"`
	MaxOpenFiles    int  `toml:"max_open_files"`
}

type RocksDBConfig struct {
	Compression                    int  `toml:"compression"`
	BlockSize                      int  `toml:"block_size"`
	WriteBufferSize                int  `toml:"write_buffer_size"`
	CacheSize                      int  `toml:"cache_size"`
	MaxOpenFiles                   int  `toml:"max_open_files"`
	MaxWriteBufferNum              int  `toml:"max_write_buffer_num"`
	MinWriteBufferNumberToMerge    int  `toml:"min_write_buffer_number_to_merge"`
	NumLevels                      int  `toml:"num_levels"`
	Level0FileNumCompactionTrigger int  `toml:"level0_file_num_compaction_trigger"`
	Level0SlowdownWritesTrigger    int  `toml:"level0_slowdown_writes_trigger"`
	Level0StopWritesTrigger        int  `toml:"level0_stop_writes_trigger"`
	TargetFileSizeBase             int  `toml:"target_file_size_base"`
	TargetFileSizeMultiplier       int  `toml:"target_file_size_multiplier"`
	MaxBytesForLevelBase           int  `toml:"max_bytes_for_level_base"`
	MaxBytesForLevelMultiplier     int  `toml:"max_bytes_for_level_multiplier"`
	DisableAutoCompactions         bool `toml:"disable_auto_compactions"`
	DisableDataSync                bool `toml:"disable_data_sync"`
	UseFsync                       bool `toml:"use_fsync"`
	MaxBackgroundCompactions       int  `toml:"max_background_compactions"`
	MaxBackgroundFlushes           int  `toml:"max_background_flushes"`
	AllowOsBuffer                  bool `toml:"allow_os_buffer"`
	EnableStatistics               bool `toml:"enable_statistics"`
	StatsDumpPeriodSec             int  `toml:"stats_dump_period_sec"`
	BackgroundThreads              int  `toml:"background_theads"`
	HighPriorityBackgroundThreads  int  `toml:"high_priority_background_threads"`
	DisableWAL                     bool `toml:"disable_wal"`
}

type LMDBConfig struct {
	MapSize int  `toml:"map_size"`
	NoSync  bool `toml:"nosync"`
}

type ReplicationConfig struct {
	Path             string `toml:"path"`
	Sync             bool   `toml:"sync"`
	WaitSyncTime     int    `toml:"wait_sync_time"`
	WaitMaxSlaveAcks int    `toml:"wait_max_slave_acks"`
	ExpiredLogDays   int    `toml:"expired_log_days"`
	SyncLog          int    `toml:"sync_log"`
	Compression      bool   `toml:"compression"`
}

type SnapshotConfig struct {
	Path   string `toml:"path"`
	MaxNum int    `toml:"max_num"`
}

type Config struct {
	FileName string `toml:"-"`

	Addr string `toml:"addr"`

	HttpAddr string `toml:"http_addr"`

	SlaveOf string `toml:"slaveof"`

	Readonly bool `toml:readonly`

	DataDir string `toml:"data_dir"`

	DBName       string `toml:"db_name"`
	DBPath       string `toml:"db_path"`
	DBSyncCommit int    `toml:"db_sync_commit"`

	LevelDB LevelDBConfig `toml:"leveldb"`
	RocksDB RocksDBConfig `toml:"rocksdb"`

	LMDB LMDBConfig `toml:"lmdb"`

	AccessLog string `toml:"access_log"`

	UseReplication bool              `toml:"use_replication"`
	Replication    ReplicationConfig `toml:"replication"`

	Snapshot SnapshotConfig `toml:"snapshot"`

	ConnReadBufferSize    int `toml:"conn_read_buffer_size"`
	ConnWriteBufferSize   int `toml:"conn_write_buffer_size"`
	ConnKeepaliveInterval int `toml:"conn_keepavlie_interval"`

	TTLCheckInterval int `toml:"ttl_check_interval"`
}

func NewConfigWithFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if cfg, err := NewConfigWithData(data); err != nil {
		return nil, err
	} else {
		cfg.FileName = fileName
		return cfg, nil
	}
}

func NewConfigWithData(data []byte) (*Config, error) {
	cfg := NewConfigDefault()

	_, err := toml.Decode(string(data), cfg)
	if err != nil {
		return nil, err
	}

	cfg.adjust()

	return cfg, nil
}

func NewConfigDefault() *Config {
	cfg := new(Config)

	cfg.Addr = DefaultAddr
	cfg.HttpAddr = ""

	cfg.DataDir = DefaultDataDir

	cfg.DBName = DefaultDBName

	cfg.SlaveOf = ""
	cfg.Readonly = false

	// disable access log
	cfg.AccessLog = ""

	cfg.LMDB.MapSize = 20 * MB
	cfg.LMDB.NoSync = true

	cfg.UseReplication = false
	cfg.Replication.WaitSyncTime = 500
	cfg.Replication.Compression = true
	cfg.Replication.WaitMaxSlaveAcks = 2
	cfg.Replication.SyncLog = 0
	cfg.Snapshot.MaxNum = 1

	cfg.RocksDB.AllowOsBuffer = true
	cfg.RocksDB.EnableStatistics = false
	cfg.RocksDB.UseFsync = false
	cfg.RocksDB.DisableAutoCompactions = false
	cfg.RocksDB.AllowOsBuffer = true
	cfg.RocksDB.DisableWAL = false

	cfg.adjust()

	return cfg
}

func getDefault(d int, s int) int {
	if s <= 0 {
		return d
	} else {
		return s
	}
}

func (cfg *Config) adjust() {
	cfg.LevelDB.adjust()

	cfg.RocksDB.adjust()

	cfg.Replication.ExpiredLogDays = getDefault(7, cfg.Replication.ExpiredLogDays)
	cfg.ConnReadBufferSize = getDefault(4*KB, cfg.ConnReadBufferSize)
	cfg.ConnWriteBufferSize = getDefault(4*KB, cfg.ConnWriteBufferSize)
	cfg.TTLCheckInterval = getDefault(1, cfg.TTLCheckInterval)
}

func (cfg *LevelDBConfig) adjust() {
	cfg.CacheSize = getDefault(4*MB, cfg.CacheSize)
	cfg.BlockSize = getDefault(4*KB, cfg.BlockSize)
	cfg.WriteBufferSize = getDefault(4*MB, cfg.WriteBufferSize)
	cfg.MaxOpenFiles = getDefault(1024, cfg.MaxOpenFiles)
}

func (cfg *RocksDBConfig) adjust() {
	cfg.CacheSize = getDefault(4*MB, cfg.CacheSize)
	cfg.BlockSize = getDefault(4*KB, cfg.BlockSize)
	cfg.WriteBufferSize = getDefault(4*MB, cfg.WriteBufferSize)
	cfg.MaxOpenFiles = getDefault(1024, cfg.MaxOpenFiles)
	cfg.MaxWriteBufferNum = getDefault(2, cfg.MaxWriteBufferNum)
	cfg.MinWriteBufferNumberToMerge = getDefault(1, cfg.MinWriteBufferNumberToMerge)
	cfg.NumLevels = getDefault(7, cfg.NumLevels)
	cfg.Level0FileNumCompactionTrigger = getDefault(4, cfg.Level0FileNumCompactionTrigger)
	cfg.Level0SlowdownWritesTrigger = getDefault(16, cfg.Level0SlowdownWritesTrigger)
	cfg.Level0StopWritesTrigger = getDefault(64, cfg.Level0StopWritesTrigger)
	cfg.TargetFileSizeBase = getDefault(32*MB, cfg.TargetFileSizeBase)
	cfg.TargetFileSizeMultiplier = getDefault(1, cfg.TargetFileSizeMultiplier)
	cfg.MaxBytesForLevelBase = getDefault(32*MB, cfg.MaxBytesForLevelBase)
	cfg.MaxBytesForLevelMultiplier = getDefault(1, cfg.MaxBytesForLevelMultiplier)
	cfg.MaxBackgroundCompactions = getDefault(1, cfg.MaxBackgroundCompactions)
	cfg.MaxBackgroundFlushes = getDefault(1, cfg.MaxBackgroundFlushes)
	cfg.StatsDumpPeriodSec = getDefault(3600, cfg.StatsDumpPeriodSec)
	cfg.BackgroundThreads = getDefault(2, cfg.BackgroundThreads)
	cfg.HighPriorityBackgroundThreads = getDefault(1, cfg.HighPriorityBackgroundThreads)
}

func (cfg *Config) Dump(w io.Writer) error {
	e := toml.NewEncoder(w)
	e.Indent = ""
	return e.Encode(cfg)
}

func (cfg *Config) DumpFile(fileName string) error {
	var b bytes.Buffer

	if err := cfg.Dump(&b); err != nil {
		return err
	}

	return ioutil2.WriteFileAtomic(fileName, b.Bytes(), 0644)
}

func (cfg *Config) Rewrite() error {
	if len(cfg.FileName) == 0 {
		return ErrNoConfigFile
	}

	return cfg.DumpFile(cfg.FileName)
}
