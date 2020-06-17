package tools

import "sync"

var copiedByHostRegistry map[string]*CopyState
var once sync.Once
var copyMutex = &sync.Mutex{}

func GetCopiedRegistry() map[string]*CopyState {
	once.Do(func() {
		copiedByHostRegistry = make(map[string]*CopyState)
	})
	return copiedByHostRegistry
}

type CopyState struct {
	Copying bool
	Ready   bool
	Wg      sync.WaitGroup
}

func CopyOnce(src string, dest string) (int64, error) {

	var size int64
	var err error

	keypath := src + ";" + dest

	copyMutex.Lock()
	r := GetCopiedRegistry()
	if r[keypath] == nil {
		r[keypath] = &CopyState{}
	}
	if r[keypath].Ready {
		copyMutex.Unlock()
		return size, err
	}
	if r[keypath].Copying {
		r[keypath].Wg.Wait()
		if r[keypath].Ready {
			copyMutex.Unlock()
			return size, err
		}
	}
	r[keypath].Wg.Add(1)
	r[keypath].Copying = true
	copyMutex.Unlock()

	size, err = Copy(src, dest)

	if err == nil {
		r[keypath].Ready = true
	} else {
		r[keypath].Ready = false
	}
	r[keypath].Copying = false
	r[keypath].Wg.Done()

	return size, err
}
