{{ simple.name|make_list|join(", ") }}
{% for char in simple.name|make_list %}{{ char }}{% endfor %}
