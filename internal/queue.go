package internal

type Queue []interface{}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Enqueue(val interface{}) {
	*q = append(*q, val)
}

func (q *Queue) Dequeue() interface{} {

	h := *q
	l := len(h)

	var el interface{}
	el, *q = h[0], h[1:l]
	return el
}

func (q *Queue) len() int {
	return len(*q)
}
