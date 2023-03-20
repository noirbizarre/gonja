{% for path in "fileset/**/*txt" | fileset %}
folder: {{ path | dir | basename }}
file: {{ path | basename }}
{%- endfor %}