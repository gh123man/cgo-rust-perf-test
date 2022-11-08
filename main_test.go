package main

import "testing"

// Run `./build.sh` first!

func benchmarkRustRegex(j int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < j; i++ {
			processStringRs("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}

func BenchmarkRustRegex1(b *testing.B)      { benchmarkRustRegex(1, b) }
func BenchmarkRustRegex10(b *testing.B)     { benchmarkRustRegex(10, b) }
func BenchmarkRustRegex100(b *testing.B)    { benchmarkRustRegex(100, b) }
func BenchmarkRustRegex1000(b *testing.B)   { benchmarkRustRegex(1000, b) }
func BenchmarkRustRegex10000(b *testing.B)  { benchmarkRustRegex(10000, b) }
func BenchmarkRustRegex100000(b *testing.B) { benchmarkRustRegex(100000, b) }

func benchmarkGoRegex(j int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < j; i++ {
			processStringGo("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}

func BenchmarkGoRegex1(b *testing.B)      { benchmarkGoRegex(1, b) }
func BenchmarkGoRegex10(b *testing.B)     { benchmarkGoRegex(10, b) }
func BenchmarkGoRegex100(b *testing.B)    { benchmarkGoRegex(100, b) }
func BenchmarkGoRegex1000(b *testing.B)   { benchmarkGoRegex(1000, b) }
func BenchmarkGoRegex10000(b *testing.B)  { benchmarkGoRegex(10000, b) }
func BenchmarkGoRegex100000(b *testing.B) { benchmarkGoRegex(100000, b) }

func benchmarkRustPassthrough(j int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < j; i++ {
			simpleStringRs("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}

func BenchmarkRustPassthrough1(b *testing.B)      { benchmarkRustPassthrough(1, b) }
func BenchmarkRustPassthrough10(b *testing.B)     { benchmarkRustPassthrough(10, b) }
func BenchmarkRustPassthrough100(b *testing.B)    { benchmarkRustPassthrough(100, b) }
func BenchmarkRustPassthrough1000(b *testing.B)   { benchmarkRustPassthrough(1000, b) }
func BenchmarkRustPassthrough10000(b *testing.B)  { benchmarkRustPassthrough(10000, b) }
func BenchmarkRustPassthrough100000(b *testing.B) { benchmarkRustPassthrough(100000, b) }

func benchmarkGoPassthrough(j int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < j; i++ {
			simpleStringGo("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}
func BenchmarkGoPassthrough1(b *testing.B)      { benchmarkGoPassthrough(1, b) }
func BenchmarkGoPassthrough10(b *testing.B)     { benchmarkGoPassthrough(10, b) }
func BenchmarkGoPassthrough100(b *testing.B)    { benchmarkGoPassthrough(100, b) }
func BenchmarkGoPassthrough1000(b *testing.B)   { benchmarkGoPassthrough(1000, b) }
func BenchmarkGoPassthrough10000(b *testing.B)  { benchmarkGoPassthrough(10000, b) }
func BenchmarkGoPassthrough100000(b *testing.B) { benchmarkGoPassthrough(100000, b) }
