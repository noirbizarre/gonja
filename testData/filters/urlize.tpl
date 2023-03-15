{{ "http://www.florian-schlachter.de"|urlize|safe }}
{{ "http://www.florian-schlachter.de"|urlize(rel='nofollow')|safe }}
{{ "http://www.florian-schlachter.de"|urlize(rel='nofollow', target='_blank')|safe }}
{{ "http://www.florian-schlachter.de"|urlize(rel='noopener')|safe }}
{{ "www.florian-schlachter.de"|urlize|safe }}
{{ "florian-schlachter.de"|urlize|safe }}
--
{% filter urlize|safe %}
Please mail me at demo@example.com or visit mit on:
- lorem ipsum github.com/nikolalohinski/gonja lorem ipsum
- lorem ipsum http://www.florian-schlachter.de lorem ipsum
- lorem ipsum https://www.florian-schlachter.de lorem ipsum
- lorem ipsum https://www.florian-schlachter.de lorem ipsum
- lorem ipsum www.florian-schlachter.de lorem ipsum
- lorem ipsum www.florian-schlachter.de/test="test" lorem ipsum
{% endfilter %}
--
{% filter urlize(target='_blank', rel="nofollow")|safe %}
Please mail me at demo@example.com or visit mit on:
- lorem ipsum github.com/nikolalohinski/gonja lorem ipsum
- lorem ipsum http://www.florian-schlachter.de lorem ipsum
- lorem ipsum https://www.florian-schlachter.de lorem ipsum
- lorem ipsum https://www.florian-schlachter.de lorem ipsum
- lorem ipsum www.florian-schlachter.de lorem ipsum
- lorem ipsum www.florian-schlachter.de/test="test" lorem ipsum
{% endfilter %}
--
{% filter urlize(15)|safe %}
Please mail me at demo@example.com or visit mit on:
- lorem ipsum github.com/nikolalohinski/gonja lorem ipsum
- lorem ipsum http://www.florian-schlachter.de lorem ipsum
- lorem ipsum https://www.florian-schlachter.de lorem ipsum
- lorem ipsum https://www.florian-schlachter.de lorem ipsum
- lorem ipsum www.florian-schlachter.de lorem ipsum
- lorem ipsum www.florian-schlachter.de/test="test" lorem ipsum
{% endfilter %}
