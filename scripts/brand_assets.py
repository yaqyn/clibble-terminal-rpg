#!/usr/bin/env python3
"""Generate vector marks and palette adapters from the shared brand identity."""
import json
from pathlib import Path

root = Path(__file__).resolve().parent.parent
identity = json.loads((root / 'internal/brand/identity.json').read_text())
colors = identity['colors']
mark = f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><rect width="48" height="48" rx="8" fill="{colors['ink']}"/><path d="m12 15 10 9-10 9" fill="none" stroke="{colors['fern']}" stroke-width="4" stroke-linecap="square" stroke-linejoin="miter"/><path d="M27 30h10v5H27z" fill="{colors['cloud']}"/></svg>\n'''
(root / 'readme-assets/mark.svg').write_text(mark)
(root / 'internal/webui/assets/mark.svg').write_text(mark)
css = ':root{' + ''.join(f'--{key}:{value};' for key, value in colors.items()) + '}\n'
(root / 'internal/webui/assets/tokens.css').write_text(css)
# Recolor the established canonical icon geometry for the explicitly requested rebrand.
for icon in (root / 'readme-assets/icons').glob('*.svg'):
    import re
    content = icon.read_text()
    content = re.sub(r'fill="#[0-9a-fA-F]{6}"', f'fill="{colors["ink"]}"', content)
    content = re.sub(r'stroke="#[0-9a-fA-F]{6}"', f'stroke="{colors["fern"]}"', content)
    icon.write_text(content)
