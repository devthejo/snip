package tools

import (
	"sync"

	cache "github.com/patrickmn/go-cache"
)

var mu = &sync.Mutex{}

type RequiredOnceState struct {
	Loading bool
	Ready   bool
	ValidId string
	Wg      sync.WaitGroup
}

func RequiredOnce(r *cache.Cache, keys []string, validId string, runTask func() (interface{}, error)) (interface{}, error) {

	var result interface{}
	var err error

	var b []byte
	b, err = json.Marshal(keys)
	if err != nil {
		return result, err
	}

	keypath := string(b)

	mu.Lock()
	var state *RequiredOnceState
	if stateI, found := r.Get(keypath); found {
		state = stateI.(*RequiredOnceState)
		if state.ValidId != validId {
			state = nil
		}
	}

	if state == nil {
		state = &RequiredOnceState{
			ValidId: validId,
		}
		r.Set(keypath, state, cache.DefaultExpiration)
	}

	if state.Ready {
		mu.Unlock()
		return result, err
	}
	if state.Loading {
		state.Wg.Wait()
		if state.Ready {
			mu.Unlock()
			return result, err
		}
	}
	state.Wg.Add(1)
	state.Loading = true
	mu.Unlock()

	result, err = runTask()

	if err == nil {
		state.Ready = true
	} else {
		state.Ready = false
	}
	state.Loading = false
	state.Wg.Done()

	return result, err
}
