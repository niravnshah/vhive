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
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ease-lab/vhive/idxd"
	"github.com/ease-lab/vhive/metrics"
	"gonum.org/v1/gonum/stat"

	log "github.com/sirupsen/logrus"
)

const (
	serveUniqueMetric = "ServeUnique"
	installWSMetric   = "InstallWS"
	fetchStateMetric  = "FetchState"
)

// MemoryManagerCfg Global config of the manager
type MemoryManagerCfg struct {
	MetricsModeOn   bool
	InMemWorkingSet bool
	UseDSA          bool
}

// MemoryManager Serves page faults coming from VMs
type MemoryManager struct {
	sync.Mutex
	MemoryManagerCfg
	instances map[string]*SnapshotState // Indexed by vmID
}

// NewMemoryManager Initializes a new memory manager
func NewMemoryManager(cfg MemoryManagerCfg) *MemoryManager {
	log.Debug("Initializing the memory manager")

	m := new(MemoryManager)
	m.instances = make(map[string]*SnapshotState)
	m.MemoryManagerCfg = cfg

	return m
}

// RegisterVM Registers a VM within the memory manager
func (m *MemoryManager) RegisterVM(cfg SnapshotStateCfg) error {
	m.Lock()
	defer m.Unlock()

	vmID := cfg.VMID

	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Registering the VM with the memory manager")

	if _, ok := m.instances[vmID]; ok {
		logger.Error("VM already registered with the memory manager")
		return errors.New("VM already registered with the memory manager")
	}

	cfg.metricsModeOn = m.MetricsModeOn
	cfg.InMemWorkingSet = m.InMemWorkingSet
	cfg.UseDSA = m.UseDSA
	if m.UseDSA {
		idxd.DSA_setup_c("/dev/dsa/wq0.0")
	}

	state := NewSnapshotState(cfg)

	m.instances[vmID] = state

	return nil
}

// DeregisterVM Deregisters a VM from the memory manager
func (m *MemoryManager) DeregisterVM(vmID string) error {
	m.Lock()
	defer m.Unlock()

	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Deregistering VM from the memory manager")

	state, ok := m.instances[vmID]
	if !ok {
		logger.Error("VM is not registered with the memory manager")
		return errors.New("VM is not registered with the memory manager")
	}

	if state.isActive {
		logger.Error("Failed to deactivate, VM still active")
		return errors.New("Failed to deactivate, VM still active")
	}

	if m.UseDSA {
		idxd.DSA_close_c()
	}

	delete(m.instances, vmID)

	return nil
}

// Activate Creates an epoller to serve page faults for the VM
func (m *MemoryManager) Activate(vmID string) error {
	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Activating instance in the memory manager")

	var (
		ok      bool
		state   *SnapshotState
		readyCh chan int = make(chan int)
	)

	m.Lock()

	state, ok = m.instances[vmID]
	if !ok {
		m.Unlock()
		logger.Error("VM not registered with the memory manager")
		return errors.New("VM not registered with the memory manager")
	}

	m.Unlock()

	if state.isActive {
		logger.Error("VM already active")
		return errors.New("VM already active")
	}

	if err := state.mapGuestMemory(); err != nil {
		logger.Error("Failed to map guest memory")
		return err
	}

	if err := state.getUFFD(); err != nil {
		logger.Error("Failed to get uffd")
		return err
	}

	state.setupStateOnActivate()

	go state.pollUserPageFaults(readyCh)

	<-readyCh

	return nil
}

// FetchState Fetches the working set file (or the whole guest memory) and the VMM state file
func (m *MemoryManager) FetchState(vmID string) error {
	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Activating instance in the memory manager")

	var (
		ok     bool
		state  *SnapshotState
		tStart time.Time
		err    error
	)

	m.Lock()

	state, ok = m.instances[vmID]
	if !ok {
		m.Unlock()
		logger.Error("VM not registered with the memory manager")
		return errors.New("VM not registered with the memory manager")
	}

	m.Unlock()

	if state.isRecordReady && !state.IsLazyMode {
		if state.metricsModeOn {
			tStart = time.Now()
		}
		err = state.fetchState()
		if state.metricsModeOn {
			state.currentMetric.MetricMap[fetchStateMetric] = metrics.ToUS(time.Since(tStart))
		}
	}

	return err
}

// Deactivate Removes the epoller which serves page faults for the VM
func (m *MemoryManager) Deactivate(vmID string) error {
	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Deactivating instance from the memory manager")

	var (
		state *SnapshotState
		ok    bool
	)

	m.Lock()

	state, ok = m.instances[vmID]
	if !ok {
		m.Unlock()
		logger.Error("VM not registered with the memory manager")
		return errors.New("VM not registered with the memory manager")
	}

	m.Unlock()

	if !state.isEverActivated {
		return nil
	}

	if !state.isActive {
		logger.Error("VM not activated")
		return errors.New("VM not activated")
	}

	state.quitCh <- 0
	if err := state.unmapGuestMemory(); err != nil {
		logger.Error("Failed to munmap guest memory")
		return err
	}

	state.processMetrics()

	state.userFaultFD.Close()
	if !state.isRecordReady && !state.IsLazyMode {
		state.ProcessRecord(state.GuestMemPath, state.WorkingSetPath)
	}

	state.isRecordReady = true
	state.isActive = false

	return nil
}

