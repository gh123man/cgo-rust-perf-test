package main

import "testing"

// Run `cargo build` first!

func BenchmarkRustRegex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < 10000; i++ {
			processStringRs("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}

func BenchmarkGoRegex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < 10000; i++ {
			processStringGo("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}

func BenchmarkRustPassthrough(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < 10000; i++ {
			simpleStringRs("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}

func BenchmarkGoPassthrough(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < 10000; i++ {
			simpleStringGo("Oct 17 14:33:33 | XSS | ERROR | (/viral/interactive/deliverables/holistic.go:3) | sed et dolorem minima et corrupti abcd veniam qui blanditiis optio explicabo et amet qui sint ut iure neque eveniet quod odio distinctio quas veniam voluptatibus quibusdam esse maiores dolores magni numquam sed deserunt quia odio fuga deserunt cumque a aliquam ad dolores dolore aut sapiente necessitatibus ut autem necessitatibus quam eveniet et omnis aut quos dolorem culpa nostrum quas provident tempora voluptate iure quos iste consequatur minima accusantium molestiae consequatur perspiciatis quis quia at incidunt non veritatis deserunt totam iure autem asperiores rerum officiis iusto et explicabo sunt et rerum molestiae hic dolore neque eum vel rerum perspiciatis autem et consequuntur consequatur aliquam dolore magni ea est illum accusamus rerum magnam neque odio voluptatibus est temporibus quo ullam nobis soluta quo ipsum temporibus perferendis et esse repellendus ea id explicabo nostrum repellat vero perferendis possimus optio consectetur deserunt aspern")
		}
	}
}
