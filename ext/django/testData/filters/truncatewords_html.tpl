{{ "This is a long test which will be cutted after some words."|truncatewords_html(25)|safe }}
{{ "<div class=\"foo\"><ul class=\"foo\"><li class=\"foo\"><p class=\"foo\">This is a long test which will be cutted after some chars.</p></li></ul></div>"|truncatewords_html(5) }}
{{ "<p>This. is. a. long test. Test test, test.</p>"|truncatewords_html(8) }}
{{ "<a name='link' href=\"https://....\"><p class=\"foo\">This </a>is a long test, which will be cutted after some words.</p>"|truncatewords_html(5) }}
{{ "<p>This </a>is a long test, which will be cutted after some words.</p>"|truncatewords_html(5) }}
{{ "<p>This is a long test which will be cutted after some words.</p>"|truncatewords_html(2) }}
{{ "<p>This is a long test which will be cutted after some words.</p>"|truncatewords_html(0) }}
