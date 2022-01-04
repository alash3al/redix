package manager

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/alash3al/redix/internals/wal"
	"github.com/docker/go-units"
	"github.com/go-redis/redis/v8"
	"github.com/syndtr/goleveldb/leveldb"
)

// Manager represents a datasource/engines manager
type Manager struct {
	db    contract.Engine
	wal   *wal.Wal
	opts  *Options
	state *leveldb.DB

	minClusterOffsetTimeNS *int64
	minClusterOffsetID     *int64
	maxWalSize             int64

	masterRESPClient *redis.Client

	InternalReplicationHeartBeatChan chan struct{}
}

// New initializes a new manager
func New(opts *Options) (*Manager, error) {
	var err error

	if !contract.Exists(opts.DefaultEngine) {
		return nil, fmt.Errorf("unknown specified driver (%s)", opts.DefaultEngine)
	}

	if opts.InstanceRole != InstanceRoleMaster && opts.InstanceRole != InstanceRoleReplica {
		return nil, fmt.Errorf("unknown instance role (%s) specified", opts.InstanceRole)
	}

	if opts.InstanceRole == InstanceRoleReplica && opts.MasterRESPDSN == "" {
		return nil, fmt.Errorf("empty master specified, please specify a valid master dsn")
	}

	if err = os.MkdirAll(opts.DataDirPath(opts.DefaultEngine), 0755); err != nil {
		return nil, err
	}

	var masterRESPClient *redis.Client

	if opts.InstanceRole == InstanceRoleReplica {
		url, err := redis.ParseURL(opts.MasterRESPDSN)
		if err != nil {
			return nil, fmt.Errorf("invalid master dsn (%s) specified: %s", opts.MasterRESPDSN, err.Error())
		}

		cli := redis.NewClient(url)

		if cli.Ping(context.Background()).Err() != nil {
			return nil, fmt.Errorf("unable to ping the master (%s)", opts.MasterRESPDSN)
		}

		masterRESPClient = cli
	}

	db, err := contract.Open(
		opts.DefaultEngine,
		opts.DataDirPath(opts.DefaultEngine),
	)

	if err != nil {
		return nil, err
	}

	statedb, err := leveldb.OpenFile(opts.DataDirPath("state"), nil)
	if err != nil {
		return nil, err
	}

	clusterMinOffsetTimeNano := int64(0)
	clusterMinOffsetID := int64(0)

	mngr := &Manager{
		opts:                             opts,
		db:                               db,
		state:                            statedb,
		masterRESPClient:                 masterRESPClient,
		minClusterOffsetTimeNS:           &clusterMinOffsetTimeNano,
		minClusterOffsetID:               &clusterMinOffsetID,
		InternalReplicationHeartBeatChan: make(chan struct{}),
	}

	if mngr.opts.InstanceRole == InstanceRoleMaster {
		maxSize, err := units.FromHumanSize(mngr.opts.MaxWalSize)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the WAL_MAX_SIZE due to: %s", err)
		}

		mngr.maxWalSize = maxSize
	}

	if opts.InstanceRole == InstanceRoleMaster {
		waldb, err := wal.Open(opts.DataDirPath("wal"))
		if err != nil {
			return nil, err
		}

		mngr.wal = waldb
	}

	go (func() {
		for {
			select {
			case <-mngr.InternalReplicationHeartBeatChan:
				mngr.replicate()
			}
		}
	})()

	if opts.InstanceRole == InstanceRoleReplica {
		go (func() {
			for {
				time.Sleep(time.Millisecond * 100)

				mngr.InternalReplicationHeartBeatChan <- struct{}{}
			}
		})()

		go (func() {
			for {
				time.Sleep(1 * time.Minute)

				offset, err := mngr.CurrentOffset()
				if err != nil {
					mngr.Report(fmt.Errorf("unable to fetch current instance wal offset due to: %s", err), true)
				}

				if err := mngr.masterRESPClient.Do(context.Background(), "SETCLUSTERWALOFFSET", offset).Err(); err != nil {
					mngr.Report(fmt.Errorf("unable to propogate my offset to the master due to: %s", err), true)
				}
			}
		})()
	}

	go (func() {
		for {
			time.Sleep(time.Minute * 1)

			if opts.InstanceRole == InstanceRoleMaster {
				size, err := mngr.wal.Size()
				if err != nil {
					mngr.Report(err, true)
				}

				if size >= mngr.maxWalSize {
					log.Println("trimming the wal ...")

					if err := mngr.wal.TrimBefore(atomic.LoadInt64(mngr.minClusterOffsetTimeNS), atomic.LoadInt64(mngr.minClusterOffsetID)); err != nil {
						mngr.Report(err, true)
					}

					log.Println("trimmed the wal !")
				}
			}
		}
	})()

	return mngr, nil
}

// UpdateClusterMinimumOffset sync the minimum offset of the cluster to help in trimming the wal
func (m *Manager) UpdateClusterMinimumOffset(offset string) error {
	parts := strings.Split(offset, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid offset specified")
	}

	offsetNS, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("offset parsing failed due to: %s", err)
	}

	offsetID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("offset parsing failed due to: %s", err)
	}

	minClusterOffsetNano := atomic.LoadInt64(m.minClusterOffsetTimeNS)
	minClussterOffsetID := atomic.LoadInt64(m.minClusterOffsetID)

	if minClusterOffsetNano != 0 && minClusterOffsetNano > offsetNS {
		atomic.StoreInt64(m.minClusterOffsetTimeNS, offsetNS)
		if minClussterOffsetID > offsetID {
			atomic.StoreInt64(m.minClusterOffsetID, offsetID)
		}
	}

	return nil
}

