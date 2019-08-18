{{ "This is a long test which will be cutted after some chars."|truncatechars_html(25) }}
{{ "<div class=\"foo\"><ul class=\"foo\"><li class=\"foo\"><p class=\"foo\">This is a long test which will be cutted after some chars.</p></li></ul></div>"|truncatechars_html(25) }}
{{ "<p class='test' id='foo'>This is a long test which will be cutted after some chars.</p>"|truncatechars_html(25) }}
{{ "<a name='link'><p>This </a>is a long test which will be cutted after some chars.</p>"|truncatechars_html(25) }}
{{ "<p>This </a>is a long test which will be cutted after some chars.</p>"|truncatechars_html(25) }}
{{ "<p>This is a long test which will be cutted after some chars.</p>"|truncatechars_html(7) }}
