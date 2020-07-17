package replace

import (
	"bytes"

	"golang.org/x/text/transform"
)

type Replacer struct {
	old []byte
	new []byte

	transform.NopResetter
}

var _ transform.Transformer = (*Replacer)(nil)

func New(old, new []byte) *Replacer {
	return &Replacer{old: old, new: new}
}

func (r *Replacer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	// don't do anything for empty old string
	if len(r.old) == 0 {
		n, err := fullcopy(dst, src)
		if err != nil {
			return 0, 0, err
		}
		return n, n, nil
	}
	// make sure there's enough to even find a match
	if len(src) < len(r.old) {
		if atEOF {
			n, err := fullcopy(dst, src)
			if err != nil {
				return 0, 0, err
			}
			return n, n, nil
		}
		return 0, 0, transform.ErrShortSrc
	}
	// replace all instances of old with new
	for {
		i := bytes.Index(src[nSrc:], r.old)
		if i == -1 {
			break
		}
		n1, err := fullcopy(dst[nDst:], src[nSrc:nSrc+i])
		if err != nil {
			return nDst, nSrc, err
		}
		n2, err := fullcopy(dst[nDst+i:], r.new)
		if err != nil {
			return nDst, nSrc, err
		}
		nDst += n1 + n2
		nSrc += i + len(r.old)
	}
	// skip everything except the trailing len(r.old) - 1
	if len(src[nSrc:]) >= len(r.old) {
		skip := len(src[nSrc:]) - len(r.old) + 1
		n, err := fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		if err != nil {
			return nDst, nSrc, err
		}
		nSrc += n
		nDst += n
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err := fullcopy(dst[nDst:], src[nSrc:])
		if err != nil {
			return nDst, nSrc, err
		}
		nDst += n
		nSrc += n
		return nDst, nSrc, nil
	}
	return nDst, nSrc, transform.ErrShortSrc
}

func fullcopy(dst, src []byte) (int, error) {
	n := copy(dst, src)
	if n < len(src) {
		return 0, transform.ErrShortDst
	}
	return n, nil
}
