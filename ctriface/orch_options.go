// MIT License
//
// Copyright (c) 2020 Plamen Petrov and EASE lab
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

package ctriface

// OrchestratorOption Options to pass to Orchestrator
type OrchestratorOption func(*Orchestrator)

// WithTestModeOn Sets the test mode
func WithTestModeOn(testModeOn bool) OrchestratorOption {
	return func(o *Orchestrator) {
		if !testModeOn {
			o.setupCloseHandler()
			o.setupHeartbeat()
		}
	}
}

// WithSnapshots Sets the snapshot mode on or off
func WithSnapshots(snapshotsEnabled bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.snapshotsEnabled = snapshotsEnabled
	}
}

// WithUPF Sets the user-page faults mode on or off
func WithUPF(isUPFEnabled bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.isUPFEnabled = isUPFEnabled
	}
}

// InMemWorkingSet Sets the in-memory working set on or off
func InMemWorkingSet(inMemWorkingSet bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.inMemWorkingSet = inMemWorkingSet
	}
}

// InNumaWorkingSet Sets the remote numa working set on or off
func InNumaWorkingSet(inNumaWorkingSet bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.inNumaWorkingSet = inNumaWorkingSet
	}
}

// InCxlMem Sets the in-memory working set on or off
func InCxlMem(inCxlMem bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.inCxlMem = inCxlMem
	}
}

// UseDSA Sets whether to use DSA or not
func UseDSA(useDSA bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.useDSA = useDSA
	}
}

// MovePages Sets whether to use move_pages or not
func MovePages(movePages bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.movePages = movePages
	}
}

// WithSnapshotsDir Sets the directory where
// snapshots should be stored
func WithSnapshotsDir(snapshotsDir string) OrchestratorOption {
	return func(o *Orchestrator) {
		o.snapshotsDir = snapshotsDir
	}
}

// WithLazyMode Sets the lazy paging mode on (or off),
// where all guest memory pages are brought on demand.
// Only works if snapshots are enabled
func WithLazyMode(isLazyMode bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.isLazyMode = isLazyMode
	}
}

// WithMetricsMode Sets the metrics mode
func WithMetricsMode(isMetricsMode bool) OrchestratorOption {
	return func(o *Orchestrator) {
		o.isMetricsMode = isMetricsMode
	}
}

// WithCustomHostIface Sets the custom host net interface
// for the VMs to link to
func WithCustomHostIface(hostIface string) OrchestratorOption {
	return func(o *Orchestrator) {
		o.hostIface = hostIface
	}
}
