package lazy

type consumeFunc[IN any] func(v IN) error

func Consume[IN any](stream reader[IN], consumer consumeFunc[IN]) error {
	for v := range stream.ch {
		if err := consumer(v); err != nil {
			stream.Close(err)
			return err
		}
	}
	return nil
}
