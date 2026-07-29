package skiplist

type Scanner struct {
	release func()
	ready   *SkipListNode
	next    *SkipListNode
}

func (s Scanner) Next() bool {
	if nil == s.next {
		return false
	}

	s.ready = s.next
	s.next = s.ready.next[0]

	return true
}

func (s Scanner) Key() []byte {
	if nil == s.ready {
		return nil
	}

	return s.ready.key
}

func (s Scanner) Value() []byte {
	if nil == s.ready {
		return nil
	}

	return s.ready.value
}

func (s Scanner) Release() error {
	s.release()
	return nil
}