// DumpUPFPageStats Saves the per VM stats
func (m *MemoryManager) DumpUPFPageStats(vmID, functionName, metricsOutFilePath string) error {
	var (
		statHeader []string
		stats      []string
	)

	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Dumping stats about number of page faults")

	m.Lock()

	state, ok := m.instances[vmID]
	if !ok {
		m.Unlock()
		logger.Error("VM not registered with the memory manager")
		return errors.New("VM not registered with the memory manager")
	}

	m.Unlock()

	if state.isActive {
		logger.Error("Cannot get stats while VM is active")
		return errors.New("Cannot get stats while VM is active")
	}

	if !m.MetricsModeOn || !state.metricsModeOn {
		logger.Error("Metrics mode is not on")
		return errors.New("Metrics mode is not on")
	}

	if state.IsLazyMode {
		statHeader, stats = getLazyHeaderStats(state, functionName)
	} else {
		statHeader, stats = getRecRepHeaderStats(state, functionName)
	}

	return writeUPFPageStats(metricsOutFilePath, statHeader, stats)
}

// DumpUPFLatencyStats Dumps latency stats collected for the VM
func (m *MemoryManager) DumpUPFLatencyStats(vmID, functionName, latencyOutFilePath string) error {
	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("Dumping stats about latency of UPFs")

	m.Lock()

	state, ok := m.instances[vmID]
	if !ok {
		m.Unlock()
		logger.Error("VM not registered with the memory manager")
		return errors.New("VM not registered with the memory manager")
	}

	m.Unlock()

	if state.isActive {
		logger.Error("Cannot get stats while VM is active")
		return errors.New("Cannot get stats while VM is active")
	}

	if !m.MetricsModeOn || !state.metricsModeOn {
		logger.Error("Metrics mode is not on")
		return errors.New("Metrics mode is not on")
	}

	return metrics.PrintMeanStd(latencyOutFilePath, functionName, state.latencyMetrics...)

}

// GetUPFLatencyStats Returns the gathered metrics for the VM
func (m *MemoryManager) GetUPFLatencyStats(vmID string) ([]*metrics.Metric, error) {
	logger := log.WithFields(log.Fields{"vmID": vmID})

	logger.Debug("returning stats about latency of UPFs")

	m.Lock()

	state, ok := m.instances[vmID]
	if !ok {
		m.Unlock()
		logger.Error("VM not registered with the memory manager")
		return nil, errors.New("VM not registered with the memory manager")
	}

	m.Unlock()

	if state.isActive {
		logger.Error("Cannot get stats while VM is active")
		return nil, errors.New("Cannot get stats while VM is active")
	}

	if !m.MetricsModeOn || !state.metricsModeOn {
		logger.Error("Metrics mode is not on")
		return nil, errors.New("Metrics mode is not on")
	}

	return state.latencyMetrics, nil
}

func getLazyHeaderStats(state *SnapshotState, functionName string) ([]string, []string) {
	header := []string{
		"FuncName",
		"RecPages",
		"RepPages",
		"StdDev",
		"Reused",
		"StdDev",
		"Unique",
		"StdDev",
	}

	uniqueMean, uniqueStd := stat.MeanStdDev(state.uniquePFServed, nil)
	totalMean, totalStd := stat.MeanStdDev(state.totalPFServed, nil)
	reusedMean, reusedStd := stat.MeanStdDev(state.reusedPFServed, nil)

	stats := []string{
		functionName,
		strconv.Itoa(len(state.trace.trace)), // number of records (i.e., offsets)
		strconv.Itoa(int(totalMean)),         // number of pages served
		fmt.Sprintf("%.1f", totalStd),
		strconv.Itoa(int(reusedMean)), // number of pages found in the trace
		fmt.Sprintf("%.1f", reusedStd),
		strconv.Itoa(int(uniqueMean)), // number of pages not found in the trace
		fmt.Sprintf("%.1f", uniqueStd),
	}

	return header, stats
}

func getRecRepHeaderStats(state *SnapshotState, functionName string) ([]string, []string) {
	header := []string{
		"FuncName",
		"RecPages",
		"RecRegions",
		"Unique",
		"StdDev",
	}

	uniqueMean, uniqueStd := stat.MeanStdDev(state.uniquePFServed, nil)

	stats := []string{
		functionName,
		strconv.Itoa(len(state.trace.trace)),   // number of records (i.e., offsets)
		strconv.Itoa(len(state.trace.regions)), // number of contiguous regions in the trace
		strconv.Itoa(int(uniqueMean)),          // number of pages not found in the trace
		fmt.Sprintf("%.1f", uniqueStd),
	}

	return header, stats
}

func writeUPFPageStats(metricsOutFilePath string, statHeader, stats []string) error {
	csvFile, err := os.OpenFile(metricsOutFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Failed to create csv file for writing stats")
		return err
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	fileInfo, err := csvFile.Stat()
	if err != nil {
		log.Errorf("Failed to stat csv file: %v", err)
		return err
	}

	if fileInfo.Size() == 0 {
		if err := writer.Write(statHeader); err != nil {
			log.Errorf("Failed to write header to csv file: %v", err)
			return err
		}
	}

	if err := writer.Write(stats); err != nil {
		log.Errorf("Failed to write to csv file: %v ", err)
		return err
	}

	return nil
}
