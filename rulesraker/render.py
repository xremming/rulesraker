import jinja2

from .parse import Rules


def render(rules: Rules) -> str:
    env = jinja2.Environment(
        loader=jinja2.PackageLoader("rulesraker"),
        autoescape=jinja2.select_autoescape()
    )

    tmpl = env.get_template("main.html.j2")
    return tmpl.render(rules=rules)
