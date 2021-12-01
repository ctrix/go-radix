go-radix [![Build Status](https://travis-ci.org/armon/go-radix.png)](https://travis-ci.org/armon/go-radix)
=========

Provides the `radix` package that implements a [radix tree](http://en.wikipedia.org/wiki/Radix_tree).
The package only provides a single `Tree` implementation, optimized for sparse nodes.

As a radix tree, it provides the following:
 * O(k) operations. In many cases, this can be faster than a hash table since
   the hash function is an O(k) operation, and hash tables have very poor cache locality.
 * Minimum / Maximum value lookups
 * Ordered iteration

This repository has been forked from [github.com/armon/go-radix](https://github.com/armon/go-radix), most pending pull requests have been committed 
and the code has been modified to accept `[]byte` instead of `string` as keys.

For an immutable variant, see [go-immutable-radix](https://github.com/hashicorp/go-immutable-radix).

Documentation
=============

The full documentation is available on [Godoc](http://godoc.org/github.com/armon/go-radix).

Example
=======

Below is a simple example of usage

```go
	// Create a tree
	r := New()
	r.Insert([]byte("foo"), 1)
	r.Insert([]byte("bar"), 2)
	r.Insert([]byte("foobar"), 2)

	// Find the longest prefix match
	m, _, _ := r.LongestPrefix([]byte("foozip"))
	if bytes.Compare(m, []byte("foo")) != 0 {
		panic("should be foo")
	}
```
