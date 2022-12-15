// MIT License
//
// Copyright (c) 2020 Dmitrii Ustiugov, Plamen Petrov and EASE lab
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package manager

/*
#cgo LDFLAGS: -lnuma
#include "user_page_faults.h"
#include "cxl_mem.h"
#include <numa.h>
*/
import "C"

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/ftrvxmtrx/fd"
	"github.com/intel/idxd"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/ease-lab/vhive/metrics"

	"unsafe"
)

// SnapshotStateCfg Config to initialize SnapshotState
type SnapshotStateCfg struct {
	VMID string

	VMMStatePath, GuestMemPath, WorkingSetPath string

	InstanceSockAddr string
	BaseDir          string // base directory for the instance
	MetricsPath      string // path to csv file where the metrics should be stored
	IsLazyMode       bool
	GuestMemSize     int
	metricsModeOn    bool
	InMemWorkingSet  bool
	InCxlMem         bool
	UseDSA           bool
	MovePages        bool
}

// SnapshotState Stores the state of the snapshot
// of the VM.
type SnapshotState struct {
	SnapshotStateCfg
	firstPageFaultOnce *sync.Once // to initialize the start virtual address and replay
	startAddress       uint64
	userFaultFD        *os.File
	trace              *Trace
	epfd               int
	quitCh             chan int

	// to indicate whether the instance has even been activated. this is to
	// get around cases where offload is called for the first time
	isEverActivated bool
	// for sanity checking on deactivate/activate
	isActive bool

	isRecordReady bool

	guestMem         []byte
	workingSet       []byte
	workingSet_InMem []byte

	// Stats
	totalPFServed  []float64
	uniquePFServed []float64
	reusedPFServed []float64
	latencyMetrics []*metrics.Metric

	replayedNum   int // only valid for lazy serving
	uniqueNum     int
	currentMetric *metrics.Metric

	isGuestMemMapped bool
}

// NewSnapshotState Initializes a snapshot state
func NewSnapshotState(cfg SnapshotStateCfg) *SnapshotState {
	s := new(SnapshotState)
	s.SnapshotStateCfg = cfg

	s.trace = initTrace(s.getTraceFile())
	if s.metricsModeOn {
		s.totalPFServed = make([]float64, 0)
		s.uniquePFServed = make([]float64, 0)
		s.reusedPFServed = make([]float64, 0)
		s.latencyMetrics = make([]*metrics.Metric, 0)
	}

	s.isGuestMemMapped = false

	return s
}

func (s *SnapshotState) setupStateOnActivate() {
	s.isActive = true
	s.isEverActivated = true
	s.firstPageFaultOnce = new(sync.Once)
	s.quitCh = make(chan int)

	if s.metricsModeOn {
		s.uniqueNum = 0
		s.replayedNum = 0
		s.currentMetric = metrics.NewMetric()
	}
}

func (s *SnapshotState) getUFFD() error {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		c, err := d.DialContext(ctx, "unix", s.InstanceSockAddr)
		if err != nil {
			if ctx.Err() != nil {
				log.Error("Failed to dial within the context timeout")
				return err
			}
			time.Sleep(1 * time.Millisecond)
			continue
		}

		defer c.Close()

		sendfdConn := c.(*net.UnixConn)

		fs, err := fd.Get(sendfdConn, 1, []string{"a file"})
		if err != nil {
			log.Error("Failed to receive the uffd")
			return err
		}

		s.userFaultFD = fs[0]

		return nil
	}
}

func (s *SnapshotState) processMetrics() {
	if s.metricsModeOn && s.isRecordReady {
		s.uniquePFServed = append(s.uniquePFServed, float64(s.uniqueNum))

		if s.IsLazyMode {
			s.totalPFServed = append(s.totalPFServed, float64(s.replayedNum))
			s.reusedPFServed = append(
				s.reusedPFServed,
				float64(s.replayedNum-s.uniqueNum),
			)
		}

		s.latencyMetrics = append(s.latencyMetrics, s.currentMetric)
	}
}

func (s *SnapshotState) getTraceFile() string {
	return filepath.Join(s.BaseDir, "trace")
}

