import DOMPurify from 'dompurify';
import { marked } from 'marked';

/** Render markdown to sanitized HTML for safe use with v-html. */
export function renderMarkdown(md: string): string {
    if (!md?.trim()) return '';
    const raw = marked.parse(md, { async: false }) as string;
    return DOMPurify.sanitize(raw, {
        ALLOWED_TAGS: [
            'p', 'br', 'strong', 'em', 'u', 's', 'code', 'pre',
            'ul', 'ol', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
            'blockquote', 'a', 'hr', 'table', 'thead', 'tbody', 'tr', 'th', 'td',
        ],
        ALLOWED_ATTR: ['href', 'target', 'rel'],
    });
}
