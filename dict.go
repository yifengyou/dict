// Copyright (c) 2019 srfrog - https://srfrog.me
// Use of this source code is governed by the license in the LICENSE file.

package dict

// Dict is a type that uses a hash mapping index, also known as a dictionary.
type Dict struct {
	size, version int
	keys          []*Key
	values        map[uint64]interface{}
}

// Version returns the version of the dictionary. The version is increased after every
// change to dict items.
// Returns version, which is zero (0) initially.
func (d *Dict) Version() int { return d.version }

// Len returns the size of a Dict.
func (d *Dict) Len() int { return d.size }

// New returns a new Dict object.
// vargs can be any Go basic type, slices, and maps. The keys in a map are
// used as keys in the dict. The map keys must be hashable.
// map的key必须支持序列化
// 支持切片和map
func New(vargs ...interface{}) *Dict {
	d := &Dict{values: make(map[uint64]interface{})}
	// 关键在于Update
	d.Update(vargs...)
	return d
}

// Set inserts a new item into the dict. If a value matching the key already exists,
// its value is replaced, otherwise a new item is added.
func (d *Dict) Set(key, value interface{}) *Dict {
	// Sanity: don't panic on nil dict, just create a new one.
	if d == nil {
		d = New()
	}

	k := MakeKey(key)
	if k == nil {
		return d
	}

	// 如果k.ID已经存在
	if _, ok := d.values[k.ID]; ok {
		d.values[k.ID] = value
		return d
	}
	// 如果不存在，则新建
	d.keys = append(d.keys, k)
	d.values[k.ID] = value
	d.size++
	d.version++

	return d
}

// Get retrieves an item from dict by key. If alt value is passed, it will be used as
// default value if no item is found.
// Returns a value matching key in dict, otherwise nil or alt if given.
func (d *Dict) Get(key interface{}, alt ...interface{}) interface{} {
	// 判断是否为空
	if d.IsEmpty() {
		return nil
	}
	// 根据key获取ID
	h, ok := d.GetKeyID(key)
	if ok {
		return d.values[h]
	}
	// 如果有默认值，则返回默认值
	if alt != nil {
		return alt[0]
	}
	return nil
}

// GetKeyID retrieves the ID of an item in dict, if found.
// Returns the item ID and true, or 0 and false if not found.
func (d *Dict) GetKeyID(key interface{}) (uint64, bool) {
	if d.IsEmpty() {
		return 0, false
	}
	k := MakeKey(key)
	if k == nil {
		return 0, false
	}
	_, ok := d.values[k.ID]
	return k.ID, ok
}

func (d *Dict) deleteItem(idx int) {
	if d.IsEmpty() || idx >= d.size {
		return
	}

	delete(d.values, d.keys[idx].ID)
	copy(d.keys[idx:], d.keys[idx+1:])

	l := len(d.keys)
	// 最后一个元素置空
	d.keys[l-1] = nil
	d.keys = d.keys[:l-1]
	d.size = l
	d.version++ // 每次变动，version都要+1
}

// Del removes an item from dict by key name.
// Returns true if an item is found and removed, false otherwise.
func (d *Dict) Del(key interface{}) bool {
	id, ok := d.GetKeyID(key)
	if !ok {
		return false
	}

	var idx int
	for i := range d.keys {
		if d.keys[i].ID == id {
			idx = i
			break
		}
	}
	if idx > d.size || d.keys[idx].ID != id {
		return false
	}

	d.deleteItem(idx)

	return true
}

// Pop gets the value of a key and removes the item from the dict.
// If the item is not found it returns alt. Otherwise it will return the value or nil.
func (d *Dict) Pop(key interface{}, alt ...interface{}) interface{} {
	value := d.Get(key, alt)
	if value != nil {
		d.Del(key)
	}
	return value
}

// PopItem removes the most recent item added to the dict and returns it. If the dict is
// empty, returns nil.
func (d *Dict) PopItem() *Item {
	if d.IsEmpty() {
		return nil
	}

	key := d.keys[d.size-1]
	value := d.values[key.ID]
	d.deleteItem(d.size - 1)

	return &Item{
		Key:   key.Name,
		Value: value,
	}
}

// Key returns true if key is in dict d, false otherwise.
func (d *Dict) Key(key interface{}) bool {
	_, ok := d.GetKeyID(key)
	return ok
}

// IsEmpty returns true if the dict is empty, false otherwise.
func (d *Dict) IsEmpty() bool {
	return d == nil || d.size == 0
}

// Clear empties a Dict d.
// Returns true if the dict was actually cleared, otherwise false if nothing was done.
func (d *Dict) Clear() bool {
	if d.IsEmpty() {
		return false
	}
	d.size = 0
	d.version++ // not a new dict
	d.keys = []*Key{}
	d.values = make(map[uint64]interface{})
	return true
}

// Keys returns a string slice of all dict keys, or nil if dict is empty.
func (d *Dict) Keys() []string {
	if d.IsEmpty() {
		return nil
	}
	keys := make([]string, d.size)
	for i := range d.keys {
		keys[i] = d.keys[i].Name
	}
	return keys
}

// Values returns a slice of all dict values, or nil if dict is empty.
func (d *Dict) Values() []interface{} {
	if d.IsEmpty() {
		return nil
	}
	values := make([]interface{}, d.size)
	for i, key := range d.keys {
		values[i] = d.values[key.ID]
	}
	return values
}

// Items returns a channel of key-value items, or nil if the dict is empty.
func (d *Dict) Items() <-chan Item {
	// 无缓冲chan
	ci := make(chan Item)

	go func() {
		defer close(ci)
		if d.IsEmpty() {
			return
		}
		for i := range d.keys {
			ci <- Item{
				Key:   d.keys[i].Name,
				Value: d.values[d.keys[i].ID],
			}
		}
	}()

	return ci
}

// Update adds to d the key-value items from iterables, scalars and other dicts. Also replacing
// any existing values that match the keys. This func is used by New() when initializing a
// dict with values.
// Returns true if any changes were made.
// vargs必须为切片或map类型
func (d *Dict) Update(vargs ...interface{}) bool {
	if vargs == nil {
		return false
	}
	// New()则初始化为零值，int类型
	ver := d.Version()
	for i := range vargs {
		// 遍历vargs,i为从0开始的序列
		// 如果是用其他字典拼接，则依次添加
		if other, ok := vargs[i].(*Dict); ok {
			// 支持多个dict拼接，单个拼接仍是一个for循环，依次迭代
			for item := range other.Items() {
				d.Set(item.Key, item.Value)
			}
			continue
		}
		// 如果是切片或map类型，iterables and scalars
		// toIterable 为关键函数，将类型转为dict
		for item := range toIterable(vargs[i]) {
			if item.Key == nil {
				// 如果key不存在，则用大小来表示
				item.Key = d.size
			}
			d.Set(item.Key, item.Value)
		}
	}
	// 返回是否更新
	return ver != d.Version()
}