func (s *SnapshotState) mapGuestMemory() error {
	if s.isGuestMemMapped {
		return nil
	}

	fd, err := os.OpenFile(s.GuestMemPath, os.O_RDONLY, 0444)
	if err != nil {
		log.Errorf("Failed to open guest memory file: %v", err)
		return err
	}

	s.guestMem, err = unix.Mmap(int(fd.Fd()), 0, s.GuestMemSize, unix.PROT_READ, unix.MAP_PRIVATE)
	if err != nil {
		log.Errorf("Failed to mmap guest memory file: %v", err)
		return err
	}

	if s.InMemWorkingSet || s.MovePages {
		s.isGuestMemMapped = true
	}

	return nil
}

func (s *SnapshotState) unmapGuestMemory() error {
	if s.InMemWorkingSet || s.MovePages {
		return nil
	}

	if err := unix.Munmap(s.guestMem); err != nil {
		log.Errorf("Failed to munmap guest memory file: %v", err)
		return err
	}

	return nil
}

// alignment returns alignment of the block in memory
// with reference to alignSize
//
// Can't check alignment of a zero sized block as &block[0] is invalid
func alignment(block []byte, alignSize int) int {
	return int(uintptr(unsafe.Pointer(&block[0])) & uintptr(alignSize-1))
}

// AlignedBlock returns []byte of size BlockSize aligned to a multiple
// of alignSize in memory (must be power of two)
func AlignedBlock(blockSize int) []byte {
	alignSize := os.Getpagesize() // must be multiple of the filesystem block size

	if blockSize == 0 {
		return nil
	}

	block := make([]byte, blockSize+alignSize)

	a := alignment(block, alignSize)
	offset := 0
	if a != 0 {
		offset = alignSize - a
	}
	block = block[offset : offset+blockSize]

	// Check
	if blockSize != 0 {
		a = alignment(block, alignSize)
		if a != 0 {
			log.Fatal("Failed to align block")
		}
	}
	return block
}

type NoCxlNode struct{}

func (m *NoCxlNode) Error() string {
	return "No CXL memory node found"
}

type CxlMemAllocErr struct{}

func (m *CxlMemAllocErr) Error() string {
	return "CXL memory could not be allocated"
}

// returns the cxl node where memory will be allocated
//it is suppose to return cxl, currently it is returning
//any node other than current process
func fetchMemoryNode() (int, error) {
	totalNodes := C.get_total_nodes()
	curNode := C.get_cur_node()
	n := totalNodes - 1
	nodes := totalNodes
	for nodes > 0 {
		if n != curNode {
			return int(n), nil
			//C.numa_free(p,1024)
		}
		nodes--
		n = (n + 1) % totalNodes
	}
	return -1, &NoCxlNode{}

}

// AlignedBlock returns []byte of size BlockSize aligned to a multiple
// of blockSize in CXL memory (must be power of two)
func AlignedCxlBlock(blockSize int) ([]byte, error) {
	alignSize := os.Getpagesize() // must be multiple of the filesystem block size

	if blockSize == 0 {
		return nil, nil
	}
	if blockSize%alignSize > 0 {
		blockSize = ((blockSize / alignSize) + 1) * alignSize
	}

	node, err := fetchMemoryNode()
	if err != nil {
		return nil, err
	}
	p := C.numa_alloc_onnode(C.ulong(blockSize), C.int(node))
	C.setval(p) //hack to force allocate the memory
	if p == nil {
		log.Warnf("Error numa memory allocation")
		return nil, &CxlMemAllocErr{}
	}
	block := unsafe.Slice((*byte)(unsafe.Pointer(p)), blockSize)
	return block, nil
}

//numa allocated node must be freed
func FreeCxlMem(block []byte, blockSize int) {
	C.numa_free(unsafe.Pointer(&block[0]), C.ulong(blockSize))
}

