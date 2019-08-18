# Gonja

[![GoDoc](https://godoc.org/github.com/noirbizarre/gonja?status.svg)](https://godoc.org/github.com/noirbizarre/gonja)
[![Build Status](https://travis-ci.org/noirbizarre/gonja.svg?branch=master)](https://travis-ci.org/noirbizarre/gonja)
[![Coverage Status](https://codecov.io/gh/noirbizarre/gonja/branch/master/graph/badge.svg)](https://codecov.io/gh/noirbizarre/gonja)

`gonja` is [`pongo2`](https://github.com/flosch/pongo2) fork intended to be aligned on `Jinja` template syntax instead of the `Django` one.

Install/update using `go get` (no dependencies required by `gonja`):
```
go get github.com/noirbizarre/gonja
```

Please use the [issue tracker](https://github.com/noirbizarre/gonja/issues) if you're encountering any problems with gonja or if you need help with implementing tags or filters ([create a ticket!](https://github.com/noirbizarre/gonja/issues/new)).

## First impression of a template

```HTML+Django
<html><head><title>Our admins and users</title></head>
{# This is a short example to give you a quick overview of gonja's syntax. #}

{% macro user_details(user, is_admin=false) %}
	<div class="user_item">
		<!-- Let's indicate a user's good karma -->
		<h2 {% if (user.karma >= 40) || (user.karma > calc_avg_karma(userlist)+5) %}
			class="karma-good"{% endif %}>
			
			<!-- This will call user.String() automatically if available: -->
			{{ user }}
		</h2>

		<!-- Will print a human-readable time duration like "3 weeks ago" -->
		<p>This user registered {{ user.register_date|naturaltime }}.</p>
		
		<!-- Let's allow the users to write down their biography using markdown;
		     we will only show the first 15 words as a preview -->
		<p>The user's biography:</p>
		<p>{{ user.biography|markdown|truncatewords_html:15 }}
			<a href="/user/{{ user.id }}/">read more</a></p>
		
		{% if is_admin %}<p>This user is an admin!</p>{% endif %}
	</div>
{% endmacro %}

<body>
	<!-- Make use of the macro defined above to avoid repetitive HTML code
	     since we want to use the same code for admins AND members -->
	
	<h1>Our admins</h1>
	{% for admin in adminlist %}
		{{ user_details(admin, true) }}
	{% endfor %}
	
	<h1>Our members</h1>
	{% for user in userlist %}
		{{ user_details(user) }}
	{% endfor %}
</body>
</html>
```

## Features (and new in gonja)

 * Entirely rewritten from the ground-up.
 * [Advanced C-like expressions](https://github.com/noirbizarre/gonja/blob/master/template_tests/expressions.tpl).
 * [Complex function calls within expressions](https://github.com/noirbizarre/gonja/blob/master/template_tests/function_calls_wrapper.tpl).
 * [Easy API to create new filters and tags](http://godoc.org/github.com/noirbizarre/gonja#RegisterFilter) ([including parsing arguments](http://godoc.org/github.com/noirbizarre/gonja#Parser))
 * Additional features:
    * Macros including importing macros from other files (see [template_tests/macro.tpl](https://github.com/noirbizarre/gonja/blob/master/template_tests/macro.tpl))
    * [Template sandboxing](https://godoc.org/github.com/noirbizarre/gonja#TemplateSet) ([directory patterns](http://golang.org/pkg/path/filepath/#Match), banned tags/filters)


## How you can help

 * Write [filters](https://github.com/noirbizarre/gonja/blob/master/builtins/filters.go#L3) / [statements](https://github.com/noirbizarre/gonja/blob/master/builtins/statements.go#L4)
 * Write/improve code tests (use the following command to see what tests are missing: `go test -v -cover -covermode=count -coverprofile=cover.out && go tool cover -html=cover.out` or have a look on [gocover.io/github.com/noirbizarre/gonja](http://gocover.io/github.com/noirbizarre/gonja))
 * Write/improve template tests (see the `testData/` directory)
 * Write middleware, libraries and websites using gonja. :-)

# Documentation

For a documentation on how the templating language works you can [head over to the Jinja documentation](https://jinja.palletsprojects.com). gonja aims to be compatible with it.

You can access gonja's API documentation on [godoc](https://godoc.org/github.com/noirbizarre/gonja).

## Caveats 

### Filters

 * **format**: `format` does **not** take Python's string format syntax as a parameter, instead it takes Go's. Essentially `{{ 3.14|stringformat:"pi is %.2f" }}` is `fmt.Sprintf("pi is %.2f", 3.14)`.
 * **escape** / **force_escape**: Unlike Jinja's behaviour, the `escape`-filter is applied immediately. Therefore there is no need for a `force_escape`-filter yet.

# API-usage examples

Please see the documentation for a full list of provided API methods.

## A tiny example (template string)

```Go
// Compile the template first (i. e. creating the AST)
tpl, err := gonja.FromString("Hello {{ name|capfirst }}!")
if err != nil {
	panic(err)
}
// Now you can render the template with the given 
// gonja.Context how often you want to.
out, err := tpl.Execute(gonja.Context{"name": "axel"})
if err != nil {
	panic(err)
}
fmt.Println(out) // Output: Hello Axel!
```

## Example server-usage (template file)

```Go
package main

import (
	"github.com/noirbizarre/gonja"
	"net/http"
)

// Pre-compiling the templates at application startup using the
// little Must()-helper function (Must() will panic if FromFile()
// or FromString() will return with an error - that's it).
// It's faster to pre-compile it anywhere at startup and only
// execute the template later.
var tpl = gonja.Must(gonja.FromFile("example.html"))

func examplePage(w http.ResponseWriter, r *http.Request) {
	// Execute the template per HTTP request
	out, err := tpl.Execute(gonja.Context{"query": r.FormValue("query")})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteString(out)
}

func main() {
	http.HandleFunc("/", examplePage)
	http.ListenAndServe(":8080", nil)
}
```

# Benchmark

The benchmarks have been run on the my machine (`Intel(R) Core(TM) i7-2600 CPU @ 3.40GHz`) using the command:

    go test -bench . -cpu 1,2,4,8

All benchmarks are compiling (depends on the benchmark) and executing the `testData/complex.tpl` template.

The results are:

	BenchmarkFromCache             	   30000	     41259 ns/op
	BenchmarkFromCache-2           	   30000	     42776 ns/op
	BenchmarkFromCache-4           	   30000	     44432 ns/op
	BenchmarkFromFile              	    3000	    437755 ns/op
	BenchmarkFromFile-2            	    3000	    472828 ns/op
	BenchmarkFromFile-4            	    2000	    519758 ns/op
	BenchmarkExecute               	   30000	     41984 ns/op
	BenchmarkExecute-2             	   30000	     48546 ns/op
	BenchmarkExecute-4             	   20000	    104469 ns/op
	BenchmarkCompileAndExecute     	    3000	    428425 ns/op
	BenchmarkCompileAndExecute-2   	    3000	    459058 ns/op
	BenchmarkCompileAndExecute-4   	    3000	    488519 ns/op
	BenchmarkParallelExecute       	   30000	     45262 ns/op
	BenchmarkParallelExecute-2     	  100000	     23490 ns/op
	BenchmarkParallelExecute-4     	  100000	     24206 ns/op

Benchmarked on August 18th 2019.
