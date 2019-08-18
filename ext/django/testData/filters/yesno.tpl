{{ simple.bool_true|yesno }}
{{ simple.bool_false|yesno }}
{{ simple.nil|yesno }}
{{ simple.nothing|yesno }}
{{ simple.bool_true|yesno("ja,nein,vielleicht") }}
{{ simple.bool_false|yesno("ja,nein,vielleicht") }}
{{ simple.nothing|yesno("ja,nein") }}
