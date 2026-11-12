package files

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/hy-shine/rightclick/locales"
	"hash"
	"io"
	"os"
)

// Algorithm is a hash function offered by the menu.
type Algorithm string

const (
	SHA256 Algorithm = "SHA-256"
	SHA1   Algorithm = "SHA-1"
	MD5    Algorithm = "MD5"
)

// Hash returns the lowercase hex digest of the file at path. Folders and other
// non-regular files are refused.
func Hash(path string, algo Algorithm) (string, error) {
	h, err := newHash(algo)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New(locales.HashNeedsFile(path))
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CompareHashes hashes two files and reports whether the digests are equal,
// along with both digests.
func CompareHashes(a, b string, algo Algorithm) (same bool, hashA, hashB string, err error) {
	if hashA, err = Hash(a, algo); err != nil {
		return false, "", "", err
	}
	if hashB, err = Hash(b, algo); err != nil {
		return false, "", "", err
	}
	return hashA == hashB, hashA, hashB, nil
}

func newHash(algo Algorithm) (hash.Hash, error) {
	switch algo {
	case SHA256:
		return sha256.New(), nil
	case SHA1:
		return sha1.New(), nil
	case MD5:
		return md5.New(), nil
	}
	return nil, errors.New(locales.UnsupportedHash(string(algo)))
}
