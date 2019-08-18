# [pongo](https://en.wikipedia.org/wiki/Pongo_%28genus%29)2

[![Join the chat at https://gitter.im/noirbizarre/gonja](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/noirbizarre/gonja)
[![GoDoc](https://godoc.org/github.com/noirbizarre/gonja?status.svg)](https://godoc.org/github.com/noirbizarre/gonja)
[![Build Status](https://travis-ci.org/noirbizarre/gonja.svg?branch=master)](https://travis-ci.org/noirbizarre/gonja)
[![Coverage Status](https://coveralls.io/repos/noirbizarre/gonja/badge.svg?branch=master)](https://coveralls.io/r/noirbizarre/gonja?branch=master)
[![gratipay](http://img.shields.io/badge/gratipay-support%20pongo-brightgreen.svg)](https://gratipay.com/flosch/)
[![Bountysource](https://www.bountysource.com/badge/tracker?tracker_id=3654947)](https://www.bountysource.com/trackers/3654947-gonja?utm_source=3654947&utm_medium=shield&utm_campaign=TRACKER_BADGE)

`gonja` is [`gonja`](https://github.com/noirbizarre/gonja) fork intended to be aligned on `Jinja` template syntax instead of the `Django` one.

Install/update using `go get` (no dependencies required by `gonja`):
```
go get github.com/noirbizarre/gonja
```

Please use the [issue tracker](https://github.com/noirbizarre/gonja/issues) if you're encountering any problems with gonja or if you need help with implementing tags or filters ([create a ticket!](https://github.com/noirbizarre/gonja/issues/new)). If possible, please use [playground](https://www.florian-schlachter.de/gonja/) to create a short test case on what's wrong and include the link to the snippet in your issue.

**New**: [Try gonja out in the gonja playground.](https://www.florian-schlachter.de/gonja/)

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

## Development status

**Latest stable release**: v3.0 (`go get -u gopkg.in/noirbizarre/gonja.v3` / [`v3`](https://github.com/noirbizarre/gonja/tree/v3)-branch) [[read the announcement](https://www.florian-schlachter.de/post/gonja-v3/)]

**Current development**: v4 (`master`-branch)

*Note*: With the release of pongo v4 the branch v2 will be deprecated.

**Deprecated versions** (not supported anymore): v1

| Topic                                | Status                                                                                 |
| ------------------------------------ | -------------------------------------------------------------------------------------- |       
| Django version compatibility:        | [1.7](https://docs.djangoproject.com/en/1.7/ref/templates/builtins/)                  |
| *Missing* (planned) **filters**:     | none ([hints](https://github.com/noirbizarre/gonja/blob/master/filters_builtin.go#L3))     | 
| *Missing* (planned) **tags**:        | none ([hints](https://github.com/noirbizarre/gonja/blob/master/tags.go#L3))                |

Please also have a look on the [caveats](https://github.com/noirbizarre/gonja#caveats) and on the [official add-ons](https://github.com/noirbizarre/gonja#official).

## Features (and new in gonja)

 * Entirely rewritten from the ground-up.
 * [Advanced C-like expressions](https://github.com/noirbizarre/gonja/blob/master/template_tests/expressions.tpl).
 * [Complex function calls within expressions](https://github.com/noirbizarre/gonja/blob/master/template_tests/function_calls_wrapper.tpl).
 * [Easy API to create new filters and tags](http://godoc.org/github.com/noirbizarre/gonja#RegisterFilter) ([including parsing arguments](http://godoc.org/github.com/noirbizarre/gonja#Parser))
 * Additional features:
    * Macros including importing macros from other files (see [template_tests/macro.tpl](https://github.com/noirbizarre/gonja/blob/master/template_tests/macro.tpl))
    * [Template sandboxing](https://godoc.org/github.com/noirbizarre/gonja#TemplateSet) ([directory patterns](http://golang.org/pkg/path/filepath/#Match), banned tags/filters)

## Recent API changes within gonja

If you're using the `master`-branch of gonja, you might be interested in this section. Since gonja is still in development (even though there is a first stable release!), there could be (backwards-incompatible) API changes over time. To keep track of these and therefore make it painless for you to adapt your codebase, I'll list them here.

 * Function signature for tag execution changed: not taking a `bytes.Buffer` anymore; instead `Execute()`-functions are now taking a `TemplateWriter` interface.
 * Function signature for tag and filter parsing/execution changed (`error` return type changed to `*Error`).
 * `NodeEvaluator` has been removed and got replaced by `Expression`. You can change your existing tags/filters by simply replacing the interface.
 * Two new helper functions: [`RenderTemplateFile()`](https://godoc.org/github.com/noirbizarre/gonja#RenderTemplateFile) and [`RenderTemplateString()`](https://godoc.org/github.com/noirbizarre/gonja#RenderTemplateString).
 * `Template.ExecuteRW()` is now [`Template.ExecuteWriter()`](https://godoc.org/github.com/noirbizarre/gonja#Template.ExecuteWriter)
 * `Template.Execute*()` functions do now take a `gonja.Context` directly (no pointer anymore).

## How you can help

 * Write [filters](https://github.com/noirbizarre/gonja/blob/master/filters_builtin.go#L3) / [tags](https://github.com/noirbizarre/gonja/blob/master/tags.go#L4) (see [tutorial](https://www.florian-schlachter.de/post/gonja/)) by forking gonja and sending pull requests
 * Write/improve code tests (use the following command to see what tests are missing: `go test -v -cover -covermode=count -coverprofile=cover.out && go tool cover -html=cover.out` or have a look on [gocover.io/github.com/noirbizarre/gonja](http://gocover.io/github.com/noirbizarre/gonja))
 * Write/improve template tests (see the `template_tests/` directory)
 * Write middleware, libraries and websites using gonja. :-)

# Documentation

For a documentation on how the templating language works you can [head over to the Django documentation](https://docs.djangoproject.com/en/dev/topics/templates/). gonja aims to be compatible with it.

You can access gonja's API documentation on [godoc](https://godoc.org/github.com/noirbizarre/gonja).

## Blog post series
 
 * [gonja v3 released](https://www.florian-schlachter.de/post/gonja-v3/)
 * [gonja v2 released](https://www.florian-schlachter.de/post/gonja-v2/)
 * [gonja 1.0 released](https://www.florian-schlachter.de/post/gonja-10/) [August 8th 2014]
 * [gonja playground](https://www.florian-schlachter.de/post/gonja-playground/) [August 1st 2014]
 * [Release of gonja 1.0-rc1 + gonja-addons](https://www.florian-schlachter.de/post/gonja-10-rc1/) [July 30th 2014]
 * [Introduction to gonja + migration- and "how to write tags/filters"-tutorial.](https://www.florian-schlachter.de/post/gonja/) [June 29th 2014]

## Caveats 

### Filters

 * **date** / **time**: The `date` and `time` filter are taking the Golang specific time- and date-format (not Django's one) currently. [Take a look on the format here](http://golang.org/pkg/time/#Time.Format).
 * **stringformat**: `stringformat` does **not** take Python's string format syntax as a parameter, instead it takes Go's. Essentially `{{ 3.14|stringformat:"pi is %.2f" }}` is `fmt.Sprintf("pi is %.2f", 3.14)`.
 * **escape** / **force_escape**: Unlike Django's behaviour, the `escape`-filter is applied immediately. Therefore there is no need for a `force_escape`-filter yet.

### Tags

 * **for**: All the `forloop` fields (like `forloop.counter`) are written with a capital letter at the beginning. For example, the `counter` can be accessed by `forloop.Counter` and the parentloop by `forloop.Parentloop`.
 * **now**: takes Go's time format (see **date** and **time**-filter).

### Misc

 * **not in-operator**: You can check whether a map/struct/string contains a key/field/substring by using the in-operator (or the negation of it):
    `{% if key in map %}Key is in map{% else %}Key not in map{% endif %}` or `{% if !(key in map) %}Key is NOT in map{% else %}Key is in map{% endif %}`.

# Add-ons, libraries and helpers

## Official

 * [ponginae](https://github.com/flosch/ponginae) - A web-framework for Go (using gonja).
 * [gonja-tools](https://github.com/noirbizarre/gonja-tools) - Official tools and helpers for gonja
 * [gonja-addons](https://github.com/noirbizarre/gonja-addons) - Official additional filters/tags for gonja (for example a **markdown**-filter). They are in their own repository because they're relying on 3rd-party-libraries.

## 3rd-party

 * [beego-gonja](https://github.com/oal/beego-gonja) - A tiny little helper for using gonja with [Beego](https://github.com/astaxie/beego).
 * [beego-gonja.v2](https://github.com/ipfans/beego-gonja.v2) - Same as `beego-gonja`, but for gonja v2.
 * [macaron-gonja](https://github.com/macaron-contrib/gonja) - gonja support for [Macaron](https://github.com/Unknwon/macaron), a modular web framework.
 * [gingonja](https://github.com/ngerakines/gingonja) - middleware for [gin](github.com/gin-gonic/gin) to use gonja templates
 * [Build'n support for Iris' template engine](https://github.com/kataras/iris) 
 * [gonjagin](https://gitlab.com/go-box/gonjagin) - alternative renderer for [gin](github.com/gin-gonic/gin) to use gonja templates
 * [gonja-trans](https://github.com/digitalcrab/gonjatrans) - `trans`-tag implementation for internationalization
 * [tgonja](https://github.com/tango-contrib/tgonja) - gonja support for [Tango](https://github.com/lunny/tango), a micro-kernel & pluggable web framework.
 * [p2cli](https://github.com/wrouesnel/p2cli) - command line templating utility based on gonja
 
Please add your project to this list and send me a pull request when you've developed something nice for gonja.

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
out, err := tpl.Execute(gonja.Context{"name": "florian"})
if err != nil {
	panic(err)
}
fmt.Println(out) // Output: Hello Florian!
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
var tplExample = gonja.Must(gonja.FromFile("example.html"))

func examplePage(w http.ResponseWriter, r *http.Request) {
	// Execute the template per HTTP request
	err := tplExample.ExecuteWriter(gonja.Context{"query": r.FormValue("query")}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", examplePage)
	http.ListenAndServe(":8080", nil)
}
```

# Benchmark

The benchmarks have been run on the my machine (`Intel(R) Core(TM) i7-2600 CPU @ 3.40GHz`) using the command:

    go test -bench . -cpu 1,2,4,8

All benchmarks are compiling (depends on the benchmark) and executing the `template_tests/complex.tpl` template.

The results are:

    BenchmarkExecuteComplexWithSandboxActive                50000             60450 ns/op
    BenchmarkExecuteComplexWithSandboxActive-2              50000             56998 ns/op
    BenchmarkExecuteComplexWithSandboxActive-4              50000             60343 ns/op
    BenchmarkExecuteComplexWithSandboxActive-8              50000             64229 ns/op
    BenchmarkCompileAndExecuteComplexWithSandboxActive      10000            164410 ns/op
    BenchmarkCompileAndExecuteComplexWithSandboxActive-2    10000            156682 ns/op
    BenchmarkCompileAndExecuteComplexWithSandboxActive-4    10000            164821 ns/op
    BenchmarkCompileAndExecuteComplexWithSandboxActive-8    10000            171806 ns/op
    BenchmarkParallelExecuteComplexWithSandboxActive        50000             60428 ns/op
    BenchmarkParallelExecuteComplexWithSandboxActive-2      50000             31887 ns/op
    BenchmarkParallelExecuteComplexWithSandboxActive-4     100000             22810 ns/op
    BenchmarkParallelExecuteComplexWithSandboxActive-8     100000             18820 ns/op
    BenchmarkExecuteComplexWithoutSandbox                   50000             56942 ns/op
    BenchmarkExecuteComplexWithoutSandbox-2                 50000             56168 ns/op
    BenchmarkExecuteComplexWithoutSandbox-4                 50000             57838 ns/op
    BenchmarkExecuteComplexWithoutSandbox-8                 50000             60539 ns/op
    BenchmarkCompileAndExecuteComplexWithoutSandbox         10000            162086 ns/op
    BenchmarkCompileAndExecuteComplexWithoutSandbox-2       10000            159771 ns/op
    BenchmarkCompileAndExecuteComplexWithoutSandbox-4       10000            163826 ns/op
    BenchmarkCompileAndExecuteComplexWithoutSandbox-8       10000            169062 ns/op
    BenchmarkParallelExecuteComplexWithoutSandbox           50000             57152 ns/op
    BenchmarkParallelExecuteComplexWithoutSandbox-2         50000             30276 ns/op
    BenchmarkParallelExecuteComplexWithoutSandbox-4        100000             22065 ns/op
    BenchmarkParallelExecuteComplexWithoutSandbox-8        100000             18034 ns/op

Benchmarked on October 2nd 2014.
