package lfu

// Just, you know, a bidirectional linked list
type linkedListNode[Data any] struct {
	data Data
	next *linkedListNode[Data]
	prev *linkedListNode[Data]
}

type linkedList[Data any] struct {
	head *linkedListNode[Data]
	tail *linkedListNode[Data]
}

func (l *linkedList[Data]) add(data Data) *linkedListNode[Data] {
	node := new(linkedListNode[Data])
	node.data = data

	return l.addNode(node)
}

func (l *linkedList[Data]) addNode(node *linkedListNode[Data]) *linkedListNode[Data] {
	node.prev = l.tail
	node.next = nil
	if l.tail == nil {
		l.head = node
	} else {
		l.tail.next = node
	}
	l.tail = node

	return node
}

func (l *linkedList[Data]) insertAfter(pivot *linkedListNode[Data], data Data) *linkedListNode[Data] {
	pNext := pivot.next

	node := new(linkedListNode[Data])
	node.data = data

	pivot.next = node
	node.prev = pivot
	node.next = pNext
	if pNext != nil {
		pNext.prev = node
	} else {
		l.tail = node
	}

	return node
}

func (l *linkedList[Data]) isEmpty() bool {
	return l.head == nil
}

func (l *linkedList[Data]) remove(node *linkedListNode[Data]) {
	if l.head == node && l.tail == node {
		l.head = nil
		l.tail = nil
		return
	}

	nPrev := node.prev
	nNext := node.next

	if nPrev != nil {
		nPrev.next = nNext
	} else {
		l.head = nNext
	}

	if nNext != nil {
		nNext.prev = nPrev
	} else {
		l.tail = nPrev
	}

}
