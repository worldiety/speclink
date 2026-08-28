// Package classfile reads compiled JVM class files.
//
// speclink reads bytecode rather than source for the JVM, and the reason is not
// convenience. A class file has already been resolved: every type it names is
// fully qualified, every supertype and interface is named outright, every
// annotation has had its import worked out. A source reader has to redo all of
// that — imports, wildcards, the implicit java.lang, and inheritance that
// disappears the moment it leaves the project — and every one of those is a
// place to be silently wrong about which framework marker applies.
//
// It also means one reader serves three targets. Kotlin compiles to JVM
// bytecode and only then is dexed, so Java on Spring, Kotlin on Spring and
// Kotlin on Android all arrive here in the same shape. What differs between
// them is where the files sit, not what is in them.
//
// The cost is paid in positions. LineNumberTable is an attribute of Code, and
// fields have no code, so a field's declaration line does not exist here at all.
// That is recovered separately and best effort; it is not recovered by pretending.
//
// # Forward compatibility
//
// The ClassFile structure has not changed since version 45.3. Everything since
// is additive: new constant pool tags and new attributes. JVMS §4.7.1 requires
// readers to silently ignore attributes they do not recognise, and doing so is
// what makes one reader work from Java 8 to Java 25 without a version table.
//
// The parts that must be exactly right are the ones where being wrong is quiet:
// Long and Double occupy two constant pool slots, and CONSTANT_Utf8 is not UTF-8.
package classfile

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unicode/utf16"
	"unicode/utf8"
)

// Magic opens every class file.
const Magic = 0xCAFEBABE

// Constant pool tags, JVMS §4.4.
const (
	TagUtf8               = 1
	TagInteger            = 3
	TagFloat              = 4
	TagLong               = 5
	TagDouble             = 6
	TagClass              = 7
	TagString             = 8
	TagFieldref           = 9
	TagMethodref          = 10
	TagInterfaceMethodref = 11
	TagNameAndType        = 12
	TagMethodHandle       = 15
	TagMethodType         = 16
	TagDynamic            = 17
	TagInvokeDynamic      = 18
	TagModule             = 19
	TagPackage            = 20
)

// Constant is one entry of the constant pool.
//
// It is deliberately one struct rather than an interface per tag. Almost every
// entry is read for a name or a value, the rest exist only so the pool can be
// walked to its end, and a dozen types to express that would be ceremony.
type Constant struct {
	Tag byte
	// Text holds the decoded string of a CONSTANT_Utf8.
	Text string
	// Index and Index2 hold the referenced pool entries, meaning what the tag
	// says: for Class the name, for NameAndType the name and descriptor, for
	// the ref kinds the class and the name-and-type.
	Index, Index2 uint16
	// Value holds the constant of an Integer, Float, Long or Double.
	Value any
}

// Pool is a constant pool, indexed as the class file indexes it: entries start
// at 1, and a Long or Double is followed by an unusable slot.
type Pool []Constant

// UTF8 returns the string at an index, and whether it is one.
func (p Pool) UTF8(i uint16) (string, bool) {
	c, ok := p.at(i)
	if !ok || c.Tag != TagUtf8 {
		return "", false
	}
	return c.Text, true
}

// Class returns the internal name of a CONSTANT_Class, e.g. "java/lang/String".
func (p Pool) Class(i uint16) (string, bool) {
	c, ok := p.at(i)
	if !ok || c.Tag != TagClass {
		return "", false
	}
	return p.UTF8(c.Index)
}

// Const returns the value of a numeric or string constant.
func (p Pool) Const(i uint16) (any, bool) {
	c, ok := p.at(i)
	if !ok {
		return nil, false
	}
	switch c.Tag {
	case TagInteger, TagFloat, TagLong, TagDouble:
		return c.Value, true
	case TagString:
		s, ok := p.UTF8(c.Index)
		return s, ok
	}
	return nil, false
}

func (p Pool) at(i uint16) (Constant, bool) {
	if int(i) == 0 || int(i) >= len(p) {
		return Constant{}, false
	}
	return p[i], true
}

// reader walks a class file, tracking the first failure.
//
// Errors are sticky rather than checked at every read. A class file is either
// well formed or it is not, and threading an error through a hundred two byte
// reads would bury the structure this file exists to make visible.
type reader struct {
	b   []byte
	pos int
	err error
}

func (r *reader) fail(format string, args ...any) {
	if r.err == nil {
		r.err = fmt.Errorf(format, args...)
	}
}

func (r *reader) u1() byte {
	if r.err != nil || r.pos+1 > len(r.b) {
		r.fail("unexpected end of class file at offset %d", r.pos)
		return 0
	}
	v := r.b[r.pos]
	r.pos++
	return v
}

func (r *reader) u2() uint16 {
	if r.err != nil || r.pos+2 > len(r.b) {
		r.fail("unexpected end of class file at offset %d", r.pos)
		return 0
	}
	v := binary.BigEndian.Uint16(r.b[r.pos:])
	r.pos += 2
	return v
}

func (r *reader) u4() uint32 {
	if r.err != nil || r.pos+4 > len(r.b) {
		r.fail("unexpected end of class file at offset %d", r.pos)
		return 0
	}
	v := binary.BigEndian.Uint32(r.b[r.pos:])
	r.pos += 4
	return v
}

