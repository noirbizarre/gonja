// A jinja like template-engine
//
// Blog posts about gonja (including introduction and migration):
// https://www.florian-schlachter.de/?tag=gonja
//
// Complete documentation on the template language:
// https://docs.djangoproject.com/en/dev/topics/templates/
//
// Try out gonja live in the gonja playground:
// https://www.florian-schlachter.de/gonja/
//
// Make sure to read README.md in the repository as well.
//
// A tiny example with template strings:
//
// (Snippet on playground: https://www.florian-schlachter.de/gonja/?id=1206546277)
//
//     // Compile the template first (i. e. creating the AST)
//     tpl, err := gonja.FromString("Hello {{ name|capfirst }}!")
//     if err != nil {
//         panic(err)
//     }
//     // Now you can render the template with the given
//     // gonja.Context how often you want to.
//     out, err := tpl.Execute(gonja.Context{"name": "fred"})
//     if err != nil {
//         panic(err)
//     }
//     fmt.Println(out) // Output: Hello Fred!
//
package gonja
