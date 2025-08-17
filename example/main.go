package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gosuda/lazy"
)

func main() {
	nums, w := lazy.New[int](
		lazy.WithContext(context.Background()),
	)
	go func() {
		cnt := 0
		for {
			cnt++
			if err := w.Emit(cnt); err != nil {
				fmt.Println("emit error:", err)
				break
			}
		}
	}()

	evenNums := lazy.Filter(nums, func(ctx *lazy.Context, v int) (bool, error) {
		return v%2 == 0, nil
	})

	formattedNums := lazy.Map(evenNums, func(ctx *lazy.Context, v int) (string, error) {
		time.Sleep(1 * time.Second)
		return fmt.Sprintf("%03d", v), nil
	}, lazy.WithSize(1), lazy.WithParallelism(5))

	summedEverySec := lazy.MapMany(formattedNums, func(ctx *lazy.Context, ch <-chan string, yield func(int) error) error {
		sum := 0
		last := time.Now()
		for v := range ch {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			sum += parsed
			if time.Second < time.Since(last) {
				if err := yield(sum); err != nil {
					return err
				}
				last = time.Now()
				sum = 0
			}
		}
		return nil
	})

	if err := lazy.Consume(summedEverySec, func(ctx *lazy.Context, v int) error {
		fmt.Println(v)
		return nil
	}); err != nil {
		fmt.Println(err)
	}

	time.Sleep(1 * time.Second)
}
