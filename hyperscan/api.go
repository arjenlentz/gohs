package hyperscan

import (
	"fmt"
	"io"
)

type matchEvent struct {
	id       uint
	from, to uint64
	flags    ScanFlag
}

func (e *matchEvent) Id() uint { return e.id } // nolint: golint,revive,stylecheck

func (e *matchEvent) From() uint64 { return e.from }

func (e *matchEvent) To() uint64 { return e.to }

func (e *matchEvent) Flags() ScanFlag { return e.flags }

type matchRecorder struct {
	matched []matchEvent
	err     error
}

func (h *matchRecorder) Matched() bool { return h.matched != nil }

func (h *matchRecorder) Handle(id uint, from, to uint64, flags uint, context interface{}) error {
	if len(h.matched) > 0 {
		tail := &h.matched[len(h.matched)-1]

		if tail.id == id && tail.from == from && tail.to < to {
			tail.to = to

			return h.err
		}
	}

	h.matched = append(h.matched, matchEvent{id, from, to, ScanFlag(flags)})

	return h.err
}

// Match reports whether the byte slice b contains any match of the regular expression pattern.
func Match(pattern string, data []byte) (bool, error) {
	p, err := ParsePattern(pattern)
	if err != nil {
		return false, fmt.Errorf("parse pattern, %w", err)
	}

	p.Flags |= SingleMatch

	db, err := NewBlockDatabase(p)
	if err != nil {
		return false, fmt.Errorf("create block database, %w", err)
	}
	defer db.Close()

	s, err := NewScratch(db)
	if err != nil {
		return false, fmt.Errorf("create scratch, %w", err)
	}

	defer func() {
		_ = s.Free()
	}()

	h := &matchRecorder{}

	if err = db.Scan(data, s, h.Handle, nil); err != nil {
		return false, err // nolint: wrapcheck
	}

	return h.Matched(), h.err
}

// MatchReader reports whether the text returned by the Reader contains any match of the regular expression pattern.
func MatchReader(pattern string, reader io.Reader) (bool, error) {
	p, err := ParsePattern(pattern)
	if err != nil {
		return false, fmt.Errorf("parse pattern, %w", err)
	}

	p.Flags |= SingleMatch

	db, err := NewStreamDatabase(p)
	if err != nil {
		return false, fmt.Errorf("create stream database, %w", err)
	}
	defer db.Close()

	s, err := NewScratch(db)
	if err != nil {
		return false, fmt.Errorf("create scratch, %w", err)
	}

	defer func() {
		_ = s.Free()
	}()

	h := &matchRecorder{}

	if err = db.Scan(reader, s, h.Handle, nil); err != nil {
		return false, err // nolint: wrapcheck
	}

	return h.Matched(), h.err
}

// MatchString reports whether the string s contains any match of the regular expression pattern.
func MatchString(pattern string, s string) (matched bool, err error) {
	return Match(pattern, []byte(s))
}
