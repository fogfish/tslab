# Typed Slab allocation

A slab-like allocator library in the Go Programming Language. Classical [slab allocator](https://en.wikipedia.org/wiki/Slab_allocation) provides `[]byte`. This library allocates instances of type `T`. 

## Inspiration

The library helps to reduce number of GC and implement allocation pools for generic objects of type `T`. It is developed to address the challenge of data co-allocation within data structures and externalize (swap in/out) the allocated slabs. 

## Getting started

The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

```go
import "github.com/fogfish/tslab"

// data structure uses `tslab.Pointer[T]`, which contains both
// pointer to struct and metadata about allocations
type Node struct {
  left, right tslab.Pointer[Node]
  value       int
}

// Allocate and initialize memory 
func New(heap tslab.Allocator[Node], value int) NodeID { 
  addr := heap.Alloc()
  node := addr.ValueOf
	node.Value = value
  return addr
}

// heap implements an allocator for the given type
var heap *tslab.Heap[Node] = tslab.New[Node](4 * 1024)

// Use the node
var node NodeID = New(heap, 10)
```


## How To Contribute

The library is [MIT](LICENSE) licensed and accepts contributions via GitHub pull requests:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

The build and testing process requires [Go](https://golang.org) version 1.13 or later.

**build** and **test** library.

```bash
git clone https://github.com/fogfish/tslab
cd tslab
go test
```

### commit message

The commit message helps us to write a good release note, speed-up review process. The message should address two question what changed and why. The project follows the template defined by chapter [Contributing to a Project](http://git-scm.com/book/ch5-2.html) of Git book.

### bugs

If you experience any issues with the library, please let us know via [GitHub issues](https://github.com/fogfish/tslab/issue). We appreciate detailed and accurate reports that help us to identity and replicate the issue. 


## License

[![See LICENSE](https://img.shields.io/github/license/fogfish/tslab.svg?style=for-the-badge)](LICENSE)