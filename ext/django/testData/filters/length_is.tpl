{{ simple.name|length_is(8) }}
{{ simple.name|length_is(10) }}
{{ simple.name|length_is("8") }}
{{ simple.name|length_is("10") }}
{{ 5|length_is(1) }}
{{ simple.chinese_hello_world|length_is(4) }}
{{ simple.chinese_hello_world|length_is(3) }}
{{ simple.chinese_hello_world|length_is(5) }}
