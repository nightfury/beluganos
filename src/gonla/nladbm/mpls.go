// -*- coding: utf-8 -*-

// Copyright (C) 2017 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nladbm

import (
	"gonla/nlamsg"
	"sync"
)

//
// Key
//
type MplsKey struct {
	NId    uint8
	LLabel uint32
}

func NewMplsKey(nid uint8, llabel uint32) *MplsKey {
	return &MplsKey{
		NId:    nid,
		LLabel: llabel,
	}
}

func MplsToKey(r *nlamsg.Route) *MplsKey {
	if r.Route.MPLSDst == nil {
		return nil
	}
	return NewMplsKey(r.NId, uint32(*r.Route.MPLSDst))
}

//
// Table interface
//
type MplsTable interface {
	Insert(*nlamsg.Route) *nlamsg.Route
	Update(*nlamsg.Route) *nlamsg.Route
	InsOrUpdate(*nlamsg.Route) *nlamsg.Route
	Select(*MplsKey) *nlamsg.Route
	Delete(*MplsKey) *nlamsg.Route
	Walk(f func(*nlamsg.Route) error) error
	WalkFree(f func(*nlamsg.Route) error) error
}

func NewMplsTable() MplsTable {
	return &mplsTable{
		Mplss: make(map[MplsKey]*nlamsg.Route),
	}
}

//
// Table
//
type mplsTable struct {
	Mutex sync.RWMutex
	Mplss map[MplsKey]*nlamsg.Route
}

func (t *mplsTable) find(key *MplsKey) *nlamsg.Route {
	n, _ := t.Mplss[*key]
	return n
}

func (t *mplsTable) Insert(r *nlamsg.Route) (old *nlamsg.Route) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	key := MplsToKey(r)
	if old = t.find(key); old == nil {
		t.Mplss[*key] = r.Copy()
	}

	return
}

func (t *mplsTable) Update(r *nlamsg.Route) (old *nlamsg.Route) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	key := MplsToKey(r)
	if old = t.find(key); old != nil {
		t.Mplss[*key] = r.Copy()
	}

	return
}

func (t *mplsTable) InsOrUpdate(r *nlamsg.Route) (old *nlamsg.Route) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	key := MplsToKey(r)
	old = t.find(key)
	t.Mplss[*key] = r.Copy()

	return
}

func (t *mplsTable) Select(key *MplsKey) *nlamsg.Route {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()

	return t.find(key)
}

func (t *mplsTable) Walk(f func(*nlamsg.Route) error) error {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()

	return t.WalkFree(f)
}

func (t *mplsTable) WalkFree(f func(*nlamsg.Route) error) error {
	for _, n := range t.Mplss {
		if err := f(n); err != nil {
			return err
		}
	}
	return nil
}

func (t *mplsTable) Delete(key *MplsKey) (old *nlamsg.Route) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	if old = t.find(key); old != nil {
		delete(t.Mplss, *key)
	}

	return
}
