{% filter truncatewords(9) %}{{ lorem(25, "w") }}{% endfilter %}
{% filter wordcount %}{% filter truncatewords(9) %}{{ lorem(25, "w") }}{% endfilter %}{% endfilter %}
{{ simple.chinese_hello_world|truncatewords(0) }}
{{ simple.chinese_hello_world|truncatewords(1) }}
{{ simple.chinese_hello_world|truncatewords(2) }}
