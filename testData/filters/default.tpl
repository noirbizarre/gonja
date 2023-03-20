{{ simple.nothing | default("n/a") }}
{{ nothing | default(simple.number) }}
{{ simple.number | default("n/a") }}
{{ 5 | d("n/a") }}