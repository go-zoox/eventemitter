# EventEmitter - a simple event emitter

[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-zoox/eventemitter)](https://pkg.go.dev/github.com/go-zoox/eventemitter)
[![Build Status](https://github.com/go-zoox/eventemitter/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-zoox/eventemitter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-zoox/eventemitter)](https://goreportcard.com/report/github.com/go-zoox/eventemitter)
[![Coverage Status](https://coveralls.io/repos/github/go-zoox/eventemitter/badge.svg?branch=master)](https://coveralls.io/github/go-zoox/eventemitter?branch=master)
[![GitHub issues](https://img.shields.io/github/issues/go-zoox/eventemitter.svg)](https://github.com/go-zoox/eventemitter/issues)
[![Release](https://img.shields.io/github/tag/go-zoox/eventemitter.svg?label=Release)](https://github.com/go-zoox/eventemitter/tags)

## Installation
To install the package, run:
```bash
go get github.com/go-zoox/eventemitter
```

## Getting Started

```go
import (
  "testing"
  "github.com/go-zoox/eventemitter"
)

func main(t *testing.T) {
	e := eventemitter.New()
	count := 0
	e.On("test", eventemitter.HandleFunc(func(payload any) {
		count++
		t.Log("test", payload)
	}))

	e.Start()

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		index := i
		wg.Add(1)
		go func() {
			e.Emit("test", index)
			wg.Done()
		}()
	}

	wg.Wait()
}
```

## License
GoZoox is released under the [MIT License](./LICENSE).
