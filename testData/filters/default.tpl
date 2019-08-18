{{ simple.nothing|default("n/a") }}
{{ nothing|default(simple.number) }}
{{ simple.number|default("n/a") }}
{{ 5|default("n/a") }}
{{ ''|default('the string was empty') }}
{{ ''|default('the string was empty', true) }}
{{ ''|default('the string was empty', boolean=true) }}
