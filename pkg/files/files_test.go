package files

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__PathFromSource(t *testing.T) {
	check := func(dst, src, expDst string) {
		result := pathFromSource(dst, src)
		assert.Equal(t, expDst, result, dst, src)
	}

	check("", "/long/path/to/source", "source")
	check("", "/long/path/to/.source", ".source")
	check("", "long/path/to/source", "source")
	check("", "long/path/to/.source", ".source")
	check("destination", "/long/path/to/source", "destination")
	check(".destination", "/long/path/to/source", ".destination")
	check("destination", "/long/path/to/.source", "destination")
	check(".destination", "/long/path/to/.source", ".destination")
	check("destination", "long/path/to/source", "destination")
	check(".destination", "long/path/to/source", ".destination")
	check("destination", "long/path/to/.source", "destination")
	check(".destination", "long/path/to/.source", ".destination")
	check("long/path/to/destination", "long/path/to/source", "long/path/to/destination")
	check(".long/path/to/destination", "long/path/to/source", ".long/path/to/destination")
	check("long/path/to/destination", ".long/path/to/source", "long/path/to/destination")
	check(".long/path/to/destination", ".long/path/to/source", ".long/path/to/destination")
	check("/long/path/to/destination", "long/path/to/source", "/long/path/to/destination")
	check("/.long/path/to/destination", "long/path/to/source", "/.long/path/to/destination")
	check("/long/path/to/destination", ".long/path/to/source", "/long/path/to/destination")
	check("/.long/path/to/destination", ".long/path/to/source", "/.long/path/to/destination")
	check("./long/path/to/destination", "long/path/to/source", "./long/path/to/destination")
	check("./.long/path/to/destination", "long/path/to/source", "./.long/path/to/destination")
	check("./long/path/to/destination", ".long/path/to/source", "./long/path/to/destination")
	check("./.long/path/to/destination", ".long/path/to/source", "./.long/path/to/destination")
}

func Test__ToRelative(t *testing.T) {
	check := func(src, expected string) {
		result := ToRelative(src)
		assert.Equal(t, expected, result, src)
	}

	check("", "")
	check("./../source", "source")
	check("./../.source", ".source")
	check("./../source/..", "")
	check("./../source/../longer", "longer")
	check("./../source/../longer/", "longer")
	check("./../source/../.longer/", ".longer")
	check("./../source/../longer/.", "longer")
	check("./../source/../.longer/.", ".longer")
	check("./../.source/../longer/.", "longer")
	check("./../.source/../.longer/.", ".longer")
	check("source", "source")
	check(".source", ".source")
	check("/source", "source")
	check("./source", "source")
	check("/.source", ".source")
	check("long/path/to/source", "long/path/to/source")
	check(".long/path/to/source", ".long/path/to/source")
	check("/long/path/to/source", "long/path/to/source")
	check("/.long/path/to/source", ".long/path/to/source")
	check("./.long/path/to/source", ".long/path/to/source")
}