// Write writes the specified input
func (m *Manager) Write(input *contract.WriteInput) error {
	if m.opts.InstanceRole == InstanceRoleReplica {
		return fmt.Errorf("unable to perform a write operation in a read-only instance, are you sure that you're connected to the right instance?")
	}

	if bytes.TrimSpace(input.Key) == nil {
		return fmt.Errorf("empty key specified, are cheating?")
	}

	if input.Value == nil {
		input.Delete = true
	}

	rawData, err := input.Marshal()
	if err != nil {
		return err
	}

	if err := m.wal.Write(rawData); err != nil {
		return err
	}

	go (func() {
		m.InternalReplicationHeartBeatChan <- struct{}{}
	})()

	return nil
}

// Get fetches the specified input into the specified dbIndex
func (m *Manager) Get(input *contract.GetInput) (*contract.GetOutput, error) {
	return m.db.Get(input)
}

// ForEach iterate over each key-value in the store using the fn, when fn returns false means break the loop
func (m *Manager) ForEach(fn contract.IteratorFunc) error {
	return m.db.ForEach(fn)
}

// Export dumps the database
func (m *Manager) Export(fn contract.ExportFunc) error {
	return m.db.Export(fn)
}

// Wal returns the wal handler
func (m *Manager) Wal() *wal.Wal {
	if m.opts.InstanceRole != InstanceRoleMaster {
		panic("trying to get a wal instance in a none-master instance/node")
	}
	return m.wal
}

// CurrentOffset returns the current offset of the current instance
func (m *Manager) CurrentOffset() (string, error) {
	currentOffset, err := m.state.Get([]byte("current_offset"), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return "", err
	}

	return string(currentOffset), nil
}

// Report report the specified error
func (m *Manager) Report(err error, shouldPanic bool) {
	if shouldPanic {
		panic(err)
	}

	log.Println(err)
}

func (m *Manager) replicate() {
	currentOffset, err := m.state.Get([]byte("current_offset"), nil)
	if err != nil && err != leveldb.ErrNotFound {
		m.Report(fmt.Errorf("unable to fetch the latest state from state db due to: %s", err.Error()), true)
	}

	if m.opts.InstanceRole == InstanceRoleMaster {
		err := m.wal.Range(func(offset, payload []byte) bool {
			if err := m.directWrite(offset, payload); err != nil {
				m.Report(fmt.Errorf("unable to write to disk due to: %s", err.Error()), true)
				return false
			}

			return true
		}, &wal.RangeOpts{Offset: currentOffset, Limit: 1})

		if err != nil {
			m.Report(fmt.Errorf("unable to read from wal due to: %s", err.Error()), true)
		}

		return
	} else if m.opts.InstanceRole == InstanceRoleReplica {
		if currentOffset == nil {
			m.importFromMaster()
			return
		}

		args := []interface{}{"WAL", 0}
		if len(currentOffset) > 0 {
			args = append(args, currentOffset)
		}

		rows, err := m.masterRESPClient.Do(context.Background(), args...).Slice()
		if err != nil {
			m.Report(fmt.Errorf("unable to fetch data to be replicated from the master due to: %s", err.Error()), true)
		}

		for i := 0; i < len(rows)/2; i += 2 {
			if err := m.directWrite([]byte(rows[i].(string)), []byte(rows[i+1].(string))); err != nil {
				m.Report(fmt.Errorf("unable to apply replicated data due to: %s", err.Error()), true)
			}
		}
	}
}

func (m *Manager) importFromMaster() {
	currentOffset, err := m.state.Get([]byte("current_offset"), nil)
	if err != nil && err != leveldb.ErrNotFound {
		m.Report(fmt.Errorf("unable to fetch the latest state from state db due to: %s", err.Error()), true)
	}

	if currentOffset == nil && m.opts.InstanceRole != InstanceRoleMaster {
		log.Println("=> asking for a dump from the master instance ...")

		resp, err := http.Get(fmt.Sprintf("%s/dump", m.opts.MasterHTTPBaseURL))
		if err != nil {
			m.Report(fmt.Errorf("unable to fetch the dump from the master due to: %s", err.Error()), true)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			bdy, _ := ioutil.ReadAll(resp.Body)
			m.Report(fmt.Errorf("unable to fetch the dump from the master due to: %s", string(bdy)), true)
		}

		currentOffset := resp.Header.Get("X-Current-Offset")
		written, err := m.db.Import(resp.Body)
		if err != nil {
			m.Report(fmt.Errorf("unable to import from the master dump due to: %s", err.Error()), true)
		}

		if written != resp.ContentLength {
			m.Report(fmt.Errorf("the dump content-length isn't the same as written dump, something went wrong there, could you retry again?"), true)
		}

		if err := m.state.Put([]byte("current_offset"), []byte(currentOffset), nil); err != nil {
			resp.Body.Close()
			m.Report(fmt.Errorf("unable to fetch the dump from the master due to: %s", err.Error()), true)
		}

		log.Println("=> finalized the dump ...")
	}
}

func (m *Manager) directWrite(offset, payload []byte) error {
	input, err := contract.UnmarshalWriteInput(payload)
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)
	batch.Put([]byte("current_offset"), offset)

	if err := m.state.Write(batch, nil); err != nil {
		return fmt.Errorf("state err: %s", err)
	}

	if _, err := m.db.Write(input); err != nil {
		return fmt.Errorf("db error: %s", err)
	}

	return nil
}
