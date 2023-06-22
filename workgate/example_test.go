// nolint: goconst // using consts in examples makes them harder to read
package workgate

import "fmt"

func ExampleWorkGate_Do() {
	wg := New(10)
	returnValue, err := wg.Do(func() (interface{}, error) {
		// Do some work (blocking)
		return "foo", nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(returnValue)
	// Output: foo
}

func ExampleWorkGate_DoAsync() {
	wg := New(10)
	done := make(chan struct{})
	wg.DoAsync(
		func() {
			// Do some work (async)
			fmt.Println("foo")
			close(done)
		},
		func(err error) {
			fmt.Println(err)
			close(done)
		},
	)
	<-done
	// Output: foo
}

func ExampleWorkGate_TryDo() {
	wg := New(10)

	// Make sure the gate is full
	free := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.DoAsync(
			func() {
				<-free
			},
			nil,
		)
	}
	defer close(free)

	_, err := wg.TryDo(func() (interface{}, error) {
		// Do some work (blocking)
		return "foo", nil
	})
	fmt.Println(err)
	// Output: gate full
}
