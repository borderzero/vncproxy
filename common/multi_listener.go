package common

type MultiListener struct {
	listeners []SegmentConsumer
}

func (m *MultiListener) AddListener(listener SegmentConsumer) {
	m.listeners = append(m.listeners, listener)
}

func (m *MultiListener) Consume(seg *RfbSegment) error {
	for _, li := range m.listeners {
		err := li.Consume(seg)
		if err != nil {
			return err
		}
	}
	return nil
}

type BestEffortMultiListener struct {
	listeners []SegmentConsumer
}

func (m *BestEffortMultiListener) AddListener(listener SegmentConsumer) {
	m.listeners = append(m.listeners, listener)
}

func (m *BestEffortMultiListener) Consume(seg *RfbSegment) error {
	for _, li := range m.listeners {
		go li.Consume(seg)
	}
	return nil
}
