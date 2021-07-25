package list

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

type List Item

func (r *List) Head() *Item {
	return (*Item)(r)
}

func (r *List) Count() int {
	_, count := r.Head().Tail()
	return count
}
func (r *List) Append(list *List) {
	r.Head().Append(list.Head())
}

func (r *List) SelectByIndex(i int) *Item {
	item := r.Head()
	index := 0
	for {
		if item == nil {
			return nil
		}
		if index == i {
			return item
		}
		item = item.Next()
		index++
	}
}
func (r *List) SelectByValue(value interface{}) (item *Item, index int) {
	item = r.Head()
	for {
		if item == nil {
			return nil, -1
		}
		if item.Value() == value {
			return item, index
		}
		item = item.Next()
		index++
	}
}
func (r *List) VisitAll(cb func(value interface{}, index int) (Continue bool)) {
	item := r.Head()
	index := 0
	for {
		if item == nil {
			return
		}
		if !cb(item.Value(), index) {
			return
		}
		item = item.Next()
		index++
	}
}

type Item struct {
	value interface{}
	next  *Item
}

func New(head interface{}) *List {
	return (&Item{
		value: head,
		next:  nil,
	}).AsList()
}

func (r *Item) Value() interface{} {
	return r.value
}

func (r *Item) Next() *Item {
	ptr := unsafe.Pointer(r.next)
	return (*Item)(atomic.LoadPointer(&ptr))
}
func (r *Item) AsList() *List {
	return (*List)(r)
}

func (r *Item) UnlinkNext() bool {
	for {
		next := r.Next()
		if next == nil {
			return false
		}
		if r.casNext(next, next.Next()) {
			return true
		}
		runtime.Gosched()
	}
}

func (r *Item) casNext(old, new *Item) bool {
	next := unsafe.Pointer(r.next)
	return atomic.CompareAndSwapPointer(&next, unsafe.Pointer(old), unsafe.Pointer(new))
}

func (r *Item) Append(value interface{}) {
	item := unsafe.Pointer(&Item{
		value: value,
		next:  nil,
	})
	for {
		tail, _ := r.Tail()
		ptr := unsafe.Pointer(tail.next)
		if atomic.CompareAndSwapPointer(&ptr, nil, item) {
			break
		}
		runtime.Gosched()
	}
}
func (r *Item) Tail() (tail *Item, count int) {
	count = 1
	tail = r
	next := r.Next()
	for next != nil {
		count++
		tail = next
		next = next.Next()
	}
	return
}