// fetchState Fetches the working set file (or the whole guest memory) and the VMM state file
func (s *SnapshotState) fetchState() error {
	if s.MovePages || (s.InMemWorkingSet && !s.InCxlMem) {
		log.Infof("Nothing to fetch for move pages or inMem working set")
		return nil
	}
	size := len(s.trace.trace) * os.Getpagesize()

	log.Infof("Starting to fetch working set")

	if !s.InMemWorkingSet {
		if _, err := ioutil.ReadFile(s.VMMStatePath); err != nil {
			log.Errorf("Failed to fetch VMM state: %v\n", err)
			return err
		}

		// O_DIRECT allows to fully leverage disk bandwidth by bypassing the OS page cache
		f, err := os.OpenFile(s.WorkingSetPath, os.O_RDONLY|syscall.O_DIRECT, 0600)
		if err != nil {
			log.Errorf("Failed to open the working set file for direct-io: %v\n", err)
			return err
		}

		s.workingSet = AlignedBlock(size) // direct io requires aligned buffer
		if n, err := f.Read(s.workingSet); n != size || err != nil {
			log.Errorf("Reading working set file failed: %v\n", err)
			return err
		} else {
			log.Infof("Copied %d bytes from file to working set", n)
		}
		if err := f.Close(); err != nil {
			log.Errorf("Failed to close the working set file: %v\n", err)
			return err
		}
	} else {
		if s.UseDSA {
			res := idxd.DSA_memmove_sync_go(s.workingSet, s.workingSet_InMem, uint32(size))
			if res != 0 {
				log.Warnf("DSA Copy failed with status = 0x%x for size = %d", res, size)
			} else {
				log.Infof("Copied %d bytes from in mem to working set using DSA", size)
			}
		} else {
			nb_bytes := copy(s.workingSet, s.workingSet_InMem)
			log.Infof("Copied %d bytes from in mem to working set using CPU", nb_bytes)
		}
	}

	return nil
}

func (s *SnapshotState) pollUserPageFaults(readyCh chan int) {
	logger := log.WithFields(log.Fields{"vmID": s.VMID})

	var events [1]syscall.EpollEvent

	if err := s.registerEpoller(); err != nil {
		logger.Fatalf("register_epoller: %v", err)
	}

	logger.Debug("Starting polling loop")

	defer syscall.Close(s.epfd)

	readyCh <- 0

	for {
		select {
		case <-s.quitCh:
			logger.Debug("Handler received a signal to quit")
			return
		default:
			nevents, err := syscall.EpollWait(s.epfd, events[:], -1)
			if err != nil {
				logger.Fatalf("epoll_wait: %v", err)
				break
			}

			if nevents < 1 {
				panic("Wrong number of events")
			}

			for i := 0; i < nevents; i++ {
				event := events[i]

				fd := int(event.Fd)

				stateFd := int(s.userFaultFD.Fd())

				if fd != stateFd && stateFd != -1 {
					logger.Fatalf("Received event from unknown fd")
				}

				goMsg := make([]byte, sizeOfUFFDMsg())

				if nread, err := syscall.Read(fd, goMsg); err != nil || nread != len(goMsg) {
					if !errors.Is(err, syscall.EBADF) {
						log.Fatalf("Read uffd_msg failed: %v", err)
					}
					break
				}

				if event := uint8(goMsg[0]); event != uffdPageFault() {
					log.Fatal("Received wrong event type")
				}

				address := binary.LittleEndian.Uint64(goMsg[16:])

				if err := s.servePageFault(fd, address); err != nil {
					log.Fatalf("Failed to serve page fault")
				}
			}
		}
	}
}

func (s *SnapshotState) registerEpoller() error {
	logger := log.WithFields(log.Fields{"vmID": s.VMID})

	var (
		err   error
		event syscall.EpollEvent
		fdInt int
	)

	fdInt = int(s.userFaultFD.Fd())

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fdInt)

	s.epfd, err = syscall.EpollCreate1(0)
	if err != nil {
		logger.Errorf("Failed to create epoller %v", err)
		return err
	}

	if err := syscall.EpollCtl(
		s.epfd,
		syscall.EPOLL_CTL_ADD,
		fdInt,
		&event,
	); err != nil {
		logger.Errorf("Failed to subscribe VM %v", err)
		return err
	}

	return nil
}

