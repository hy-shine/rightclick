import { copyFile, mkdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const directory = fileURLToPath(new URL('../internal/catalog/icons/', import.meta.url));
const source = fileURLToPath(new URL('../designs/new-create-v1/assets/', import.meta.url));
await mkdir(directory, { recursive: true });
const files = { txt: 'text', md: 'markdown', json: 'json', csv: 'excel', html: 'html', css: 'css', js: 'js_official', swift: 'swift', py: 'python', docx: 'word', xlsx: 'excel', pptx: 'powerpoint', rtf: 'text' };
for (const [extension, name] of Object.entries(files)) {
  await copyFile(`${source}file-${name}.svg`, `${directory}${extension}.svg`);
}
for (const name of ['vscode-icons-LICENSE.txt', 'lucide-LICENSE.txt', 'markdown-LICENSE.txt']) {
  await copyFile(`${source}${name}`, `${directory}${name}`);
}
