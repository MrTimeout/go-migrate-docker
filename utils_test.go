package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTmpFile(t *testing.T, chmod os.FileMode, content ...string) *os.File {
	f, err := os.CreateTemp(".", "regexFile")
	if err != nil {
		t.Fatal(err)
	}

	for _, line := range content {
		if _, err = f.WriteString(line); err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(line, "\n") {
			if _, err = f.WriteString("\n"); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err = f.Chmod(chmod); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(f.Name())
	})

	return f
}

func TestGroupRegexFirstSubMatchFromFile(t *testing.T) {

	t.Run("regex matching all lines inside the file", func(t *testing.T) {
		input := []string{"first", "second", "third"}
		r := regexp.MustCompile(`([a-z]+)`)
		f := createTmpFile(t, 0666, input...)
		want := input

		got, err := groupRegexFirstSubMatchFromFile(f.Name(), r)
		if err != nil {
			t.Fatal(err)
		}

		assert.ElementsMatch(t, want, got)
	})

	t.Run("regex matching some lines inside the file", func(t *testing.T) {
		input := []string{"123", "123", "third"}
		r := regexp.MustCompile(`([0-9]+)`)
		f := createTmpFile(t, 0666, input...)
		want := input[:2]

		got, err := groupRegexFirstSubMatchFromFile(f.Name(), r)
		if err != nil {
			t.Fatal(err)
		}

		assert.ElementsMatch(t, want, got)
	})

	t.Run("regex matching none lines inside the file", func(t *testing.T) {
		input := []string{"first", "second", "third"}
		r := regexp.MustCompile(`([0-9]+)`)
		f := createTmpFile(t, 0666, input...)

		got, err := groupRegexFirstSubMatchFromFile(f.Name(), r)
		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, got)
	})

	t.Run("regex can't be accomplished because file has not the right permissions", func(t *testing.T) {
		r := regexp.MustCompile(`[0-9]+`)
		_, err := groupRegexFirstSubMatchFromFile("notexistentfile", r)

		assert.ErrorContains(t, err, "no such file or directory")
	})
}

func TestMatchesOfStringArr(t *testing.T) {
	t.Run("matches all of the elements of input string arr", func(t *testing.T) {
		input := []string{"123", "789", "456"}
		r := regexp.MustCompile(`[0-9]+`)
		want := input

		got := matchesOfStringArr(r, input)

		assert.ElementsMatch(t, want, got)
	})

	t.Run("matches nonw of the elements of input string arr", func(t *testing.T) {
		input := []string{"bye", "hello"}
		r := regexp.MustCompile(`[0-9]+[a-z]+`)

		got := matchesOfStringArr(r, input)

		assert.Empty(t, got)
	})
}
