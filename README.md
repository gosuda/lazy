# lazy

The type-safe Go library for building data pipelines for clean and maintainable code. Using channels and generics, you can build pipelines that are **cancelable**, **backpressure-aware**, and **optionally parallel** — without wiring channels and goroutines every time.

## Key Features

- **Composable pipelines:** Write clear, dataflow-style code with familiar functions like `Filter`, `Map`, `Reduce`.
- **Backpressure-aware:** Uses buffered channels; `WithSize` controls how much work to prepare ahead.
- **Cancelable:** Pass a `context.Context` to stop pipelines promptly.
- **Error propagation:** Propagate errors upstream and stop the whole pipeline if needed.
- **Tiny, zero-dependency:** Just Go generics and channels.

## Install

```bash
go get github.com/gosuda/lazy@latest
```

## Quick Start

### 1. Create a new lazy stream

```go
nums, w := lazy.New[int](
  lazy.WithContext(context.Background()), // optional
)

go func() {
  for i := 1; i <= 1000; i++ {
    if err := w.Emit(i); err != nil {
      // You may receive an error propagated from another component
      // in the same pipeline, or when context is done.
    }
  }
}()
```

### 2. Filter even numbers

```go
evens := lazy.Filter(nums, func(v int) (bool, error) {
  return v%2 == 0, nil
})
```

### 3. Map to zero-padded strings

```go
padded := lazy.Map(evens, func(v int) (string, error) {
  return fmt.Sprintf("%03d", v), nil
})
```

### 4. Consume the results

```go
lazy.Consume(padded, func(v string) error {
  // 002, 004, 006, ...
  fmt.Println(v)
  return nil
})
```

Check out the full [example](./example/main.go) for a complete end-to-end pipeline.

## References

### [New[T]](./new.go)

```go
func New[T any](opts ...optionFunc) (reader[T], writer[T])
```

New creates a new lazy stream. It returns a reader and a writer. You can `Emit` values into the writer and the emitted values will be available in the reader.

Available options:

| Option                         | Description                                                                                | Default                |
| ------------------------------ | ------------------------------------------------------------------------------------------ | ---------------------- |
| `WithSize(int)`                | The buffer size of the stream. Prepares some items in advance.                             | `0`                    |
| `WithContext(context.Context)` | Set the context for cancellation. When the context is done, the stream closes immediately. | `context.Background()` |

Example:

```go
nums, w := lazy.New[int](lazy.WithSize(2), lazy.WithContext(context.Background()))
go func() {
  defer w.Close()
  for i := 1; i <= 3; i++ {
    _ = w.Emit(i)
  }
}()
```

### [NewSlice[T]](./new.go)

```go
func NewSlice[T any](slice []T, opts ...optionFunc) reader[T]
```

Creates a stream from a slice. The stream closes when the slice is exhausted or the context is done.

Available options:

| Option                         | Description                                                                                | Default                |
| ------------------------------ | ------------------------------------------------------------------------------------------ | ---------------------- |
| `WithSize(int)`                | The buffer size of the stream. Prepares some items in advance.                             | `0`                    |
| `WithContext(context.Context)` | Set the context for cancellation. When the context is done, the stream closes immediately. | `context.Background()` |

Example:

```go
r := lazy.NewSlice([]int{1, 2, 3, 4}, lazy.WithSize(1))
```

### [Filter](./filter.go)

```go
func Filter[T any](stream reader[T], filter func(v T) (bool, error), opts ...optionFunc) reader[T]
```

Keeps values for which the predicate returns `true`.

- Uses `WithParallelism(n)` to process items with `n` worker goroutines.
- Uses `WithSize(k)` to set the output buffer size.
- On predicate error, behavior is controlled by `WithErrHandler` (default: ignore and continue). Returning `DecisionStop` stops the whole pipeline.

Available options:

