package gogroup

import (
	"fmt"
	"regexp"
	"strings"
)

type grouper struct {
	regexps map[int]string

	// The group numbers of prefixed packages.
	prefixes map[int]string

	// The group numbers of standard packages and unidentified packages.
	std, other, named int

	// The next integer to assign
	next int
}

func NewGrouper() *grouper {
	return &grouper{
		regexps:  make(map[int]string),
		prefixes: make(map[int]string),
		std:      0,
		other:    1,
		named:    2,
		next:     3,
	}
}

func (g *grouper) Group(pkg string) int {
	for n, regexpStr := range g.regexps {
		if regexp.MustCompile(regexpStr).FindStringSubmatch(pkg) != nil {
			return n
		}
	}
	for n, prefix := range g.prefixes {
		if strings.HasPrefix(pkg, prefix) {
			return n
		}
	}

	// A dot distinguishes non-standard packages.
	if strings.Contains(pkg, " ") {
		return g.named
	} else if strings.Contains(pkg, ".") {
		return g.other
	} else {
		return g.std
	}
}

func (g *grouper) wasSet() bool {
	return g.next > 3
}

func (g *grouper) String() string {
	parts := []string{}
	remain := len(g.prefixes) + len(g.regexps)
	for i := 0; i <= g.std || i <= g.named || remain > 0; i++ {
		if g.std == i {
			parts = append(parts, "std")
		} else if g.named == i {
			parts = append(parts, "named")
		} else if g.other == i {
			parts = append(parts, "other")
		} else if p, ok := g.prefixes[i]; ok {
			parts = append(parts, fmt.Sprintf("prefix=%s", p))
			remain--
		} else if p, ok := g.regexps[i]; ok {
			parts = append(parts, fmt.Sprintf("regexp=%s", p))
			remain--
		}
	}
	return strings.Join(parts, ",")
}

var rePrefix = regexp.MustCompile(`^prefix=(.*)$`)
var reRegexp = regexp.MustCompile(`^regexp=(.*)$`)

func (g *grouper) Set(s string) error {
	parts := strings.Split(s, ",")
	for _, p := range parts {
		if p == "std" {
			g.std = g.next
		} else if p == "named" {
			g.named = g.next
		} else if p == "other" {
			g.other = g.next
		} else if match := rePrefix.FindStringSubmatch(p); match != nil {
			g.prefixes[g.next] = match[1]
		} else if match := reRegexp.FindStringSubmatch(p); match != nil {
			g.regexps[g.next] = match[1]
		} else {
			return fmt.Errorf("Unknown order specification '%s'", p)
		}
		g.next++
	}
	return nil
}
