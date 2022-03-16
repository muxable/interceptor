package rfc8888

import (
	"time"

	"github.com/pion/rtcp"
)

type streamLog struct {
	ssrc               uint32
	sequence           unwrapper
	init               bool
	nextSequenceNumber int64 // next to report
	lastSequenceNumber int64 // highest received
	log                map[int64]*packetReport
}

func newStreamLog(ssrc uint32) *streamLog {
	return &streamLog{
		ssrc:               ssrc,
		sequence:           unwrapper{},
		init:               false,
		nextSequenceNumber: 0,
		lastSequenceNumber: 0,
		log:                map[int64]*packetReport{},
	}
}

func (l *streamLog) add(ts time.Time, sequenceNumber uint16, ecn uint8) {
	unwrappedSequenceNumber := l.sequence.unwrap(sequenceNumber)
	if !l.init {
		l.init = true
		l.nextSequenceNumber = unwrappedSequenceNumber
	}
	l.log[unwrappedSequenceNumber] = &packetReport{
		arrivalTime: ts,
		ecn:         ecn,
	}
	if l.lastSequenceNumber < unwrappedSequenceNumber {
		l.lastSequenceNumber = unwrappedSequenceNumber
	}
}

// metricsAfter iterates over all packets order of their sequence number.
// Packets are removed until the first loss is detected.
func (l *streamLog) metricsAfter(reference time.Time) rtcp.CCFeedbackReportBlock {
	if len(l.log) == 0 {
		return rtcp.CCFeedbackReportBlock{
			MediaSSRC:     l.ssrc,
			BeginSequence: uint16(l.nextSequenceNumber),
			MetricBlocks:  []rtcp.CCFeedbackMetricBlock{},
		}
	}
	first := l.nextSequenceNumber
	last := l.nextSequenceNumber
	lost := false
	metricBlocks := make([]rtcp.CCFeedbackMetricBlock, l.lastSequenceNumber-first+1)
	for i := first; i <= l.lastSequenceNumber; i++ {
		received := false
		ecn := uint8(0)
		ato := uint16(0)
		if report, ok := l.log[i]; ok {
			received = true
			ecn = report.ecn
			ato = getArrivalTimeOffset(reference, report.arrivalTime)
		}
		metricBlocks[i-first] = rtcp.CCFeedbackMetricBlock{
			Received:          received,
			ECN:               rtcp.ECN(ecn),
			ArrivalTimeOffset: ato,
		}
		if !lost && i == last+1 {
			delete(l.log, i)
			l.nextSequenceNumber++
		} else {
			lost = true
		}
	}
	return rtcp.CCFeedbackReportBlock{
		MediaSSRC:     l.ssrc,
		BeginSequence: uint16(first),
		MetricBlocks:  metricBlocks,
	}
}

func getArrivalTimeOffset(base time.Time, arrival time.Time) uint16 {
	if base.Before(arrival) {
		return 0x1FFF
	}
	ato := uint16(base.Sub(arrival).Seconds() * 1024.0)
	if ato > 0x1FFD {
		return 0x1FFE
	}
	return ato
}