func (s *SnapshotState) servePageFault(fd int, address uint64) error {
	var (
		tStart              time.Time
		workingSetInstalled bool
	)

	s.firstPageFaultOnce.Do(
		func() {
			s.startAddress = address

			if s.isRecordReady && !s.IsLazyMode {
				if s.metricsModeOn {
					tStart = time.Now()
				}
				s.installWorkingSetPages(fd)
				if s.metricsModeOn {
					s.currentMetric.MetricMap[installWSMetric] = metrics.ToUS(time.Since(tStart))
				}

				workingSetInstalled = true
			}
		})

	if workingSetInstalled {
		return nil
	}

	offset := address - s.startAddress

	src := uint64(uintptr(unsafe.Pointer(&s.guestMem[offset])))
	dst := uint64(int64(address) & ^(int64(os.Getpagesize()) - 1))
	mode := uint64(0)

	rec := Record{
		offset: offset,
	}

	if !s.isRecordReady {
		s.trace.AppendRecord(rec)
	} else {
		log.Debug("Serving a page that is missing from the working set")
	}

	if s.metricsModeOn {
		if s.isRecordReady {
			if s.IsLazyMode {
				if !s.trace.containsRecord(rec) {
					s.uniqueNum++
				}
				s.replayedNum++
			} else {
				s.uniqueNum++
			}

		}

		tStart = time.Now()
	}

	err := installRegion(fd, src, dst, mode, 1)

	if s.metricsModeOn {
		s.currentMetric.MetricMap[serveUniqueMetric] += metrics.ToUS(time.Since(tStart))
	}

	return err
}

func (s *SnapshotState) installWorkingSetPages(fd int) {
	log.Debug("Installing the working set pages")
	defer log.Debug("Done installing the working set pages")

	defer wake(fd, s.startAddress, os.Getpagesize())

	if s.MovePages {
		s.MovePagesToNumaNode(0)
		return
	}

	// build a list of sorted regions
	keys := make([]uint64, 0)
	for k := range s.trace.regions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var (
		srcOffset uint64
		src       uint64
	)

	for _, offset := range keys {
		regLength := s.trace.regions[offset]
		regAddress := s.startAddress + offset
		mode := uint64(C.const_UFFDIO_COPY_MODE_DONTWAKE)
		if s.InMemWorkingSet {
			src = uint64(uintptr(unsafe.Pointer(&s.workingSet_InMem[srcOffset])))
		} else {
			src = uint64(uintptr(unsafe.Pointer(&s.workingSet[srcOffset])))
		}
		dst := regAddress

		if err := installRegion(fd, src, dst, mode, uint64(regLength)); err != nil {
			log.Fatalf("install_region: %v", err)
		}

		srcOffset += uint64(regLength) * 4096
	}
}

func installRegion(fd int, src, dst, mode, len uint64) error {
	cUC := C.struct_uffdio_copy{
		mode: C.ulonglong(mode),
		copy: 0,
		src:  C.ulonglong(src),
		dst:  C.ulonglong(dst),
		len:  C.ulonglong(uint64(os.Getpagesize()) * len),
	}

	err := ioctl(uintptr(fd), int(C.const_UFFDIO_COPY), unsafe.Pointer(&cUC))
	if err != nil {
		return err
	}

	return nil
}

func ioctl(fd uintptr, request int, argp unsafe.Pointer) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		fd,
		uintptr(request),
		// Note that the conversion from unsafe.Pointer to uintptr _must_
		// occur in the call expression.  See the package unsafe documentation
		// for more details.
		uintptr(argp),
	)
	if errno != 0 {
		return os.NewSyscallError("ioctl", fmt.Errorf("%d", int(errno)))
	}

	return nil
}

func wake(fd int, startAddress uint64, len int) {
	cUR := C.struct_uffdio_range{
		start: C.ulonglong(startAddress),
		len:   C.ulonglong(len),
	}

	err := ioctl(uintptr(fd), int(C.const_UFFDIO_WAKE), unsafe.Pointer(&cUR))
	if err != nil {
		log.Fatalf("ioctl failed: %v", err)
	}
}

//nolint:deadcode,unused
func registerForUpf(startAddress []byte, len uint64) int {
	return int(C.register_for_upf(unsafe.Pointer(&startAddress[0]), C.ulong(len)))
}

func sizeOfUFFDMsg() int {
	return C.sizeof_struct_uffd_msg
}

func uffdPageFault() uint8 {
	return uint8(C.const_UFFD_EVENT_PAGEFAULT)
}