func (r *reader) bytes(n int) []byte {
	if r.err != nil || n < 0 || r.pos+n > len(r.b) {
		r.fail("unexpected end of class file at offset %d", r.pos)
		return nil
	}
	v := r.b[r.pos : r.pos+n]
	r.pos += n
	return v
}

// skip advances past a region whose contents are not needed.
func (r *reader) skip(n int) {
	if r.err != nil || n < 0 || r.pos+n > len(r.b) {
		r.fail("attribute claims %d bytes but only %d remain", n, len(r.b)-r.pos)
		return
	}
	r.pos += n
}

// readPool reads the constant pool.
//
// The two rules that must hold and are easy to get wrong: entries are numbered
// from 1, and a Long or Double takes two slots with the second unusable. Miss
// the second and every index after the first long is off by one — which does
// not fail, it silently reads the wrong names.
func (r *reader) readPool() Pool {
	count := int(r.u2())
	if count < 1 {
		r.fail("constant pool count is %d", count)
		return nil
	}
	pool := make(Pool, count)

	for i := 1; i < count && r.err == nil; i++ {
		tag := r.u1()
		c := Constant{Tag: tag}

		switch tag {
		case TagUtf8:
			n := int(r.u2())
			raw := r.bytes(n)
			text, err := decodeModifiedUTF8(raw)
			if err != nil {
				r.fail("constant %d: %v", i, err)
			}
			c.Text = text

		case TagInteger:
			c.Value = int32(r.u4())
		case TagFloat:
			c.Value = math.Float32frombits(r.u4())
		case TagLong:
			hi, lo := r.u4(), r.u4()
			c.Value = int64(hi)<<32 | int64(lo)
		case TagDouble:
			hi, lo := r.u4(), r.u4()
			c.Value = math.Float64frombits(uint64(hi)<<32 | uint64(lo))

		case TagClass, TagString, TagMethodType, TagModule, TagPackage:
			c.Index = r.u2()

		case TagFieldref, TagMethodref, TagInterfaceMethodref, TagNameAndType,
			TagDynamic, TagInvokeDynamic:
			c.Index = r.u2()
			c.Index2 = r.u2()

		case TagMethodHandle:
			c.Index = uint16(r.u1()) // reference kind
			c.Index2 = r.u2()

		default:
			// Refusing rather than skipping. An unknown tag has an unknown
			// width, so there is no way to find the next entry, and guessing
			// would produce a pool that reads plausibly and means nothing.
			r.fail("unknown constant pool tag %d at entry %d", tag, i)
		}

		pool[i] = c

		if tag == TagLong || tag == TagDouble {
			// JVMS §4.4.5: "the next usable item in the pool is located at
			// index n+2. The constant_pool index n+1 must be valid but is
			// considered unusable."
			i++
		}
	}
	return pool
}

// decodeModifiedUTF8 decodes a CONSTANT_Utf8.
//
// This is not UTF-8, and treating it as such is the quiet kind of wrong. JVMS
// §4.4.7 defines two deviations:
//
//   - U+0000 is encoded as the two bytes C0 80, so that a string never contains
//     a NUL byte.
//   - Characters above the basic multilingual plane are encoded as a surrogate
//     pair, each surrogate encoded separately in three bytes — CESU-8 rather
//     than UTF-8, which encodes them in four.
//
// A naive string(bytes) survives ASCII, which is most identifiers, and then
// corrupts exactly the things that matter: Kotlin's metadata blob, and any
// identifier or annotation string outside the BMP.
func decodeModifiedUTF8(b []byte) (string, error) {
	// The overwhelmingly common case is plain ASCII, where the two encodings
	// agree. Checking first costs one pass and skips the decoder entirely.
	ascii := true
	for _, c := range b {
		if c >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		return string(b), nil
	}

	var (
		out   []rune
		i     int
		pend  rune // a high surrogate awaiting its pair
		flush = func() {
			if pend != 0 {
				out = append(out, utf8.RuneError)
				pend = 0
			}
		}
	)
	for i < len(b) {
		c := b[i]
		switch {
		case c < 0x80:
			flush()
			out = append(out, rune(c))
			i++

		case c&0xE0 == 0xC0:
			if i+1 >= len(b) {
				return "", errors.New("truncated two byte sequence in modified UTF-8")
			}
			r := rune(c&0x1F)<<6 | rune(b[i+1]&0x3F)
			flush()
			out = append(out, r)
			i += 2

		case c&0xF0 == 0xE0:
			if i+2 >= len(b) {
				return "", errors.New("truncated three byte sequence in modified UTF-8")
			}
			r := rune(c&0x0F)<<12 | rune(b[i+1]&0x3F)<<6 | rune(b[i+2]&0x3F)
			i += 3

			switch {
			case utf16.IsSurrogate(r) && pend == 0 && r < 0xDC00:
				pend = r
			case utf16.IsSurrogate(r) && pend != 0 && r >= 0xDC00:
				out = append(out, utf16.DecodeRune(pend, r))
				pend = 0
			default:
				flush()
				out = append(out, r)
			}

		default:
			// Four byte sequences do not occur in modified UTF-8; a supplementary
			// character is a surrogate pair of three byte sequences. Seeing one
			// means this is ordinary UTF-8 and something upstream is wrong.
			return "", fmt.Errorf("byte %#x is not valid in modified UTF-8", c)
		}
	}
	flush()
	return string(out), nil
}