| Option                           | Description                                                                            | Default                |
| -------------------------------- | -------------------------------------------------------------------------------------- | ---------------------- |
| `WithSize(int)`                  | Buffer size for the output stream.                                                     | `0`                    |
| `WithContext(context.Context)`   | Context for cancellation shared with the operator.                                     | `context.Background()` |
| `WithErrHandler(errHandlerFunc)` | How to handle predicate errors (`DecisionIgnore` to continue, `DecisionStop` to stop). | `DecisionIgnore`       |
| `WithParallelism(int)`           | Number of worker goroutines to run the predicate.                                      | `1`                    |

Example:

```go
evens := lazy.Filter(nums, func(v int) (bool, error) {
  return v%2 == 0, nil
}, lazy.WithParallelism(2), lazy.WithSize(1))
```

### [Map](./map.go)

```go
func Map[IN any, OUT any](stream reader[IN], mapper func(IN) (OUT, error), opts ...optionFunc) reader[OUT]
```

Transforms each value with `mapper`.

- Uses `WithParallelism(n)` to process items with `n` worker goroutines.
- Uses `WithSize(k)` to set the output buffer size.
- On `mapper` error, behavior is controlled by `WithErrHandler` (default: ignore and continue). Returning `DecisionStop` stops the whole pipeline.

Available options:

| Option                           | Description                                                                         | Default                |
| -------------------------------- | ----------------------------------------------------------------------------------- | ---------------------- |
| `WithSize(int)`                  | Buffer size for the output stream.                                                  | `0`                    |
| `WithContext(context.Context)`   | Context for cancellation shared with the operator.                                  | `context.Background()` |
| `WithErrHandler(errHandlerFunc)` | How to handle mapper errors (`DecisionIgnore` to continue, `DecisionStop` to stop). | `DecisionIgnore`       |
| `WithParallelism(int)`           | Number of worker goroutines to run `mapper`.                                        | `1`                    |

Example:

```go
formatted := lazy.Map(evens, func(v int) (string, error) {
  return fmt.Sprintf("%03d", v), nil
}, lazy.WithParallelism(4), lazy.WithSize(2))
```

### [MapMany](./map_many.go)

```go
func MapMany[IN, OUT any](stream reader[IN], f func(ctx context.Context, v <-chan IN, yield func(OUT) error) error, opts ...optionFunc) reader[OUT]
```

Generalized mapping for fan-out/fan-in, batching, windowing, etc. You read from `v` and `yield` zero or more outputs. Any returned error closes upstream. Honors context from options. Runs in a single goroutine.

Available options:

| Option                         | Description                             | Default                |
| ------------------------------ | --------------------------------------- | ---------------------- |
| `WithSize(int)`                | Buffer size for the output stream.      | `0`                    |
| `WithContext(context.Context)` | Context passed to `f` for cancellation. | `context.Background()` |

Example:

```go
batches := lazy.MapMany(nums, func(ctx context.Context, ch <-chan int, yield func([]int) error) error {
  buf := make([]int, 0, 3)
  for v := range ch {
    buf = append(buf, v)
    if len(buf) == 3 {
      if err := yield(buf); err != nil { return err }
      buf = make([]int, 0, 3)
    }
  }
  if len(buf) > 0 {
    if err := yield(buf); err != nil { return err }
  }
  return nil
})
```

### [Consume](./consume.go)

```go
func Consume[IN any](stream reader[IN], consumer func(IN) error) error
```

Drains the stream, calling `consumer` for each value. If `consumer` returns an error, the error is propagated upstream and returned.

Example:

```go
if err := lazy.Consume(strs, func(s string) error {
  fmt.Println(s)
  return nil
}); err != nil {
  fmt.Println("consume error:", err)
}
```

### Options

```go
// WithSize sets the buffer size for newly created output streams.
func WithSize(size int) optionFunc

// WithContext sets the cancellation context shared by operators that accept options.
func WithContext(ctx context.Context) optionFunc

// WithErrHandler controls how `Map` handles mapper errors (ignore or stop).
func WithErrHandler(handler errHandlerFunc) optionFunc
type errHandlerFunc func(error) Decision
const (
  DecisionStop   Decision = "stop"
  DecisionIgnore Decision = "ignore"
)

// WithParallelism sets the number of workers used by `Map`.
func WithParallelism(parallel int) optionFunc
```
