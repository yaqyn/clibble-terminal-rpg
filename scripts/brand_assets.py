#!/usr/bin/env python3
"""Generate vector marks and palette adapters from the shared brand identity."""
import json
from pathlib import Path

root = Path(__file__).resolve().parent.parent
identity = json.loads((root / 'internal/brand/identity.json').read_text())
colors = identity['colors']
mark = f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><rect width="48" height="48" rx="12" fill="{colors['cobalt']}"/><path d="M14 13h9c12 0 16 22 0 22h-9V22" fill="none" stroke="{colors['cloud']}" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/><circle cx="14" cy="15" r="3" fill="{colors['azure']}"/></svg>\n'''
(root / 'readme-assets/mark.svg').write_text(mark)
(root / 'internal/webui/assets/mark.svg').write_text(mark)
css = ':root{' + ''.join(f'--{key}:{value};' for key, value in colors.items()) + '}\n'
(root / 'internal/webui/assets/tokens.css').write_text(css)
# Recolor the established canonical icon geometry for the explicitly requested rebrand.
for icon in (root / 'readme-assets/icons').glob('*.svg'):
    import re
    content = icon.read_text()
    content = re.sub(r'fill="#[0-9a-fA-F]{6}"', f'fill="{colors["ink"]}"', content)
    content = re.sub(r'stroke="#[0-9a-fA-F]{6}"', f'stroke="{colors["azure"]}"', content)
    icon.write_text(content)
