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

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"sync"
	"unsafe"

	"github.com/intel/idxd"
	log "github.com/sirupsen/logrus"
)

// Record A tuple with an address
type Record struct {
	offset uint64
}

// Trace Contains records
type Trace struct {
	sync.Mutex
	traceFileName string

	containedOffsets map[uint64]int
	trace            []Record
	regions          map[uint64]int
}

func initTrace(traceFileName string) *Trace {
	t := new(Trace)

	t.traceFileName = traceFileName
	t.regions = make(map[uint64]int)
	t.containedOffsets = make(map[uint64]int)
	t.trace = make([]Record, 0)

	return t
}

// AppendRecord Appends a record to the trace
func (t *Trace) AppendRecord(r Record) {
	t.Lock()
	defer t.Unlock()

	t.trace = append(t.trace, r)
	t.containedOffsets[r.offset] = 0
}

// WriteTrace Writes all the records to a file
func (t *Trace) WriteTrace() {
	t.Lock()
	defer t.Unlock()

	file, err := os.Create(t.traceFileName)
	if err != nil {
		log.Fatalf("Failed to open trace file for writing: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, rec := range t.trace {
		err := writer.Write([]string{
			strconv.FormatUint(rec.offset, 16)})
		if err != nil {
			log.Fatalf("Failed to write trace: %v", err)
		}
	}
}

// readTrace Reads all the records from a CSV file
//nolint:deadcode,unused
func (t *Trace) readTrace() {
	f, err := os.Open(t.traceFileName)
	if err != nil {
		log.Fatalf("Failed to open trace file for reading: %v", err)
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("Failed to read from the trace file: %v", err)
	}

	for _, line := range lines {
		rec := readRecord(line)
		t.AppendRecord(rec)
	}
}

// readRecord Parses a record from a line
//nolint:deadcode,unused
func readRecord(line []string) Record {
	offset, err := strconv.ParseUint(line[0], 16, 64)
	if err != nil {
		log.Fatalf("Failed to convert string to offset: %v", err)
	}

	rec := Record{
		offset: offset,
	}
	return rec
}

// Search trace for the record with the same offset
func (t *Trace) containsRecord(rec Record) bool {
	_, ok := t.containedOffsets[rec.offset]

	return ok
}

// ProcessRecord Prepares the trace, the regions map, and the working set file for replay
// Must be called when record is done (i.e., it is not concurrency-safe vs. AppendRecord)
func (s *SnapshotState) ProcessRecord(GuestMemPath, WorkingSetPath string) {
	log.Debug("Preparing replay structures")

	t := s.trace
	// sort trace records in the ascending order by offset
	sort.Slice(t.trace, func(i, j int) bool {
		return t.trace[i].offset < t.trace[j].offset
	})

	// build the map of contiguous regions from the trace records
	var last, regionStart uint64
	for _, rec := range t.trace {
		if rec.offset != last+uint64(os.Getpagesize()) {
			regionStart = rec.offset
			t.regions[regionStart] = 1
		} else {
			t.regions[regionStart]++
		}

		last = rec.offset
	}

	s.writeWorkingSetPagesToFile(GuestMemPath, WorkingSetPath)
}

func (s *SnapshotState) writeWorkingSetPagesToFile(guestMemFileName, WorkingSetPath string) {
	log.Debug("Writing the working set pages to a disk")

	t := s.trace
	fSrc, err := os.Open(guestMemFileName)
	if err != nil {
		log.Fatalf("Failed to open guest memory file for reading")
	}
	defer fSrc.Close()
	fDst, err := os.Create(WorkingSetPath)
	if err != nil {
		log.Fatalf("Failed to open ws file for writing")
	}
	defer fDst.Close()

	var (
		dstOffset int64
		count     int
	)

	size := len(t.trace) * os.Getpagesize()
	s.workingSet_InMem = AlignedBlock(size) // direct io requires aligned buffer

	// Form a sorted slice of keys to access the map in a predetermined order
	keys := make([]uint64, 0)
	for k := range t.regions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	elem := len(t.regions)
	buf_slice := make([][]byte, elem)
	total_size := 0
	desc_mem := idxd.AlignedBlock(64*elem, 64)
	comp_mem := idxd.AlignedBlock(32*elem, 32)

	log.Infof("Starting to save working set pages")

	for idx, offset := range keys {
		regLength := t.regions[offset]
		copyLen := regLength * os.Getpagesize()

		buf_slice[idx] = make([]byte, copyLen)

		if n, err := fSrc.ReadAt(buf_slice[idx], int64(offset)); n != copyLen || err != nil {
			log.Fatalf("Read file failed for src")
		}

		if !s.InMemWorkingSet {
			if n, err := fDst.WriteAt(buf_slice[idx], dstOffset); n != copyLen || err != nil {
				log.Fatalf("Write file failed for dst")
			} else {
				log.Debug("Copied %d bytes from buf_slice[idx] to file", n)
			}
		} else {
			if s.UseDSA {
				desc := (*idxd.DSA_hw_desc_go)(unsafe.Pointer(&desc_mem[64*idx]))
				comp := (*idxd.DSA_completion_record_go)(unsafe.Pointer(&comp_mem[32*idx]))

				desc.Src_addr = (*uint64)(unsafe.Pointer(&buf_slice[idx][0]))
				desc.Dst_addr = (*uint64)(unsafe.Pointer(&s.workingSet_InMem[dstOffset:][0]))
				desc.Xfer_size = uint32(copyLen)
				desc.Opcode = idxd.DSA_OPCODE_MEMMOVE
				desc.Completion_addr = (*uint64)(unsafe.Pointer(comp))
				desc.Flags[0] = idxd.IDXD_FLAG_BLOCK_ON_FAULT | idxd.IDXD_FLAG_CRAV | idxd.IDXD_FLAG_RCR

				// idxd.DSA_memmove_sync_go(s.workingSet_InMem[dstOffset:], buf_slice[idx], uint32(copyLen))
				idxd.DSA_memmove_desc_go(desc, 0)
				log.Debug("Copied %d bytes from buf to im mem working set using DSA", copyLen)
			} else {
				nb_bytes := copy(s.workingSet_InMem[dstOffset:], buf_slice[idx])
				log.Debug("Copied %d bytes from buf to im mem working set using CPU", nb_bytes)
			}
			total_size += copyLen
		}

		dstOffset += int64(copyLen)
		total_size += copyLen
		count += regLength
	}

	if !s.InMemWorkingSet {
		if err := fDst.Sync(); err != nil {
			log.Fatalf("Sync file failed for dst")
		}
		log.Infof("Copied %d bytes from buf to working set file", total_size)
	} else {
		if s.UseDSA {
			for idx := 0; idx < elem; idx++ {
				desc := (*idxd.DSA_hw_desc_go)(unsafe.Pointer(&desc_mem[64*idx]))
				idxd.DSA_wait_for_comp_go(desc)
				comp := (*idxd.DSA_completion_record_go)(unsafe.Pointer(desc.Completion_addr))
				if comp.Status != 1 {
					log.Warnf("DSA Copy failed with stattus = 0x%x", comp.Status)
				}
			}
			log.Infof("Copied %d bytes from buf to im mem working set using DSA", total_size)
		} else {
			log.Infof("Copied %d bytes from buf to im mem working set using CPU", total_size)
		}
	}
	log.Infof("Done with save working set pages")
}
