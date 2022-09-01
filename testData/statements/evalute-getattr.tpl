{% set a={"b":{"key":"value"}} %}
{% set b="b" %}
{% set c={"d":"key"} %}
{% set d="d" %}

{{ a.b[c.d] }}
