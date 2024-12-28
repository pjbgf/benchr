package benchr

import "io"

type Option func(*Benchr)

func WithNoOutput() Option {
	return func(b *Benchr) {
		b.benchResults = io.Discard
	}
}

func WithAllocsChart(w io.WriteCloser) Option {
	return func(b *Benchr) {
		b.allocsChart = w
	}
}

func WithNsChart(w io.WriteCloser) Option {
	return func(b *Benchr) {
		b.nsChart = w
	}
}
